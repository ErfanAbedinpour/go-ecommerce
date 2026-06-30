package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"app/internal/config"
	"app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/internal/domain/user"
	infraauth "app/internal/infrastructure/auth"
	"app/pkg/pagination"
)

type mockUserRepo struct {
	users map[string]*user.User
	byID  map[uuid.UUID]*user.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*user.User),
		byID:  make(map[uuid.UUID]*user.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, u *user.User) error {
	cp := *u
	m.users[u.Email] = &cp
	m.byID[u.ID] = &cp
	return nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) UpdateLastLogin(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockUserRepo) UpdatePassword(_ context.Context, id uuid.UUID, passwordHash string) error {
	u, ok := m.byID[id]
	if !ok {
		return user.ErrNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}

func (m *mockUserRepo) List(context.Context, user.ListFilter, pagination.Params) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) Update(context.Context, *user.User) error { return nil }
func (m *mockUserRepo) SoftDelete(context.Context, uuid.UUID) error {
	return nil
}
func (m *mockUserRepo) CountByRole(context.Context, user.Role) (int64, error) {
	return 0, nil
}

type mockRefreshRepo struct {
	created int
	revoked int
}

func (m *mockRefreshRepo) Create(_ context.Context, _ *user.RefreshToken) error {
	m.created++
	return nil
}

func (m *mockRefreshRepo) FindByHash(_ context.Context, _ string) (*user.RefreshToken, error) {
	return nil, user.ErrInvalidToken
}

func (m *mockRefreshRepo) RevokeFamily(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockRefreshRepo) Revoke(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockRefreshRepo) RevokeAllByUser(_ context.Context, _ uuid.UUID) error {
	m.revoked++
	return nil
}

type mockResetRepo struct {
	tokens map[string]*user.PasswordResetToken
}

func newMockResetRepo() *mockResetRepo {
	return &mockResetRepo{tokens: make(map[string]*user.PasswordResetToken)}
}

func (m *mockResetRepo) Create(_ context.Context, token *user.PasswordResetToken) error {
	cp := *token
	m.tokens[token.TokenHash] = &cp
	return nil
}

func (m *mockResetRepo) FindByHash(_ context.Context, hash string) (*user.PasswordResetToken, error) {
	t, ok := m.tokens[hash]
	if !ok {
		return nil, user.ErrInvalidResetToken
	}
	cp := *t
	return &cp, nil
}

func (m *mockResetRepo) MarkUsed(_ context.Context, id uuid.UUID) error {
	for _, t := range m.tokens {
		if t.ID == id {
			now := time.Now().UTC()
			t.UsedAt = &now
			return nil
		}
	}
	return nil
}

func (m *mockResetRepo) InvalidateByUser(_ context.Context, userID uuid.UUID) error {
	for _, t := range m.tokens {
		if t.UserID == userID && t.UsedAt == nil {
			now := time.Now().UTC()
			t.UsedAt = &now
		}
	}
	return nil
}

type mockMailer struct {
	sent []string
}

func (m *mockMailer) SendPasswordReset(_ context.Context, to, _ string) error {
	m.sent = append(m.sent, to)
	return nil
}

func (m *mockMailer) SendOrderConfirmation(_ context.Context, _ string, _ string, _ float64) error {
	return nil
}

type mockJWT struct{}

func (mockJWT) GenerateTokenPair(_ infraauth.TokenInput) (*infraauth.TokenPair, string, uuid.UUID, error) {
	return &infraauth.TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	}, "hash", uuid.New(), nil
}

func (mockJWT) ValidateRefreshToken(_ string) (*infraauth.Claims, error) {
	return nil, user.ErrInvalidToken
}

func (mockJWT) RefreshTokenTTL() time.Duration {
	return time.Hour
}

type mockCustomerRepo struct {
	created int
}

func (m *mockCustomerRepo) Create(_ context.Context, _ *customer.Customer) error {
	m.created++
	return nil
}
func (m *mockCustomerRepo) FindByID(context.Context, uuid.UUID) (*customer.Customer, error) {
	return nil, customer.ErrNotFound
}
func (m *mockCustomerRepo) FindByEmail(context.Context, string) (*customer.Customer, error) {
	return nil, customer.ErrNotFound
}
func (m *mockCustomerRepo) FindByUserID(context.Context, uuid.UUID) (*customer.Customer, error) {
	return nil, customer.ErrNotFound
}
func (m *mockCustomerRepo) List(context.Context, customer.ListFilter, pagination.Params) ([]customer.Customer, int64, error) {
	return nil, 0, nil
}
func (m *mockCustomerRepo) Update(context.Context, *customer.Customer) error { return nil }
func (m *mockCustomerRepo) Delete(context.Context, uuid.UUID) error          { return nil }
func (m *mockCustomerRepo) HasOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockCustomerRepo) GetLastOrderAt(context.Context, uuid.UUID) (*time.Time, error) {
	return nil, nil
}
func (m *mockCustomerRepo) ListAddresses(context.Context, uuid.UUID) ([]customer.Address, error) {
	return nil, nil
}
func (m *mockCustomerRepo) ListOrders(context.Context, uuid.UUID, pagination.Params) ([]domainorder.Summary, int64, error) {
	return nil, 0, nil
}

func newTestService(signupEnabled bool, defaultRole string) (*AuthService, *mockUserRepo, *mockResetRepo, *mockMailer) {
	hasher := infraauth.NewPasswordHasher()
	users := newMockUserRepo()
	resetRepo := newMockResetRepo()
	mailer := &mockMailer{}

	svc := NewAuthService(
		users,
		&mockCustomerRepo{},
		&mockRefreshRepo{},
		resetRepo,
		hasher,
		mockJWT{},
		mailer,
		config.AuthConfig{
			SignupEnabled:     signupEnabled,
			SignupDefaultRole: defaultRole,
			ResetTokenTTL:     time.Hour,
			AppURL:            "http://localhost:5173",
			ResetPath:         "/reset-password",
		},
	)

	return svc, users, resetRepo, mailer
}

