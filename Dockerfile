FROM golang:1.14-alpine

WORKDIR /app

COPY go.mod ./
COPY configs/*.go ./configs/
COPY configs/certs/*.go ./configs/certs/
COPY grpcUtil/*.go ./grpcUtil/
COPY hosts/*.go ./hosts/
COPY sshUtil/*.go ./sshUtil/
COPY *.go ./

RUN go get

RUN go build -o /go-reverseproxy-ssl

EXPOSE 80

