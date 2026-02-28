FROM alpine:latest
RUN apk update
RUN apk add lldpd ethtool iptables iproute2 go bridge \
    python3 scapy libpcap libpcap-dev \
    libpcap-doc curl mandoc man-pages tcpdump vim emacs nano  \
    bridge-utils bridge-utils-doc gcc g++ musl-dev openssl openssl-dev \
    openssl-doc linux-headers py3-pip libnetfilter_queue-dev python3-dev \
    openssh nftables git bind-tools
RUN go mod init test
RUN go get github.com/google/gopacket
RUN go get github.com/AkihiroSuda/go-netfilter-queue
RUN rm go.mod
RUN rm go.sum
COPY . .
CMD /bin/ash