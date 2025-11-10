package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	netfilter "github.com/AkihiroSuda/go-netfilter-queue"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
)

const (
	accept string = "accept"
	drop          = "drop"
)

type Rule struct {
	Action    string   `json:"action"`              // accept/drop
	Transport string   `json:"transport,omitempty"` // tcp/udp
	Direction string   `json:"direction,omitempty"` // query/response/any
	Name      []string `json:"name,omitempty"`
	Opcode    []int    `json:"opcode,omitempty"`
	Rcode     []int    `json:"rcode,omitempty"`
}

type Config struct {
	DefaultAction string `json:"default_action"` // accept/drop
	Rules         []Rule `json:"rules"`
}

func main() {
	rulesPath := flag.String("rules", "rules.json", "path to rules.json")
	flag.Parse()

	cfg, err := loadConfig(*rulesPath)
	if err != nil {
		panic(err)
	}

	nfq, err := netfilter.NewNFQueue(5, 4096, netfilter.NF_DEFAULT_PACKET_SIZE)
	if err != nil {
		panic(err)
	}
	defer nfq.Close()

	log.Printf("listening on NFQUEUE 5")
	packets := nfq.GetPackets()
	for p := range packets {
		if checkPacket(cfg, p.Packet) {
			p.SetVerdict(netfilter.NF_ACCEPT)
		} else {
			p.SetVerdict(netfilter.NF_DROP)
		}
	}
}

func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	cfg.DefaultAction = strings.ToLower(cfg.DefaultAction)
	if cfg.DefaultAction == "" {
		cfg.DefaultAction = accept
	}
	return &cfg, nil
}

type parsedDNS struct {
	transport string
	msg       dns.Msg
	name      string
	isQuery   bool
	opcode    int
	rcode     int
}

func checkPacket(cfg *Config, pkt []byte) bool {
	proto, payload, err := parse(pkt)
	if err != nil || proto == "" {
		return cfg.DefaultAction == accept
	}
	p := parseDNS(proto, payload)
	if p == nil { // not dns
		return cfg.DefaultAction == accept
	}
	for _, r := range cfg.Rules {
		if matchRule(r, *p) {
			return strings.ToLower(r.Action) == accept
		}
	}
	return cfg.DefaultAction == accept
}

func parse(packet []byte) (string, []byte, error) {
	if len(packet) < 1 {
		return "", nil, fmt.Errorf("packet too short")
	}
	ver := packet[0] >> 4
	switch ver {
	case 4:
		return parseLayer(layers.LayerTypeIPv4, packet)
	case 6:
		return parseLayer(layers.LayerTypeIPv6, packet)
	default:
		return "", nil, fmt.Errorf("ip not v4 nor v6")
	}
}

func parseLayer(proto gopacket.LayerType, packetData []byte) (string, []byte, error) {
	packet := gopacket.NewPacket(packetData, proto, gopacket.NoCopy)
	if l := packet.Layer(layers.LayerTypeUDP); l != nil {
		return "udp", l.(*layers.UDP).Payload, nil
	}
	if l := packet.Layer(layers.LayerTypeTCP); l != nil {
		return "tcp", l.(*layers.TCP).Payload, nil
	}
	return "", nil, fmt.Errorf("not udp or tcp")
}

func parseDNS(transport string, payload []byte) *parsedDNS {
	if transport == "tcp" {
		if len(payload) < 2 {
			return nil
		}
		l := int(binary.BigEndian.Uint16(payload[:2]))
		if l <= 0 || len(payload) < 2+l {
			return nil
		}
		payload = payload[2 : 2+l]
	}

	var d layers.DNS
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeDNS, &d)
	lType := []gopacket.LayerType{}
	if err := parser.DecodeLayers(payload, &lType); err != nil {
		return nil
	}

	p := parsedDNS{
		transport: transport,
		isQuery:   !d.QR,
		opcode:    int(d.OpCode),
		rcode:     int(d.ResponseCode),
	}

	if len(d.Questions) > 0 {
		p.name = string(d.Questions[0].Name)
	}

	return &p
}

func matchRule(r Rule, p parsedDNS) bool {
	// transport
	if r.Transport != "" && r.Transport != "any" && strings.ToLower(r.Transport) != p.transport {
		return false
	}
	// direction
	if r.Direction != "" && r.Direction != "any" {
		if r.Direction == "query" && !p.isQuery {
			return false
		}
		if r.Direction == "response" && p.isQuery {
			return false
		}
	}
	// name
	if !slices.Contains(r.Name, p.name) {
		return false
	}
	// opcode
	if len(r.Opcode) > 0 && !slices.Contains(r.Opcode, p.opcode) {
		return false
	}
	//rcode
	if len(r.Rcode) > 0 && !slices.Contains(r.Rcode, p.rcode) {
		return false
	}
	return true
}
