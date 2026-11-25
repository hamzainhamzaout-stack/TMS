FROM golang:1.24.2 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o driver-mgmt-svc ./cmd

FROM alpine:latest

RUN apk add --no-cache ca-certificates curl bash

RUN curl -L https://github.com/pressly/goose/releases/download/v3.25.0/goose_linux_x86_64 \
    -o /usr/local/bin/goose && chmod +x /usr/local/bin/goose

WORKDIR /app

COPY --from=builder /app/driver-mgmt-svc .

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/docs ./docs

EXPOSE 8091

CMD ["sh", "-c", "goose -dir ./migrations up && ./driver-mgmt-svc"]


# FROM golang:1.24.2 AS builder
# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o driver-mgmt-svc ./cmd

# FROM gcr.io/distroless/static-debian11
# WORKDIR /app/
# COPY --from=builder /app/driver-mgmt-svc .
# COPY --from=builder /app/docs ./docs
# COPY --from=builder /app/docs/swagger-ui ./docs/swagger-ui
# EXPOSE 8000
# CMD ["./driver-mgmt-svc"]
