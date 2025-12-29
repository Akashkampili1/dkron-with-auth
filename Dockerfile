# ---------- Build stage ----------
FROM golang:1.24-alpine AS build

WORKDIR /src

# Install git for private/public modules
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o dkron-with-auth main.go


# ---------- Runtime stage ----------
FROM alpine:3.19

WORKDIR /app

# (Optional but recommended) CA certs for HTTPS calls (auth, OIDC, APIs)
RUN apk add --no-cache ca-certificates

COPY --from=build /src/dkron-with-auth /usr/local/bin/dkron-with-auth
COPY --from=build /src/config /app/config

EXPOSE 8080 8946 6868

ENTRYPOINT ["dkron-with-auth","agent"]
CMD ["--server","--bootstrap-expect","1"]
