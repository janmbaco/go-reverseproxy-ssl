FROM ubuntu:rolling

WORKDIR /app

COPY build/go-reverseproxy-ssl .

EXPOSE 80
EXPOSE 443
