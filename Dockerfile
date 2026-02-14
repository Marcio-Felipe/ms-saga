FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/${SERVICE_NAME}

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/service /service
EXPOSE 8080
ENTRYPOINT ["/service"]
