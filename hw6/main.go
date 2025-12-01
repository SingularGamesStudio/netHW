//go:build linux

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"slices"
	"strings"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func htons(i uint16) uint16 {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return binary.BigEndian.Uint16(b)
}

func socket(portName string) (int, syscall.Sockaddr, error) {
	port, err := net.InterfaceByName(portName)
	if err != nil {
		return -1, nil, err
	}
	socket, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		return -1, nil, err
	}
	addr := &syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ALL),
		Ifindex:  port.Index,
	}
	err = syscall.Bind(socket, addr)
	return socket, addr, err
}

var lines = []string{
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
	"my.soul.invaded.with.vital.force",
	"wont.spare.what.im.hunting.for",
	"its.the.animal.within.my.blood",
	"wouldnt.stop.it.if.i.could",
	"seed.is.sown.im.chippin.in",
	"roll.the.bones.im.chippin.in",
	"embed.that.code.im.chippin.in",
	"mayhem.flows",
	"not.backing.down",
	"never.backing.down",
	"not.backing.down",
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
	"suits.run.when.i.come.undone",
	"cant.kill.me.im.zero.and.one",
	"add.justice.to.the.peoples.math",
	"blaze.your.way.down.the.rebel.path",
	"hear.my.call.im.chippin.in",
	"total.war.im.chippin.in",
	"casings.fall.im.chippin.in",
	"kill.them.all",
	"not.backing.down",
	"never.backing.down",
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
	"ca-can.you.feel.it",
	"can.you.touch.it",
	"get.ready.cause.here.we.go",
}

func IPByIndex(i int) net.IP {
	return net.IPv4(10, 0, 0, byte(1+i))
}

var sockAddr = map[int]syscall.Sockaddr{}
var sockMAC = map[int]net.HardwareAddr{}

func dnsPtrToIP(name string) (net.IP, bool) {
	name = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(name, ".")))
	if !strings.HasSuffix(name, ".in-addr.arpa") {
		return nil, false
	}

	ipParts := strings.Split(strings.TrimSuffix(name, ".in-addr.arpa"), ".")
	if len(ipParts) != 4 {
		return nil, false
	}
	slices.Reverse(ipParts)
	ip := net.ParseIP(strings.Join(ipParts, "."))
	if ip == nil {
		return nil, false
	}
	return ip.To4(), true
}

func IndexByIP(ip net.IP) (int, bool) {
	ip4 := ip.To4()
	if ip4[0] != 10 || ip4[1] != 0 || ip4[2] != 0 {
		return -1, false
	}
	return int(ip4[3]) - 1, true
}

func Traceroute(sock int, eth *layers.Ethernet, ip *layers.IPv4, udp *layers.UDP) bool {
	if !ip.DstIP.Equal(IPByIndex(len(lines) - 1)) {
		return false
	}
	lineIdx := int(ip.TTL) - 1

	ethLayer := layers.Ethernet{
		SrcMAC:       sockMAC[sock],
		DstMAC:       eth.SrcMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipLayer := layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      128,
		SrcIP:    IPByIndex(lineIdx),
		DstIP:    ip.SrcIP,
		Protocol: layers.IPProtocolICMPv4,
	}
	icmpLayer := layers.ICMPv4{}
	if lineIdx >= len(lines)-1 {
		icmpLayer.TypeCode = layers.CreateICMPv4TypeCode(
			layers.ICMPv4TypeDestinationUnreachable,
			layers.ICMPv4CodePort,
		)
	} else {
		icmpLayer.TypeCode = layers.CreateICMPv4TypeCode(
			layers.ICMPv4TypeTimeExceeded,
			layers.ICMPv4CodeTTLExceeded,
		)
	}

	// First 8 bytes of payload copied
	orig := ip.Contents
	if len(ip.Payload) > 8 {
		orig = append(orig, ip.Payload[:8]...)
	} else {
		orig = append(orig, ip.Payload...)
	}
	payload := gopacket.Payload(orig)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	if err := gopacket.SerializeLayers(buf, opts, &ethLayer, &ipLayer, &icmpLayer, payload); err != nil {
		fmt.Println("ICMP serialize:", err)
		return false
	}
	if err := syscall.Sendto(sock, buf.Bytes(), 0, sockAddr[sock]); err != nil {
		fmt.Println("ICMP send:", err)
	}
	return true
}

