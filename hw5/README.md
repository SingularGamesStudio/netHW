# run:

server:

`go run cmd/main.go -ip 0.0.0.0:9999 -role server/client -cert file -key file`

client: (cert is the trusted server certificate, optional)

`go run cmd/main.go -ip 0.0.0.0:9999 -role server/client -cert file`

# config:
certificate must be named "localhost", example generation:

`openssl req -nodes -x509 -sha256 -newkey rsa:2048 -keyout server.key -out server.crt -subj "/CN=localhost" -addext "subjectAltName = DNS:localhost"`