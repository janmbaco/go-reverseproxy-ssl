# Go ReverseProxy SSL
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/janmbaco/go-reverseproxy-ssl)

**Go ReverseProxy SSL** is a tool that aims to establish a secure and authenticated communication between a client and a web service over HTTP/2. It ensures that the transport layer (TLS) is always encrypted between the client and the server. To achieve this, you need to configure the services that you want to expose through the proxy.

This proxy only accepts requests over SSL and encrypts the communication with a valid certificate provided by Let's Encrypt. The client can also provide their own certificate if needed. Moreover, the proxy can be configured to require the client to present a certificate for a specific service.

Another goal of Go ReverseProxy SSL is to identify the client in the web service by the public key that the client has sent. For this, the service must accept the HTTP header "X-Forwarded-ClientKey" that contains all the information of the public key.

The proxy can communicate with the service in any way. That means that you can add digital certificates to the proxy for communication with the service if necessary.

Go ReverseProxy SSL can be used for different communication protocols that run on top of HTTP/2, such as REST API, SSH, gRPC and gRPC Web. Here is a diagram of how the communication with the services works through Go ReverseProxy SSL.

![](https://i.ibb.co/Sx30SS9/Go-Reverse-Proxy-Esquema.png)
