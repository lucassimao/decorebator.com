package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type TrustedProxyConfig struct {
	CIDRs      []string
	DirectMode bool
}

func LoadTrustedProxyConfig() (TrustedProxyConfig, error) {
	return LoadTrustedProxyConfigFrom(os.LookupEnv)
}

func LoadTrustedProxyConfigFrom(lookup EnvironmentLookup) (TrustedProxyConfig, error) {
	if lookup == nil {
		return TrustedProxyConfig{}, fmt.Errorf("trusted proxy environment lookup is required")
	}
	environment, _ := lookup("ENV")
	directMode := false
	if rawDirect, ok := lookup("DIRECT_CLIENT_TRAFFIC"); ok && strings.TrimSpace(rawDirect) != "" {
		if strings.EqualFold(strings.TrimSpace(rawDirect), "true") {
			directMode = true
		} else if !strings.EqualFold(strings.TrimSpace(rawDirect), "false") {
			return TrustedProxyConfig{}, fmt.Errorf("DIRECT_CLIENT_TRAFFIC must be true or false")
		}
	}
	raw, ok := lookup("TRUSTED_PROXY_CIDRS")
	if !ok || strings.TrimSpace(raw) == "" {
		if environment == "production" && !directMode {
			return TrustedProxyConfig{}, fmt.Errorf("production requires TRUSTED_PROXY_CIDRS or DIRECT_CLIENT_TRAFFIC=true")
		}
		return TrustedProxyConfig{DirectMode: directMode}, nil
	}
	if directMode {
		return TrustedProxyConfig{}, fmt.Errorf("DIRECT_CLIENT_TRAFFIC and TRUSTED_PROXY_CIDRS are mutually exclusive")
	}

	seen := make(map[string]struct{})
	config := TrustedProxyConfig{}
	for _, entry := range strings.Split(raw, ",") {
		proxy := strings.TrimSpace(entry)
		if proxy == "" {
			return TrustedProxyConfig{}, fmt.Errorf("TRUSTED_PROXY_CIDRS contains an empty entry")
		}
		if net.ParseIP(proxy) == nil {
			_, network, err := net.ParseCIDR(proxy)
			if err != nil {
				return TrustedProxyConfig{}, fmt.Errorf("TRUSTED_PROXY_CIDRS contains invalid address %q", proxy)
			}
			prefix, _ := network.Mask.Size()
			if prefix == 0 {
				return TrustedProxyConfig{}, fmt.Errorf("TRUSTED_PROXY_CIDRS must not trust every address")
			}
		}
		if _, exists := seen[proxy]; exists {
			continue
		}
		seen[proxy] = struct{}{}
		config.CIDRs = append(config.CIDRs, proxy)
	}
	return config, nil
}
