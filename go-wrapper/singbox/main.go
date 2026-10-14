package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/adapter/certificate"
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/adapter/service"
	"github.com/sagernet/sing-box/dns"
	"github.com/sagernet/sing-box/dns/transport"
	"github.com/sagernet/sing-box/dns/transport/fakeip"
	"github.com/sagernet/sing-box/dns/transport/hosts"
	"github.com/sagernet/sing-box/dns/transport/local"
	"github.com/sagernet/sing-box/dns/transport/mdns"
	"github.com/sagernet/sing-box/dns/transport/quic"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/anytls"
	"github.com/sagernet/sing-box/protocol/block"
	"github.com/sagernet/sing-box/protocol/direct"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing-box/protocol/hysteria"
	"github.com/sagernet/sing-box/protocol/hysteria2"
	"github.com/sagernet/sing-box/protocol/http"
	"github.com/sagernet/sing-box/protocol/mixed"
	"github.com/sagernet/sing-box/protocol/shadowsocks"
	"github.com/sagernet/sing-box/protocol/socks"
	"github.com/sagernet/sing-box/protocol/trojan"
	"github.com/sagernet/sing-box/protocol/tuic"
	"github.com/sagernet/sing-box/protocol/vless"
	"github.com/sagernet/sing-box/protocol/vmess"
	"github.com/sagernet/sing-box/protocol/wireguard"
	"github.com/sagernet/sing-box/service/api"

	_ "github.com/sagernet/sing-box/experimental/clashapi"
)

var (
	mu       sync.Mutex
	instance *box.Box
)

func miniContext(ctx context.Context) context.Context {
	inReg := inbound.NewRegistry()
	direct.RegisterInbound(inReg)
	socks.RegisterInbound(inReg)
	http.RegisterInbound(inReg)
	mixed.RegisterInbound(inReg)
	shadowsocks.RegisterInbound(inReg)
	vmess.RegisterInbound(inReg)
	trojan.RegisterInbound(inReg)
	vless.RegisterInbound(inReg)
	anytls.RegisterInbound(inReg)
	hysteria.RegisterInbound(inReg)
	tuic.RegisterInbound(inReg)
	hysteria2.RegisterInbound(inReg)

	outReg := outbound.NewRegistry()
	direct.RegisterOutbound(outReg)
	block.RegisterOutbound(outReg)
	group.RegisterSelector(outReg)
	group.RegisterURLTest(outReg)
	socks.RegisterOutbound(outReg)
	http.RegisterOutbound(outReg)
	shadowsocks.RegisterOutbound(outReg)
	vmess.RegisterOutbound(outReg)
	trojan.RegisterOutbound(outReg)
	vless.RegisterOutbound(outReg)
	anytls.RegisterOutbound(outReg)
	hysteria.RegisterOutbound(outReg)
	tuic.RegisterOutbound(outReg)
	hysteria2.RegisterOutbound(outReg)

	epReg := endpoint.NewRegistry()
	wireguard.RegisterEndpoint(epReg)

	dnsReg := dns.NewTransportRegistry()
	transport.RegisterTCP(dnsReg)
	transport.RegisterUDP(dnsReg)
	transport.RegisterTLS(dnsReg)
	transport.RegisterHTTPS(dnsReg)
	hosts.RegisterTransport(dnsReg)
	local.RegisterTransport(dnsReg)
	mdns.RegisterTransport(dnsReg)
	fakeip.RegisterTransport(dnsReg)
	quic.RegisterTransport(dnsReg)
	quic.RegisterHTTP3Transport(dnsReg)

	svcReg := service.NewRegistry()
	api.RegisterService(svcReg)

	certReg := certificate.NewRegistry()

	return box.Context(ctx, inReg, outReg, epReg, dnsReg, svcReg, certReg)
}

//export StartSingBox
func StartSingBox(payload *C.char) C.int {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return -1
	}

	var p struct {
		Config      string `json:"config"`
		WorkingDir  string `json:"workingDir"`
		DisableColor bool  `json:"disableColor"`
	}
	if err := json.Unmarshal([]byte(C.GoString(payload)), &p); err != nil {
		return -2
	}

	if p.WorkingDir != "" {
		os.Chdir(p.WorkingDir)
	}

	configData, err := os.ReadFile(p.Config)
	if err != nil {
		return -2
	}

	var opts option.Options
	if err := json.Unmarshal(configData, &opts); err != nil {
		return -2
	}

	ctx := miniContext(context.Background())
	inst, err := box.New(box.Options{
		Context: ctx,
		Options: opts,
	})
	if err != nil {
		return -3
	}

	if err := inst.Start(); err != nil {
		return -4
	}

	instance = inst
	return 0
}

//export StopSingBox
func StopSingBox() C.int {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		return -1
	}

	instance.Close()
	instance = nil
	return 0
}

func main() {}