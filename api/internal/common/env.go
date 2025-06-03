package common

import (
	"os"

	"github.com/joho/godotenv"
)

type Environment int

const (
	Production Environment = iota
	Development
)

var Env struct {
	Env                  Environment
	OpenaiApiKey         string
	Port                 string
	DatabaseUrl          string
	GinMode              string
	JwtKey               string
	MinioHost            string
	MinioPort            string
	MinioRootUser        string
	MinioRootPassword    string
	StaticAuthentication string
	SendGridApiKey       string

	ResetPasswordPrivateKey string
}

func init() {
	_ = godotenv.Load()

	if os.Getenv("ENV") == "production" {
		Env.Env = Production
	} else {
		Env.Env = Development
	}

	Env.OpenaiApiKey = os.Getenv("OPENAI_API_KEY")
	Env.Port = os.Getenv("PORT")
	Env.DatabaseUrl = os.Getenv("DATABASE_URL")
	Env.GinMode = os.Getenv("GIN_MODE")
	Env.JwtKey = os.Getenv("JWT_KEY")
	Env.MinioHost = os.Getenv("MINIO_HOST")
	Env.MinioPort = os.Getenv("MINIO_PORT")
	Env.MinioRootUser = os.Getenv("MINIO_ROOT_USER")
	Env.MinioRootPassword = os.Getenv("MINIO_ROOT_PASSWORD")
	Env.StaticAuthentication = os.Getenv("STATIC_AUTHENTICATION")

	Env.SendGridApiKey = os.Getenv("SENDGRID_API_KEY")
	Env.ResetPasswordPrivateKey = os.Getenv("RESET_PASSWORD_PRIVATE_KEY")
}
