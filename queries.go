package graphql

import "context"

// Every operation in this package is a static, readable GraphQL document.
// Caller-supplied values are passed via the variables map ($name), never
// interpolated into the query string.

const queryMe = `
query Me {
  me {
    __typename
    ... on User {
      id
      username
      email
    }
    ... on ServiceAccount {
      id
      name
    }
  }
}`

const queryGetUser = `
query GetUser($username: String!) {
  user(username: $username) {
    id
    username
    email
  }
}`

// Me returns the identity behind the client's current token — a User or a
// ServiceAccount, distinguished by Identity.Typename. Requires auth.
func (c *Client) Me(ctx context.Context) (*Identity, error) {
	var out struct {
		Me *Identity `json:"me"`
	}
	if err := c.do(ctx, queryMe, "Me", nil, &out); err != nil {
		return nil, err
	}
	return out.Me, nil
}

// GetUser looks up a single user by username. Returns (nil, nil) when the
// user does not exist, mirroring the server's nullable return. Requires auth.
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	var out struct {
		User *User `json:"user"`
	}
	vars := map[string]interface{}{
		"username": username,
	}
	if err := c.do(ctx, queryGetUser, "GetUser", vars, &out); err != nil {
		return nil, err
	}
	return out.User, nil
}
