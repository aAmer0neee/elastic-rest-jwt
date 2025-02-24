FROM golang:latest

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod tidy
RUN go build -o app ./cmd/main.go
RUN ls -la .

RUN chmod +x ./app

CMD ["./app"]