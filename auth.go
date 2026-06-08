package graphql

import (
	"errors"
	"fmt"
)

// SignUp - Create a new user, returning the user token on success.
func (c *Client) SignUp(auth AuthStruct) (*AuthResponse, error) {
	if auth.Username == "" || auth.Password == "" {
		return nil, fmt.Errorf("define username and password")
	}

	const query = `
		mutation ($username: String!, $password: String!) {
			signUp(username: $username, password: $password) {
				token
				# Add any other fields you expect to map to AuthResponse here.
			}
		}
	`

	vars := map[string]any{
		"username": auth.Username,
		"password": auth.Password,
	}

	var res struct {
		SignUp AuthResponse `json:"signUp"`
	}

	if err := c.doGraphQL(query, vars, nil, &res); err != nil {
		return nil, err
	}

	return &res.SignUp, nil
}

// SignIn - Get a new token for the user using the client's stored credentials.
func (c *Client) SignIn() (*AuthResponse, error) {
	if c.Auth.Username == "" || c.Auth.Password == "" {
		return nil, fmt.Errorf("define username and password")
	}

	const query = `
		mutation ($username: String!, $password: String!) {
			signIn(username: $username, password: $password) {
				token
			}
		}
	`

	vars := map[string]any{
		"username": c.Auth.Username,
		"password": c.Auth.Password,
	}

	var res struct {
		SignIn AuthResponse `json:"signIn"`
	}

	if err := c.doGraphQL(query, vars, nil, &res); err != nil {
		return nil, err
	}

	return &res.SignIn, nil
}

// GetUserTokenSignIn - Get a new token using the provided credentials.
func (c *Client) GetUserTokenSignIn(auth AuthStruct) (*AuthResponse, error) {
	if auth.Username == "" || auth.Password == "" {
		return nil, fmt.Errorf("define username and password")
	}

	const query = `
		mutation ($username: String!, $password: String!) {
			signIn(username: $username, password: $password) {
				token
			}
		}
	`

	vars := map[string]any{
		"username": auth.Username,
		"password": auth.Password,
	}

	var res struct {
		SignIn AuthResponse `json:"signIn"`
	}

	if err := c.doGraphQL(query, vars, nil, &res); err != nil {
		// Preserving the original error semantics for this method.
		return nil, errors.New("unable to login")
	}

	return &res.SignIn, nil
}

// SignOut - Revoke the token for a user.
//
// authToken overrides the client's stored token when non-nil.
func (c *Client) SignOut(authToken *string) error {
	const query = `
		mutation {
			signOut
		}
	`

	var res struct {
		SignOut string `json:"signOut"`
	}

	if err := c.doGraphQL(query, nil, authToken, &res); err != nil {
		return err
	}

	// Preserving the original string validation.
	if res.SignOut != "Signed out user" {
		return errors.New(res.SignOut)
	}

	return nil
}
