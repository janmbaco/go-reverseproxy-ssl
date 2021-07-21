FROM alpine

WORKDIR /app

COPY /home/runner/work/go-reverseproxy-ssl/go-reverseproxy-ssl/build/go-reverseproxy-ssl .

EXPOSE 80
EXPOSE 443
