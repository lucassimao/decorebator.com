package common

import "testing"

func TestNormalizeEmailUsesStableASCIIIdentity(t *testing.T) {
	canonical, err := NormalizeEmail(" \tMixed.Case+tag@Example.COM\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if canonical != "mixed.case+tag@example.com" {
		t.Fatalf("canonical email = %q", canonical)
	}

	for _, invalid := range []string{
		"",
		"not-an-email",
		"Name <user@example.com>",
		"usér@example.com",
		"K@example.com",
		"user@例.example",
		"user..name@example.com",
		"user@example",
		"user@-example.com",
		"\"user name\"@example.com",
	} {
		if got, err := NormalizeEmail(invalid); err == nil || got != "" {
			t.Fatalf("NormalizeEmail(%q) = %q, %v; want invalid", invalid, got, err)
		}
	}
}
