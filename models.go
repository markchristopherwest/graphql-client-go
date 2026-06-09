package graphql

// AuthStruct carries the credentials handed to NewClient / SignIn.
// Field names match the server's SignInInput.
type AuthStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthPayload mirrors the server's AuthPayload type (signIn /
// loginServiceAccount return it).
type AuthPayload struct {
	Token string `json:"token"`
}

// AuthResponse is the enriched result of a sign-in: the JWT from
// AuthPayload plus the principal identity resolved via the `me` query.
// NOTE: graphql-server-go uses string IDs ("usr_...", "sa_...").
type AuthResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// User mirrors the server's User object type.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ServiceAccount mirrors the server's ServiceAccount object type.
type ServiceAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Identity is the decoded form of the server's Identity union
// (User | ServiceAccount), as returned by the `me` query. Typename
// distinguishes which arm populated the struct.
type Identity struct {
	Typename string `json:"__typename"`
	ID       string `json:"id"`
	Username string `json:"username,omitempty"` // User
	Email    string `json:"email,omitempty"`    // User
	Name     string `json:"name,omitempty"`     // ServiceAccount
}

// CreateUserInput mirrors the server's CreateUserInput.
type CreateUserInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

// UpdateUserInput mirrors the server's UpdateUserInput. Empty fields
// are omitted so the server only touches what the caller sets.
type UpdateUserInput struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// CreateServiceAccountInput mirrors the server's CreateServiceAccountInput.
type CreateServiceAccountInput struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
