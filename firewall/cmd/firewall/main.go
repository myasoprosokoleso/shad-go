//go:build !solution

package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"gitlab.com/slon/shad-go/firewall/internal/rules"
	"gitlab.com/slon/shad-go/firewall/internal/transport"
)

func main() {
	serviceAddr := flag.String("service-addr", "http://localhost:8080", "address of the service to protect")
	firewallAddr := flag.String("addr", "localhost:8081", "address of the firewall to listen")
	configPath := flag.String("conf", "./firewall/configs/example.yaml", "path to the firewall yaml configuration file")
	flag.Parse()

	runFirewall(*serviceAddr, *firewallAddr, *configPath)
}

func runFirewall(serviceAddr, firewallAddr, configPath string) {
	upstream, err := url.Parse(serviceAddr)
	if err != nil {
		log.Fatalf("invalid service address: %v", err)
	}

	rules, err := rules.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load rules: %v", err)
	}

	t := transport.NewTransport(rules)
	rp := httputil.NewSingleHostReverseProxy(upstream)
	rp.Transport = t

	log.Printf("Listening on %s (firewall) → %s (service)", firewallAddr, serviceAddr)
	log.Fatal(http.ListenAndServe(firewallAddr, rp))
}
