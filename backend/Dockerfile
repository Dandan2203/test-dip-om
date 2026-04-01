FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /finagent ./cmd && \
    CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /finagent ./finagent
COPY --from=builder /migrate ./migrate
COPY migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["sh", "-c", "./migrate up && ./finagent"]
