# Graphql Client (Go)

A Go client package that provides a transient connection between [terraform-provider-graphql](https://github.com/markchristopherwest/terraform-provider-graphql) and [product-api](https://github.com/markchristopherwest/product-api-go). You do not need to compile this package; the GraphQL provider uses it as a dependency. 


![Graphql diagram](./docs/terraform-provider-diagram.jpeg)

Using this module, the Graphql provider establishes a new client and sends HTTP(s) requests to the product api application to perform CRUD operations. It also handles data mapping from user's inputs to `models.go`. The Graphql URL defaults to `http://localhost:19090` and you can configure it [here](https://github.com/markchristopherwest/graphql-client-go/blob/main/client.go#L11) in case that port is already in use. This module also handles API calls to sign up, sign in and sign out for users authentication.