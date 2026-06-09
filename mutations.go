package graphql

import "context"

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

const mutationSignIn = `
mutation SignIn($input: SignInInput!) {
  signIn(input: $input) {
    token
  }
}`

// SignInContext exchanges username/password for a JWT (mutation signIn),
// stores it on the client, then resolves the principal via Me so the
// AuthResponse carries the user's id and username alongside the token.
func (c *Client) SignInContext(ctx context.Context, auth AuthStruct) (*AuthResponse, error) {
	var out struct {
		SignIn AuthPayload `json:"signIn"`
	}
	vars := map[string]interface{}{
		"input": auth,
	}
	if err := c.do(ctx, mutationSignIn, "SignIn", vars, &out); err != nil {
		return nil, err
	}

	c.Token = out.SignIn.Token

	resp := &AuthResponse{
		Username: auth.Username,
		Token:    out.SignIn.Token,
	}

	// Best effort: enrich with the server-side identity. The token is
	// already valid even if this lookup fails.
	if id, err := c.Me(ctx); err == nil && id != nil {
		resp.UserID = id.ID
		if id.Username != "" {
			resp.Username = id.Username
		}
	}

	return resp, nil
}

// SignIn signs in with the credentials stored on the client (c.Auth).
// Kept for callers built against the previous client surface, e.g.
// vault-plugin-secrets-graphql.
func (c *Client) SignIn() (*AuthResponse, error) {
	return c.SignInContext(context.Background(), c.Auth)
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

const mutationCreateUser = `
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    username
    email
  }
}`

const mutationUpdateUser = `
mutation UpdateUser($username: String!, $input: UpdateUserInput!) {
  updateUser(username: $username, input: $input) {
    id
    username
    email
  }
}`

const mutationSetPassword = `
mutation SetPassword($username: String!, $password: String!) {
  setPassword(username: $username, password: $password) {
    id
    username
  }
}`

const mutationDeleteUser = `
mutation DeleteUser($username: String!) {
  deleteUser(username: $username)
}`

// CreateUser creates a user (open mutation: sign-up / provisioning).
func (c *Client) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	var out struct {
		CreateUser *User `json:"createUser"`
	}
	vars := map[string]interface{}{
		"input": input,
	}
	if err := c.do(ctx, mutationCreateUser, "CreateUser", vars, &out); err != nil {
		return nil, err
	}
	return out.CreateUser, nil
}

// UpdateUser updates a user's email and/or password. Requires auth.
func (c *Client) UpdateUser(ctx context.Context, username string, input UpdateUserInput) (*User, error) {
	var out struct {
		UpdateUser *User `json:"updateUser"`
	}
	vars := map[string]interface{}{
		"username": username,
		"input":    input,
	}
	if err := c.do(ctx, mutationUpdateUser, "UpdateUser", vars, &out); err != nil {
		return nil, err
	}
	return out.UpdateUser, nil
}

// SetPassword sets a user's password. Requires auth. This is the server's
// designated root-rotation primitive for vault-plugin-secrets-graphql.
func (c *Client) SetPassword(ctx context.Context, username, password string) (*User, error) {
	var out struct {
		SetPassword *User `json:"setPassword"`
	}
	vars := map[string]interface{}{
		"username": username,
		"password": password,
	}
	if err := c.do(ctx, mutationSetPassword, "SetPassword", vars, &out); err != nil {
		return nil, err
	}
	return out.SetPassword, nil
}

// DeleteUser deletes a user by username. Requires auth. Idempotent on the
// server, so Vault lease revocation can safely retry it.
func (c *Client) DeleteUser(ctx context.Context, username string) (bool, error) {
	var out struct {
		DeleteUser bool `json:"deleteUser"`
	}
	vars := map[string]interface{}{
		"username": username,
	}
	if err := c.do(ctx, mutationDeleteUser, "DeleteUser", vars, &out); err != nil {
		return false, err
	}
	return out.DeleteUser, nil
}

// ---------------------------------------------------------------------------
// Service accounts
// ---------------------------------------------------------------------------

const mutationCreateServiceAccount = `
mutation CreateServiceAccount($input: CreateServiceAccountInput!) {
  createServiceAccount(input: $input) {
    id
    name
  }
}`

const mutationLoginServiceAccount = `
mutation LoginServiceAccount($id: String!, $password: String!) {
  loginServiceAccount(id: $id, password: $password) {
    token
  }
}`

const mutationDeleteServiceAccount = `
mutation DeleteServiceAccount($id: String!) {
  deleteServiceAccount(id: $id)
}`

// CreateServiceAccount creates a service account with a caller-supplied
// password (open mutation). This is the create half of Vault's dynamic
// credential flow: one service account per lease.
func (c *Client) CreateServiceAccount(ctx context.Context, input CreateServiceAccountInput) (*ServiceAccount, error) {
	var out struct {
		CreateServiceAccount *ServiceAccount `json:"createServiceAccount"`
	}
	vars := map[string]interface{}{
		"input": input,
	}
	if err := c.do(ctx, mutationCreateServiceAccount, "CreateServiceAccount", vars, &out); err != nil {
		return nil, err
	}
	return out.CreateServiceAccount, nil
}

// LoginServiceAccount exchanges a service account id/password for a JWT.
// It does not mutate c.Token: the returned credential belongs to the
// service account (e.g. a Vault lease), not to this client's session.
func (c *Client) LoginServiceAccount(ctx context.Context, id, password string) (*AuthResponse, error) {
	var out struct {
		LoginServiceAccount AuthPayload `json:"loginServiceAccount"`
	}
	vars := map[string]interface{}{
		"id":       id,
		"password": password,
	}
	if err := c.do(ctx, mutationLoginServiceAccount, "LoginServiceAccount", vars, &out); err != nil {
		return nil, err
	}
	return &AuthResponse{
		UserID: id,
		Token:  out.LoginServiceAccount.Token,
	}, nil
}

// DeleteServiceAccount deletes a service account by id. Requires auth.
// Idempotent on the server — the revoke half of Vault's dynamic
// credential flow.
func (c *Client) DeleteServiceAccount(ctx context.Context, id string) (bool, error) {
	var out struct {
		DeleteServiceAccount bool `json:"deleteServiceAccount"`
	}
	vars := map[string]interface{}{
		"id": id,
	}
	if err := c.do(ctx, mutationDeleteServiceAccount, "DeleteServiceAccount", vars, &out); err != nil {
		return false, err
	}
	return out.DeleteServiceAccount, nil
}
