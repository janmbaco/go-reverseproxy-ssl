FROM alpine

RUN apk update && apk upgrade openssl

WORKDIR /app

COPY build/go-reverseproxy-ssl .

EXPOSE 80
EXPOSE 443
