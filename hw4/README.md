run with `go run ./...` inside this folder, you can specify the config file through the -rules option (or just edit the default one)

config props:
* default_action - "drop" or "accept"

rule props:
* action - "drop" or "accept"

5 rule conditions:
* transfort - "udp" or "tcp" or "any"
* direction - "query" or "response" or "any"
* name - array of hostnames, without trailing dot
* opcode - array of action codes, as defined in DNS standard
* rcode - array of result codes, as defined in DNS standard

for specific json format, see `rules.json`

the program will output "packet blocked" or "packet accepted" when blocking/passing a dns packet.

for easy checking within GNS3, make build command pushes the code into a Docker image `eeestrelok/gns3-golang:latest`