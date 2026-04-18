# Step 1: Modules caching
FROM golang:1.25 AS modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Builder
FROM golang:1.25 AS builder
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /app/comments-feed ./cmd/main.go

# Step 3: Final
FROM alpine:3.22

RUN apk add --no-cache curl

COPY --from=builder /app/config /config
COPY --from=builder /app/comments-feed /app/comments-feed
COPY --from=builder /app/internal/database/migrations /app/database/migrations

WORKDIR /app
EXPOSE 8080
CMD ["/app/comments-feed"]
