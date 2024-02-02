FROM golang:1.14.12-alpine3.12 as builder

RUN apk add --no-cache git build-base
WORKDIR /src
COPY . /src
RUN go mod download && \
    make linux-amd64 && \
    mv ./bin/bot-linux-amd64 /bot

FROM alpine:latest

COPY --from=builder /bot /
RUN mkdir /data
ENTRYPOINT ["/bot"]
