# Stage 1: Build the Go app
FROM golang:1.22.2-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go

# Stage 2: Run the app in a smaller image
FROM alpine:latest

COPY --from=builder /app/app /usr/local/bin/app

EXPOSE 8888

CMD ["app"]