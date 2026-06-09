# graphql-server-go

A small, dependency-light GraphQL server in Go that exposes **CRUD for user
accounts and service accounts** with bearer-token auth. It is the backend the
[`vault-plugin-secrets-graphql`](https://github.com/markchristopherwest/vault-plugin-secrets-graphql)
secrets engine drives: the plugin's configured admin token authenticates every
request, and the plugin mints/revokes accounts through it.

## Run

```sh
go mod tidy          # resolve github.com/graphql-go/graphql
go run ./...         # listens on :9090, POST /query
```

Environment:

| Var | Default | Meaning |
| --- | --- | --- |
| `PORT` | `9090` | HTTP listen port |
| `GQL_ADMIN_TOKEN` | `root-admin-token` | Bearer token required for mutations / listing. Configure this same value into the Vault plugin. |

`GET /healthz` returns `ok` (no auth). `POST /query` is the GraphQL endpoint.

## Auth model

Mutations and the `user(s)` / `serviceAccount(s)` queries require
`Authorization: Bearer <GQL_ADMIN_TOKEN>`. `health` / `whoami` are public.
Created accounts receive generated secret material (`password`/`token` for
users, `clientId`/`clientSecret`/`token` for service accounts) returned only at
create/rotate time — never re-readable, which mirrors how a real IdP issues
credentials and is exactly what the Vault dynamic secret captures.

## Try it

```sh
TOKEN=root-admin-token

# create a user (returns secret material)
curl -s localhost:9090/query \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"query":"mutation($i:CreateUserInput!){createUser(input:$i){user{id username role} password token}}","variables":{"i":{"username":"alice","role":"reader"}}}'

# create a service account
curl -s localhost:9090/query \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"query":"mutation($i:CreateServiceAccountInput!){createServiceAccount(input:$i){serviceAccount{id name scopes} clientId clientSecret token}}","variables":{"i":{"name":"ci-runner","scopes":["read","deploy"]}}}'

# list / delete
curl -s localhost:9090/query -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"query":"{users{id username}}"}'
curl -s localhost:9090/query -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"query":"mutation($id:ID!){deleteUser(id:$id)}","variables":{"id":"usr_..."}}'
```

## Notes

- Storage is in-memory (`store.go`) so it runs anywhere with no DB. The method
  set is small; back it with Postgres/etc. without touching `schema.go`.
- Build is fully static (`CGO_ENABLED=0`) for the Alpine image.