# FROM golang:1.24.2 AS builder
# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o driver-mgmt-svc ./cmd

# FROM gcr.io/distroless/static-debian11
# WORKDIR /app/
# COPY --from=builder /app/driver-mgmt-svc .
# EXPOSE 8000
# CMD ["./driver-mgmt-svc"]
FROM golang:1.24.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o driver-mgmt-svc ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app/

# Copy migrations and binary
COPY --from=builder /app/driver-mgmt-svc .
COPY --from=builder /app/migrations /migrations

# Copy Goose binary
COPY --from=golang:1.24 /go/bin/goose /usr/local/bin/

EXPOSE 8000

CMD ["sh", "-c", "goose -dir /migrations up && ./driver-mgmt-svc"]