func sendDNSResponse(sock int, inEth *layers.Ethernet, inIP *layers.IPv4, inUDP *layers.UDP, resp *layers.DNS) {
	ethLayer := layers.Ethernet{
		SrcMAC:       sockMAC[sock],
		DstMAC:       inEth.SrcMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipLayer := layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    inIP.DstIP,
		DstIP:    inIP.SrcIP,
	}
	udpLayer := layers.UDP{
		SrcPort: inUDP.DstPort,
		DstPort: inUDP.SrcPort,
	}
	udpLayer.SetNetworkLayerForChecksum(&ipLayer)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	if err := gopacket.SerializeLayers(buf, opts, &ethLayer, &ipLayer, &udpLayer, resp); err != nil {
		fmt.Println("DNS serialize:", err)
		return
	}
	if err := syscall.Sendto(sock, buf.Bytes(), 0, sockAddr[sock]); err != nil {
		fmt.Println("DNS send:", err)
	}
}

func DNS(sock int, ethernet *layers.Ethernet, ip *layers.IPv4, udp *layers.UDP, packet gopacket.Packet) bool {
	if udp.DstPort != 53 {
		return false
	}
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	if dnsLayer == nil {
		return false
	}
	req := dnsLayer.(*layers.DNS)
	resp := &layers.DNS{
		ID:           req.ID,
		QR:           true,
		OpCode:       req.OpCode,
		AA:           true,
		RD:           req.RD,
		RA:           true,
		ResponseCode: layers.DNSResponseCodeNoErr,
		Questions:    req.Questions,
	}
	inSong := false

	for _, q := range req.Questions {
		name := string(q.Name)

		switch q.Type {
		case layers.DNSTypeA:
			if strings.EqualFold(strings.TrimSuffix(strings.ToLower(name), "."), "rerand0m.ru") {
				resp.Answers = append(resp.Answers, layers.DNSResourceRecord{
					Name:  q.Name,
					Type:  layers.DNSTypeA,
					Class: layers.DNSClassIN,
					TTL:   128,
					IP:    IPByIndex(len(lines) - 1),
				})
				inSong = true
			}

		case layers.DNSTypePTR:
			ipQ, ok1 := dnsPtrToIP(name)
			if !ok1 {
				break
			}
			idx, ok2 := IndexByIP(ipQ)
			if !ok2 {
				break
			}
			resp.Answers = append(resp.Answers, layers.DNSResourceRecord{
				Name:  q.Name,
				Type:  layers.DNSTypePTR,
				Class: layers.DNSClassIN,
				TTL:   128,
				PTR:   []byte(lines[idx] + "."),
			})
			inSong = true
		}
	}

	if !inSong {
		return false
	}
	resp.ANCount = uint16(len(resp.Answers))
	sendDNSResponse(sock, ethernet, ip, udp, resp)
	return true
}

func transparent(from, to int, toMAC net.HardwareAddr) {
	fmt.Printf("Started transfer %d->%d\n", from, to)
	buffer := make([]byte, 100000)
	for {
		cnt, _, err := syscall.Recvfrom(from, buffer, 0)
		if err != nil {
			panic(err)
		}
		if cnt > 0 {
			packet := gopacket.NewPacket(buffer[:cnt], layers.LinkTypeEthernet, gopacket.Default)
			ethernet := packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
			if bytes.Equal(ethernet.SrcMAC, toMAC) {
				continue // skip your own packets
			}
			if packet.Layer(layers.LayerTypeIPv4) != nil {
				ip := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
				if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
					udp := udpLayer.(*layers.UDP)
					if DNS(from, ethernet, ip, udp, packet) {
						continue
					}
					if Traceroute(from, ethernet, ip, udp) {
						continue
					}
				}
			}
			syscall.Sendto(to, buffer[:cnt], 0, sockAddr[to])
		}
	}
}

func main() {
	dev0, _ := net.InterfaceByName("eth0")
	dev2, _ := net.InterfaceByName("eth2")
	sock0, addr0, err := socket("eth0")
	if err != nil {
		panic(err)
	}
	sock1, addr1, err := socket("eth2")
	if err != nil {
		panic(err)
	}

	sockAddr[sock0] = addr0
	sockAddr[sock1] = addr1
	sockMAC[sock0] = dev0.HardwareAddr
	sockMAC[sock1] = dev2.HardwareAddr

	go transparent(sock0, sock1, dev2.HardwareAddr)
	go transparent(sock1, sock0, dev0.HardwareAddr)
	c := make(chan int)
	<-c
}
