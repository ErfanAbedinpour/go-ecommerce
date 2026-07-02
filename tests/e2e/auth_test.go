//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestAuth_Login_Success(t *testing.T) {
	resp := testClient.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    adminEmail,
		"password": adminPassword,
	}, nil)
	resp.AssertStatus(t, http.StatusOK)

	var tokens tokenResponse
	resp.JSON(t, &tokens)
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected access and refresh tokens")
	}
	if tokens.TokenType != "Bearer" {
		t.Fatalf("token_type = %q, want Bearer", tokens.TokenType)
	}
}

func TestAuth_Login_InvalidCredentials(t *testing.T) {
	resp := testClient.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    adminEmail,
		"password": "WrongPassword1!",
	}, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
	resp.AssertErrorCode(t, "INVALID_CREDENTIALS")
}

func TestAuth_Login_ValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		body map[string]string
	}{
		{
			name: "invalid email format",
			body: map[string]string{"email": "not-an-email", "password": adminPassword},
		},
		{
			name: "password too short",
			body: map[string]string{"email": adminEmail, "password": "short"},
		},
		{
			name: "missing email",
			body: map[string]string{"password": adminPassword},
		},
		{
			name: "missing password",
			body: map[string]string{"email": adminEmail},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := testClient.do(http.MethodPost, "/api/v1/auth/login", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestAuth_Signup_Success(t *testing.T) {
	email := uniqueEmail("signup")
	resp := testClient.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      email,
		"password":   "CustomerPass1!",
		"first_name": "E2E",
		"last_name":  "Customer",
		"phone":      "+989121234567",
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var tokens tokenResponse
	resp.JSON(t, &tokens)
	if tokens.AccessToken == "" {
		t.Fatal("expected access_token")
	}

	meResp := testClient.withToken(tokens.AccessToken).do(http.MethodGet, "/api/v1/auth/me", nil, nil)
	meResp.AssertStatus(t, http.StatusOK)

	var profile struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
	}
	meResp.JSON(t, &profile)
	if profile.Email != email {
		t.Fatalf("email = %q, want %q", profile.Email, email)
	}
	if profile.Role != "customer" {
		t.Fatalf("role = %q, want customer", profile.Role)
	}
	if profile.FirstName != "E2E" || profile.LastName != "Customer" {
		t.Fatalf("unexpected profile name: %+v", profile)
	}
}

func TestAuth_Signup_DuplicateEmail(t *testing.T) {
	resp := testClient.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      adminEmail,
		"password":   "CustomerPass1!",
		"first_name": "Dup",
		"last_name":  "User",
	}, nil)
	resp.AssertStatus(t, http.StatusConflict)
	resp.AssertErrorCode(t, "CONFLICT")
}

func TestAuth_Signup_ValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		body map[string]string
	}{
		{
			name: "invalid email",
			body: map[string]string{
				"email": "bad", "password": "CustomerPass1!",
				"first_name": "A", "last_name": "B",
			},
		},
		{
			name: "short password",
			body: map[string]string{
				"email": uniqueEmail("short-pw"), "password": "123",
				"first_name": "A", "last_name": "B",
			},
		},
		{
			name: "missing first name",
			body: map[string]string{
				"email": uniqueEmail("no-fn"), "password": "CustomerPass1!",
				"last_name": "B",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := testClient.do(http.MethodPost, "/api/v1/auth/signup", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestAuth_RefreshToken(t *testing.T) {
	loginResp := testClient.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    adminEmail,
		"password": adminPassword,
	}, nil)
	loginResp.AssertStatus(t, http.StatusOK)

	var loginTokens tokenResponse
	loginResp.JSON(t, &loginTokens)

	refreshResp := testClient.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": loginTokens.RefreshToken,
	}, nil)
	refreshResp.AssertStatus(t, http.StatusOK)

	var refreshTokens tokenResponse
	refreshResp.JSON(t, &refreshTokens)
	if refreshTokens.AccessToken == "" || refreshTokens.RefreshToken == "" {
		t.Fatal("expected new token pair after refresh")
	}
	if refreshTokens.RefreshToken == loginTokens.RefreshToken {
		t.Fatal("expected refresh token rotation after refresh")
	}
}

func TestAuth_Refresh_InvalidToken(t *testing.T) {
	resp := testClient.do(http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "invalid-token",
	}, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestAuth_Me_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/auth/me", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestAuth_Me_AdminProfile(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/auth/me", nil, nil)
	resp.AssertStatus(t, http.StatusOK)

	var profile struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	resp.JSON(t, &profile)
	if profile.Email != adminEmail {
		t.Fatalf("email = %q, want %q", profile.Email, adminEmail)
	}
	if profile.Role != "admin" {
		t.Fatalf("role = %q, want admin", profile.Role)
	}
}

func TestAuth_ForgotPassword(t *testing.T) {
	cases := []struct {
		name  string
		email string
	}{
		{name: "existing admin", email: adminEmail},
		{name: "unknown email", email: uniqueEmail("unknown")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := testClient.do(http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
				"email": tc.email,
			}, nil)
			resp.AssertStatus(t, http.StatusOK)

			var result struct {
				Message string `json:"message"`
			}
			resp.JSON(t, &result)
			if result.Message == "" {
				t.Fatal("expected message in forgot-password response")
			}
		})
	}
}

func TestAuth_CustomerCannotAccessAdminRoutes(t *testing.T) {
	email := uniqueEmail("customer-deny")
	signupResp := testClient.do(http.MethodPost, "/api/v1/auth/signup", map[string]string{
		"email":      email,
		"password":   "CustomerPass1!",
		"first_name": "No",
		"last_name":  "Admin",
	}, nil)
	signupResp.AssertStatus(t, http.StatusCreated)

	var tokens tokenResponse
	signupResp.JSON(t, &tokens)

	resp := testClient.withToken(tokens.AccessToken).do(http.MethodGet, "/api/v1/admin/products", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}
