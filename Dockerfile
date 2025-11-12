FROM golang:1.24.2 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o driver-mgmt-svc ./cmd

FROM gcr.io/distroless/static-debian11
WORKDIR /app/
COPY --from=builder /app/driver-mgmt-svc .
EXPOSE 8000
CMD ["./driver-mgmt-svc"]
