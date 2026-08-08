package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const minimumStaticAuthenticationBytes = 32

const httpsScheme = "https"
const legacyProductionCookieDomain = "decorebator.com"
const productionEnvironment = "production"
const stagingEnvironment = "staging"

var localAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:4000",
	"http://localhost:8081",
	"http://127.0.0.1:3000",
	"http://127.0.0.1:4000",
	"http://127.0.0.1:8081",
}

type HTTPSecurityConfig struct {
	AllowedOrigins       []string
	StaticAuthentication string
	SecureCookies        bool
	CookieDomain         string
	LegacyCookieDomain   string
}

func LoadHTTPSecurityConfig(environment string) (HTTPSecurityConfig, error) {
	return LoadHTTPSecurityConfigFrom(environment, os.LookupEnv)
}

func LoadHTTPSecurityConfigFrom(environment string, lookup EnvironmentLookup) (HTTPSecurityConfig, error) {
	if lookup == nil {
		return HTTPSecurityConfig{}, fmt.Errorf("HTTP security environment lookup is required")
	}
	deployed := environment == productionEnvironment || environment == stagingEnvironment
	production := environment == productionEnvironment
	rawOrigins, _ := lookup("CORS_ALLOWED_ORIGINS")
	rawOrigins = strings.TrimSpace(rawOrigins)
	if rawOrigins == "" {
		if deployed {
			return HTTPSecurityConfig{}, fmt.Errorf("CORS_ALLOWED_ORIGINS is required in production")
		}
		rawOrigins = strings.Join(localAllowedOrigins, ",")
	}
	origins, err := parseAllowedOrigins(rawOrigins, deployed)
	if err != nil {
		return HTTPSecurityConfig{}, err
	}
	if deployed && len(origins) != 1 {
		return HTTPSecurityConfig{}, fmt.Errorf("deployed credentialed browser sessions require exactly one CORS_ALLOWED_ORIGINS origin")
	}

	staticAuthentication, _ := lookup("STATIC_AUTHENTICATION")
	if strings.TrimSpace(staticAuthentication) != staticAuthentication {
		return HTTPSecurityConfig{}, fmt.Errorf("STATIC_AUTHENTICATION must not contain surrounding whitespace")
	}
	if deployed {
		if staticAuthentication == "" {
			return HTTPSecurityConfig{}, fmt.Errorf("STATIC_AUTHENTICATION is required while production static routes are enabled")
		}
		if len([]byte(staticAuthentication)) < minimumStaticAuthenticationBytes {
			return HTTPSecurityConfig{}, fmt.Errorf("STATIC_AUTHENTICATION must contain at least %d bytes", minimumStaticAuthenticationBytes)
		}
		lowerToken := strings.ToLower(staticAuthentication)
		for _, marker := range []string{
			"your_static", "change_me", "changeme", "placeholder", "replace-with", "replace_with", "secret_here",
		} {
			if strings.Contains(lowerToken, marker) {
				return HTTPSecurityConfig{}, fmt.Errorf("STATIC_AUTHENTICATION must not use placeholder credentials")
			}
		}
	}

	legacyCookieDomain := ""
	if production {
		legacyCookieDomain = legacyProductionCookieDomain
	}
	return HTTPSecurityConfig{
		AllowedOrigins: append([]string(nil), origins...), StaticAuthentication: staticAuthentication,
		SecureCookies: deployed, CookieDomain: "", LegacyCookieDomain: legacyCookieDomain,
	}, nil
}

func parseAllowedOrigins(raw string, production bool) ([]string, error) {
	seen := make(map[string]struct{})
	origins := make([]string, 0)
	for _, entry := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(entry)
		if origin == "" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS contains an empty origin")
		}
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Host == "" || parsed.Hostname() == "" ||
			parsed.Scheme+"://"+parsed.Host != origin ||
			(parsed.Scheme != "http" && parsed.Scheme != httpsScheme) ||
			parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS contains invalid origin %q", origin)
		}
		if production && parsed.Scheme != httpsScheme {
			return nil, fmt.Errorf("production CORS origins must use HTTPS")
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins, nil
}
