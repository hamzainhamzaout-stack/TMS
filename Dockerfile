FROM golang:1.24.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o driver-mgmt-svc ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app/
COPY --from=builder /app/driver-mgmt-svc .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/docs/swagger-ui ./docs/swagger-ui
EXPOSE 8000
CMD ["./driver-mgmt-svc"]
