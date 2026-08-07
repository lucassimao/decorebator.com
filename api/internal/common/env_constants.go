package common

import "fmt"

const (
	// Environment names
	ProductionEnv  = "production"
	DevelopmentEnv = "development"
	TestEnv        = "test"
	StagingEnv     = "staging"
)

func ValidateRuntimeEnvironment(environment string) error {
	switch environment {
	case DevelopmentEnv, TestEnv, StagingEnv, ProductionEnv:
		return nil
	default:
		return fmt.Errorf("ENV must be one of development, test, staging, or production")
	}
}
