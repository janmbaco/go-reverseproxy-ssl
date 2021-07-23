# Go ReverseProxy SSL

**Go ReverseProxy SSL** was born with the objective of establishing a secure and identified communication between a client and a web service. Its main target is that there be always a secure transport (TLS) between the client and server. For that, you need to configure the services you want to give access to.

This proxy will only accept requests over SSL and will encrypt the communication with a Let's Encrypt certificate, the client can provide their own certificate. Additionally, it can be configured to require the client to provide a certificate for a specific service.

As a secondary objective, **Go ReverseProxy SSL** has the mission of identify the client in the web service by the public key that the client has sent. For that, the service must accept the Http Header "X-Forwarded-ClientKey" that contains all the information of the public key.

The proxy can communicate with the service in any way. That is, if necessary, digital certificates can be added to the proxy for communication with the service.

Here you have a diagram of how the communication with the services is through Go ReverseProsy SSL.

![](https://i.ibb.co/Sx30SS9/Go-Reverse-Proxy-Esquema.png)
