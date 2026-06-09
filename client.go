package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// defaultHostURL matches graphql-server-go's default listen address.
const defaultHostURL = "http://localhost:8080"

// Client talks to the graphql-server-go identity server. All operations
// go through a single POST {HostURL}/graphql endpoint carrying a standard
// GraphQL request body: {"query": ..., "operationName": ..., "variables": ...}.
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string // Bearer JWT from SignIn / LoginServiceAccount
	Auth       AuthStruct
}

// NewClient builds a client and, when credentials are supplied, eagerly
// signs in (mutation signIn) so c.Token is usable immediately. Pass nil
// username/password to get an unauthenticated client (enough for the open
// mutations: signIn, createUser, createServiceAccount, loginServiceAccount).
func NewClient(host, username, password *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    defaultHostURL,
	}

	if host != nil {
		c.HostURL = strings.TrimRight(*host, "/")
	}

	if username == nil || password == nil {
		return &c, nil
	}

	c.Auth = AuthStruct{
		Username: *username,
		Password: *password,
	}

	if _, err := c.SignInContext(context.Background(), c.Auth); err != nil {
		return nil, err
	}

	return &c, nil
}

// gqlRequest is the wire format graphql-server-go's handler decodes.
type gqlRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors"`
}

// do executes one GraphQL operation. Values travel exclusively in the
// variables map — operation strings are static and never interpolated.
// out, when non-nil, receives the unmarshalled "data" object.
func (c *Client) do(ctx context.Context, query, operationName string, variables map[string]interface{}, out interface{}) error {
	payload, err := json.Marshal(gqlRequest{
		Query:         query,
		OperationName: operationName,
		Variables:     variables,
	})
	if err != nil {
		return fmt.Errorf("encoding graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.HostURL+"/graphql", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("graphql endpoint returned status %d: %s", res.StatusCode, body)
	}

	var gr gqlResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return fmt.Errorf("decoding graphql response: %w", err)
	}

	if len(gr.Errors) > 0 {
		msgs := make([]string, 0, len(gr.Errors))
		for _, e := range gr.Errors {
			msgs = append(msgs, e.Message)
		}
		return errors.New("graphql: " + strings.Join(msgs, "; "))
	}

	if out != nil && gr.Data != nil {
		if err := json.Unmarshal(gr.Data, out); err != nil {
			return fmt.Errorf("decoding graphql data: %w", err)
		}
	}

	return nil
}
