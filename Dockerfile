FROM golang:alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
WORKDIR /usr/src/app
COPY . .

ENV GO111MODULE=on

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-s -w" -o bin/main ./main.go


### Executable Image
FROM alpine

COPY config.yml .

COPY --from=builder /usr/src/app/bin/main ./main

EXPOSE 8080

ENTRYPOINT ["./main"]