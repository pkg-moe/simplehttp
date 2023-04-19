package simplehttp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

var (
	dnsClient = new(dns.Client)
)

// SetCustomDNS set default transport custom dns
func SetCustomDNS(dnsAddr string) {
	// empty dns addr to set default transport
	if dnsAddr == "" {
		dialContext = (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext

		defaultTransport = &http.Transport{
			DialContext:       dialContext,
			DisableKeepAlives: true,
		}

		return
	}

	dialContext = (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, network, dnsAddr)
		}},
	}).DialContext

	defaultTransport = &http.Transport{
		DialContext:       dialContext,
		DisableKeepAlives: true,
	}
}

// ResolveIP resolve domain ip list by dns addr
func ResolveIP(dnsAddr, domain string) ([]string, error) {
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	if dnsAddr == "" {
		return net.DefaultResolver.LookupHost(ctx, domain)
	}

	addrs := []string{}
	parsed := net.ParseIP(domain)
	if parsed != nil {
		addrs = append(addrs, domain)
		return addrs, nil
	}

	messageA := new(dns.Msg)
	messageA.SetQuestion(dns.Fqdn(domain), dns.TypeA)

	inA, _, err := dnsClient.ExchangeContext(ctx, messageA, dnsAddr)

	if err != nil {
		return nil, err
	}

	for _, record := range inA.Answer {
		if t, ok := record.(*dns.A); ok {
			addrs = append(addrs, t.A.String())
		}
	}

	return addrs, nil
}

// NewClientWithDNS create a http client with custom dns
func NewClientWithDNS(dnsAddr, domain string) *http.Client {
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		addrs, err := ResolveIP(dnsAddr, domain)
		if err != nil {
			return nil, err
		}

		return dialer.DialContext(ctx, network, net.JoinHostPort(addrs[0], "80"))
	}

	httpTransport := &http.Transport{
		DialContext:       dialContext,
		DisableKeepAlives: true,
	}

	return &http.Client{Transport: httpTransport, Timeout: 5 * time.Second}
}
