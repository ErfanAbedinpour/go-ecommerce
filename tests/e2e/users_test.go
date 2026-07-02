//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestUsers_Unauthorized(t *testing.T) {
	resp := testClient.do(http.MethodGet, "/api/v1/admin/users", nil, nil)
	resp.AssertStatus(t, http.StatusUnauthorized)
}

func TestUsers_Create_ValidationErrors(t *testing.T) {
	c := adminClient(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing email",
			body: map[string]any{
				"password":   "ValidPass1!",
				"first_name": "Test",
				"last_name":  "User",
			},
		},
		{
			name: "invalid email",
			body: map[string]any{
				"email":      "not-an-email",
				"password":   "ValidPass1!",
				"first_name": "Test",
				"last_name":  "User",
			},
		},
		{
			name: "password too short",
			body: map[string]any{
				"email":      uniqueEmail("short-pw"),
				"password":   "short",
				"first_name": "Test",
				"last_name":  "User",
			},
		},
		{
			name: "missing first name",
			body: map[string]any{
				"email":     uniqueEmail("no-fn"),
				"password":  "ValidPass1!",
				"last_name": "User",
			},
		},
		{
			name: "invalid role",
			body: map[string]any{
				"email":      uniqueEmail("bad-role"),
				"password":   "ValidPass1!",
				"first_name": "Test",
				"last_name":  "User",
				"role":       "superadmin",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := c.do(http.MethodPost, "/api/v1/admin/users", tc.body, nil)
			resp.AssertStatus(t, http.StatusBadRequest)
			resp.AssertErrorCode(t, "VALIDATION_ERROR")
		})
	}
}

func TestUsers_CRUD_Lifecycle(t *testing.T) {
	c := adminClient(t)
	email := uniqueEmail("admin-user")

	createBody := map[string]any{
		"email":      email,
		"password":   "StaffPass1!",
		"first_name": "E2E",
		"last_name":  "Staff",
		"phone":      "+989121111111",
		"role":       "admin",
		"is_active":  true,
	}

	createResp := c.do(http.MethodPost, "/api/v1/admin/users", createBody, nil)
	createResp.AssertStatus(t, http.StatusCreated)

	var created struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		FullName  string `json:"full_name"`
		Phone     string `json:"phone"`
		Role      string `json:"role"`
		IsActive  bool   `json:"is_active"`
	}
	createResp.JSON(t, &created)

	if created.ID == "" {
		t.Fatal("expected user id")
	}
	if created.Email != email {
		t.Fatalf("email = %q, want %q", created.Email, email)
	}
	if created.FirstName != "E2E" || created.LastName != "Staff" {
		t.Fatalf("unexpected name: %+v", created)
	}
	if created.FullName != "E2E Staff" {
		t.Fatalf("full_name = %q, want %q", created.FullName, "E2E Staff")
	}
	if created.Phone != "+989121111111" {
		t.Fatalf("phone = %q", created.Phone)
	}
	if created.Role != "admin" {
		t.Fatalf("role = %q, want admin", created.Role)
	}
	if !created.IsActive {
		t.Fatal("expected is_active=true")
	}

	getResp := c.do(http.MethodGet, "/api/v1/admin/users/"+created.ID, nil, nil)
	getResp.AssertStatus(t, http.StatusOK)

	var fetched struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	getResp.JSON(t, &fetched)
	if fetched.ID != created.ID || fetched.Email != email {
		t.Fatalf("get mismatch: %+v", fetched)
	}

	listResp := c.do(http.MethodGet, "/api/v1/admin/users?page=1&per_page=10&q=E2E", nil, nil)
	listResp.AssertStatus(t, http.StatusOK)

	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta struct {
			Page    int   `json:"page"`
			PerPage int   `json:"per_page"`
			Total   int64 `json:"total"`
		} `json:"meta"`
	}
	listResp.JSON(t, &list)
	if list.Meta.Page != 1 || list.Meta.PerPage != 10 {
		t.Fatalf("unexpected pagination meta: %+v", list.Meta)
	}
	if len(list.Data) == 0 {
		t.Fatal("expected at least one user in search results")
	}

	newEmail := uniqueEmail("admin-updated")
	updateResp := c.do(http.MethodPut, "/api/v1/admin/users/"+created.ID, map[string]any{
		"email":      newEmail,
		"first_name": "Updated",
		"last_name":  "Staff",
		"phone":      "+989122222222",
		"is_active":  true,
	}, nil)
	updateResp.AssertStatus(t, http.StatusOK)

	var updated struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		FullName  string `json:"full_name"`
		IsActive  bool   `json:"is_active"`
	}
	updateResp.JSON(t, &updated)
	if updated.Email != newEmail {
		t.Fatalf("updated email = %q", updated.Email)
	}
	if updated.FirstName != "Updated" || updated.LastName != "Staff" {
		t.Fatalf("unexpected updated name: %+v", updated)
	}
	if updated.FullName != "Updated Staff" {
		t.Fatalf("full_name = %q, want Updated Staff", updated.FullName)
	}
	if !updated.IsActive {
		t.Fatal("expected is_active=true after update")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/users/"+created.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusNoContent)

	notFoundResp := c.do(http.MethodGet, "/api/v1/admin/users/"+created.ID, nil, nil)
	notFoundResp.AssertStatus(t, http.StatusNotFound)
	notFoundResp.AssertErrorCode(t, "NOT_FOUND")
}

