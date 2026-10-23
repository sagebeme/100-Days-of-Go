// Package deps pins the third-party modules the later phases use, so go.mod lists them all in
// one place and every day builds offline once `go mod download` has run. Nothing imports this
// package; the blank imports only keep `go mod tidy` from dropping the modules.
package deps

import (
	_ "github.com/alicebob/miniredis/v2"               // Day 78: an in-memory Redis for tests
	_ "github.com/coder/websocket"                     // Days 71 and 87: WebSockets
	_ "github.com/golang-jwt/jwt/v5"                   // Day 58: JWTs
	_ "github.com/jackc/pgx/v5"                        // Days 55-56: Postgres
	_ "github.com/jackc/pgx/v5/stdlib"                 // database/sql driver for pgx
	_ "github.com/prometheus/client_golang/prometheus" // Day 81: metrics
	_ "github.com/redis/go-redis/v9"                   // Day 78: caching
	_ "github.com/yuin/goldmark"                       // Day 89: Markdown to HTML
	_ "golang.org/x/crypto/argon2"                     // Day 57: password hashing
	_ "golang.org/x/crypto/bcrypt"                     // Day 57: password hashing
	_ "golang.org/x/time/rate"                         // Days 47 and 61: rate limiting
	_ "google.golang.org/grpc"                         // Day 80: gRPC
	_ "google.golang.org/protobuf/proto"               // Day 80: protobuf
)
