FROM alpine

WORKDIR /app

RUN pwd

RUN ls -al

COPY go-reverseproxy-ssl .

EXPOSE 80
EXPOSE 443