FROM golang:latest AS builder
WORKDIR /go/src/build
ENV GOPATH /go
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o simpleServer .

FROM alpine:3.6
LABEL maintainer="Prabal Deb <prbldeb@gmail.com>"

RUN set -ex \
  && apk --update add --no-cache curl ca-certificates \
  && rm -rf /tmp/* \
  && rm -rf /var/lib/apt/lists/* \
  && rm /var/cache/apk/*
COPY --from=builder /go/src/build/simpleServer /simpleServer/
EXPOSE 8080
ENTRYPOINT [ "/simpleServer/simpleServer" ]
