package http

import "testing"

func TestCanonicalAuthSourceNormalizesEquivalentIPv6Forms(t *testing.T) {
	first := canonicalAuthSource("2001:0db8:0:0:0:0:0:1")
	second := canonicalAuthSource("2001:db8::1")
	if first != second || first != "2001:db8::1" {
		t.Fatalf("equivalent IPv6 sources were not canonicalized: %q != %q", first, second)
	}
}
