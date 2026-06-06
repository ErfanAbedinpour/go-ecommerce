package response

// TokenResponse is the authentication token pair returned on login/refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresIn    int64  `json:"expires_in" example:"900"`
	TokenType    string `json:"token_type" example:"Bearer"`
}

// CurrentUserResponse is the authenticated user profile.
type CurrentUserResponse struct {
	ID        string `json:"id" example:"c0000000-0000-0000-0000-000000000001"`
	Email     string `json:"email" example:"admin@shop.com"`
	FirstName string `json:"first_name" example:"Admin"`
	LastName  string `json:"last_name" example:"User"`
	Phone     string `json:"phone,omitempty" example:"+1234567890"`
	Role      string `json:"role" example:"admin" enums:"admin,customer"`
}
