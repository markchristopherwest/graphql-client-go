package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HostURL - Default Hashicups URL.
// Note: this must point at the GraphQL endpoint itself; append your route if
// the server expects one (e.g. /graphql or /query).
const HostURL string = "http://localhost:19090"

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
	Auth       AuthStruct
}

// AuthStruct -
type AuthStruct struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse -
type AuthResponse struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// graphQLRequest is the standard GraphQL-over-HTTP request envelope.
type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphQLError is a single error entry returned by a GraphQL server.
type graphQLError struct {
	Message string `json:"message"`
}

// graphQLResponse is the standard GraphQL-over-HTTP response envelope.
type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors,omitempty"`
}

// NewClient -
func NewClient(host, username, password *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	// If username or password not provided, return empty client.
	if username == nil || password == nil {
		return &c, nil
	}

	c.Auth = AuthStruct{
		Username: *username,
		Password: *password,
	}

	ar, err := c.SignIn()
	if err != nil {
		return nil, err
	}

	c.Token = ar.Token

	return &c, nil
}

// doGraphQL executes a single GraphQL operation against c.HostURL and
// unmarshals the "data" portion of the response into out.
//
// authToken overrides c.Token for this request when non-nil; pass nil to use
// the client's stored token (or no token at all, e.g. for sign-in/sign-up).
func (c *Client) doGraphQL(query string, variables map[string]any, authToken *string, out any) error {
	payload, err := json.Marshal(graphQLRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.HostURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := c.doRequest(req, authToken)
	if err != nil {
		return err
	}

	var resp graphQLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("decoding graphql response: %w", err)
	}

	if len(resp.Errors) > 0 {
		msgs := make([]string, 0, len(resp.Errors))
		for _, e := range resp.Errors {
			msgs = append(msgs, e.Message)
		}
		return fmt.Errorf("graphql: %s", strings.Join(msgs, "; "))
	}

	if out != nil {
		if err := json.Unmarshal(resp.Data, out); err != nil {
			return fmt.Errorf("decoding graphql data: %w", err)
		}
	}

	return nil
}

func (c *Client) doRequest(req *http.Request, authToken *string) ([]byte, error) {
	token := c.Token

	if authToken != nil {
		token = *authToken
	}

	// Only attach the header when we actually have a token, so unauthenticated
	// operations (sign-up / sign-in) don't send an empty Authorization header.
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, nil
}
