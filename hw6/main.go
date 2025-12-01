//go:build linux

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
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

func transparent(from, to int, toAddr syscall.Sockaddr, toMAC net.HardwareAddr) {
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
			fromIPv4 := ""
			toIPv4 := ""
			if packet.Layer(layers.LayerTypeIPv4) != nil {
				ip := packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
				fromIPv4 = "(" + ip.SrcIP.String() + ")"
				toIPv4 = "(" + ip.DstIP.String() + ")"
			}
			fmt.Printf("%s %s -> %s %s\n",
				ethernet.SrcMAC, fromIPv4,
				toIPv4, ethernet.DstMAC,
			)
			if fromIPv4 == "(8.8.8.8)" || toIPv4 == "(8.8.8.8)" {
				fmt.Println("Packet blocked")
				continue // skip packets to/from google
			}
			syscall.Sendto(to, buffer[:cnt], 0, toAddr)
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
	go transparent(sock0, sock1, addr1, dev2.HardwareAddr)
	go transparent(sock1, sock0, addr0, dev0.HardwareAddr)
	c := make(chan int)
	<-c
}