func TestSignup_Success(t *testing.T) {
	customers := &mockCustomerRepo{}
	svc := newTestServiceWithCustomers(true, "customer", customers)

	out, err := svc.Signup(context.Background(), SignupInput{
		Email:     "new@shop.com",
		Password:  "Secret123",
		FirstName: "New",
		LastName:  "User",
	})
	if err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if out.AccessToken != "access" {
		t.Fatalf("expected access token, got %q", out.AccessToken)
	}
	if customers.created != 1 {
		t.Fatalf("expected customer record on signup, created = %d", customers.created)
	}
}

func TestSignup_AdminRoleSkipsCustomerRecord(t *testing.T) {
	customers := &mockCustomerRepo{}
	svc := newTestServiceWithCustomers(true, "admin", customers)

	_, err := svc.Signup(context.Background(), SignupInput{
		Email:     "admin@shop.com",
		Password:  "Secret123",
		FirstName: "Admin",
		LastName:  "User",
	})
	if err != nil {
		t.Fatalf("Signup() error = %v", err)
	}
	if customers.created != 0 {
		t.Fatalf("expected no customer record for admin signup, created = %d", customers.created)
	}
}

func newTestServiceWithCustomers(signupEnabled bool, defaultRole string, customers *mockCustomerRepo) *AuthService {
	hasher := infraauth.NewPasswordHasher()
	users := newMockUserRepo()
	resetRepo := newMockResetRepo()
	mailer := &mockMailer{}

	return NewAuthService(
		users,
		customers,
		&mockRefreshRepo{},
		resetRepo,
		hasher,
		mockJWT{},
		mailer,
		config.AuthConfig{
			SignupEnabled:     signupEnabled,
			SignupDefaultRole: defaultRole,
			ResetTokenTTL:     time.Hour,
			AppURL:            "http://localhost:5173",
			ResetPath:         "/reset-password",
		},
	)
}

func TestSignup_EmailTaken(t *testing.T) {
	svc, users, _, _ := newTestService(true, "customer")
	id := uuid.New()
	users.users["taken@shop.com"] = &user.User{
		ID:    id,
		Email: "taken@shop.com",
		Role:  user.RoleCustomer,
	}
	users.byID[id] = users.users["taken@shop.com"]

	_, err := svc.Signup(context.Background(), SignupInput{
		Email:     "taken@shop.com",
		Password:  "Secret123",
		FirstName: "Taken",
		LastName:  "User",
	})
	if err != user.ErrEmailTaken {
		t.Fatalf("Signup() error = %v, want ErrEmailTaken", err)
	}
}

func TestSignup_Disabled(t *testing.T) {
	svc, _, _, _ := newTestService(false, "customer")

	_, err := svc.Signup(context.Background(), SignupInput{
		Email:     "new@shop.com",
		Password:  "Secret123",
		FirstName: "New",
		LastName:  "User",
	})
	if err != user.ErrSignupDisabled {
		t.Fatalf("Signup() error = %v, want ErrSignupDisabled", err)
	}
}

func TestForgotPassword_SendsEmail(t *testing.T) {
	svc, users, _, mailer := newTestService(true, "customer")
	id := uuid.New()
	users.users["user@shop.com"] = &user.User{
		ID:       id,
		Email:    "user@shop.com",
		Role:     user.RoleCustomer,
		IsActive: true,
	}
	users.byID[id] = users.users["user@shop.com"]

	out, err := svc.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "user@shop.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if out.Message != forgotPasswordMessage {
		t.Fatalf("unexpected message: %q", out.Message)
	}
	if len(mailer.sent) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.sent))
	}
}

func TestForgotPassword_UnknownEmail(t *testing.T) {
	svc, _, _, mailer := newTestService(true, "customer")

	out, err := svc.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "missing@shop.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if out.Message != forgotPasswordMessage {
		t.Fatalf("unexpected message: %q", out.Message)
	}
	if len(mailer.sent) != 0 {
		t.Fatalf("expected no emails, got %d", len(mailer.sent))
	}
}

func TestResetPassword_Success(t *testing.T) {
	svc, users, resetRepo, _ := newTestService(true, "customer")
	hasher := infraauth.NewPasswordHasher()
	oldHash, _ := hasher.Hash("OldPass123")

	id := uuid.New()
	users.users["user@shop.com"] = &user.User{
		ID:           id,
		Email:        "user@shop.com",
		PasswordHash: oldHash,
		Role:         user.RoleCustomer,
		IsActive:     true,
	}
	users.byID[id] = users.users["user@shop.com"]

	rawToken := "reset-token-abc"
	resetRepo.tokens[infraauth.HashToken(rawToken)] = &user.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    id,
		TokenHash: infraauth.HashToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		CreatedAt: time.Now().UTC(),
	}

	out, err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:    rawToken,
		Password: "NewPass456",
	})
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}
	if out.Message != resetPasswordMessage {
		t.Fatalf("unexpected message: %q", out.Message)
	}
	if !hasher.Verify(users.byID[id].PasswordHash, "NewPass456") {
		t.Fatal("password was not updated")
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	svc, _, _, _ := newTestService(true, "customer")

	_, err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:    "bad-token",
		Password: "NewPass456",
	})
	if err != user.ErrInvalidResetToken {
		t.Fatalf("ResetPassword() error = %v, want ErrInvalidResetToken", err)
	}
}
