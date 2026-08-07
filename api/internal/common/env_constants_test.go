package common

import "testing"

func TestValidateRuntimeEnvironment(t *testing.T) {
	for _, environment := range []string{DevelopmentEnv, TestEnv, StagingEnv, ProductionEnv} {
		if err := ValidateRuntimeEnvironment(environment); err != nil {
			t.Fatalf("expected %q to be valid: %v", environment, err)
		}
	}
	for _, environment := range []string{"", "prod", "Production", "unknown"} {
		if err := ValidateRuntimeEnvironment(environment); err == nil {
			t.Fatalf("expected %q to be rejected", environment)
		}
	}
}
