package common

import "os"

type Environment int

const (
	Production Environment = iota
	Development
)

var Config struct {
	env Environment
}

func init() {
	if os.Getenv("ENV") == "production" {
		Config.env = Production
	} else {
		Config.env = Development
	}
}
