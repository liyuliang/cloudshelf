package rws

import (
	"fmt"
	"net/url"
	"strings"
)

type clientOptions struct {
	host       string
	domain     string
	virtualDir string
}

func (o clientOptions) baseUrl() string {
	if strings.HasPrefix(strings.ToLower(o.domain), "http") {
		url, _ := url.JoinPath(o.domain, o.virtualDir)
		return url
	}
	host := strings.ToLower(o.host)
	if !strings.Contains(host, "%s") {
		url, _ := url.JoinPath(o.host, o.virtualDir)
		return url
	}

	url, _ := url.JoinPath(fmt.Sprintf(host, o.domain), o.virtualDir)
	return url
}

func newClientOptions() *clientOptions {
	return &clientOptions{
		domain:     "innovate",
		host:       "https://%s.mdsol.com",
		virtualDir: "RaveWebServices",
	}
}

type ClientOption func(o *clientOptions)

func WithInnovate() ClientOption {
	return func(o *clientOptions) {
		o.domain = "innovate"
	}
}
func WithRave() ClientOption {
	return func(o *clientOptions) {
		o.domain = "rave"
	}
}

func WithDomain(domain string) ClientOption {
	return func(o *clientOptions) {
		o.domain = domain
	}
}

func WithHost(host string) ClientOption {
	return func(o *clientOptions) {
		o.host = host
	}
}

func WithVirtualDir(virtualDir string) ClientOption {
	return func(o *clientOptions) {
		o.virtualDir = virtualDir
	}
}