func TestUsers_DuplicateEmail(t *testing.T) {
	c := adminClient(t)
	email := uniqueEmail("dup-admin")

	firstResp := c.do(http.MethodPost, "/api/v1/admin/users", map[string]any{
		"email":      email,
		"password":   "StaffPass1!",
		"first_name": "First",
		"last_name":  "User",
	}, nil)
	firstResp.AssertStatus(t, http.StatusCreated)

	var first struct {
		ID string `json:"id"`
	}
	firstResp.JSON(t, &first)
	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, "/api/v1/admin/users/"+first.ID, nil, nil)
	})

	secondResp := c.do(http.MethodPost, "/api/v1/admin/users", map[string]any{
		"email":      email,
		"password":   "StaffPass2!",
		"first_name": "Second",
		"last_name":  "User",
	}, nil)
	secondResp.AssertStatus(t, http.StatusConflict)
	secondResp.AssertErrorCode(t, "CONFLICT")
}

func TestUsers_CannotDeleteSelf(t *testing.T) {
	c := adminClient(t)

	meResp := c.do(http.MethodGet, "/api/v1/auth/me", nil, nil)
	meResp.AssertStatus(t, http.StatusOK)

	var me struct {
		ID string `json:"id"`
	}
	meResp.JSON(t, &me)
	if me.ID == "" {
		t.Fatal("expected admin id from /auth/me")
	}

	deleteResp := c.do(http.MethodDelete, "/api/v1/admin/users/"+me.ID, nil, nil)
	deleteResp.AssertStatus(t, http.StatusUnprocessableEntity)
	deleteResp.AssertErrorCode(t, "UNPROCESSABLE_ENTITY")
}

func TestUsers_CustomerForbidden(t *testing.T) {
	customer := customerClient(t)
	resp := customer.do(http.MethodGet, "/api/v1/admin/users", nil, nil)
	resp.AssertStatus(t, http.StatusForbidden)
	resp.AssertErrorCode(t, "FORBIDDEN")
}

func TestUsers_Get_NotFound(t *testing.T) {
	c := adminClient(t)
	resp := c.do(http.MethodGet, "/api/v1/admin/users/00000000-0000-0000-0000-000000000099", nil, nil)
	resp.AssertStatus(t, http.StatusNotFound)
	resp.AssertErrorCode(t, "NOT_FOUND")
}

func TestUsers_DefaultRoleIsAdmin(t *testing.T) {
	c := adminClient(t)
	email := uniqueEmail("default-role")

	resp := c.do(http.MethodPost, "/api/v1/admin/users", map[string]any{
		"email":      email,
		"password":   "StaffPass1!",
		"first_name": "Default",
		"last_name":  "Admin",
	}, nil)
	resp.AssertStatus(t, http.StatusCreated)

	var user struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	resp.JSON(t, &user)
	if user.Role != "admin" {
		t.Fatalf("role = %q, want admin", user.Role)
	}

	t.Cleanup(func() {
		_ = c.do(http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%s", user.ID), nil, nil)
	})
}
