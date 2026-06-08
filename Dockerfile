# syntax=docker/dockerfile:1
# Explicitly use the host's native platform as the build platform
FROM --platform=${BUILDPLATFORM} golang:1.26.4-alpine AS builder

WORKDIR /app

# Download Go dependencies first to leverage caching
# COPY go.mod go.sum ./
RUN go mod init github.com/markchristopherwest/graphql-client-go

# Copy the rest of the source code
COPY . .

# Dynamically map the BuildKit TARGETOS/TARGETARCH to Go's GOOS/GOARCH
# and disable CGO for a purely static binary
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app/graphql-client .

# Final minimal stage (e.g., distroless or scratch)
FROM scratch
COPY --from=builder /app/graphql-client /graphql-client

ENTRYPOINT [ "sh", "-c", "/graphql-client"] 
