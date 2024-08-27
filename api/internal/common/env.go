package common

import (
	"os"
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
	DatabaseUrl          string // if set has precendence over  PostgresUser, PostgresPassword,PostgresDB,PostgresPort, and PostgresHost
	PostgresUser         string
	PostgresPassword     string
	PostgresDB           string
	PostgresPort         string
	PostgresHost         string
	GinMode              string
	JwtKey               string
	RedisAddr            string
	MinioHost            string
	MinioPort            string
	MinioRootUser        string
	MinioRootPassword    string
	StaticAuthentication string

	Aws struct {
		AccessKeyId     string
		SecretAccessKey string
		Region          string
	}

	StabilityAIApiKey string
}

func init() {
	if os.Getenv("ENV") == "production" {
		Env.Env = Production
	} else {
		Env.Env = Development
	}

	Env.OpenaiApiKey = os.Getenv("OPENAI_API_KEY")
	Env.Port = os.Getenv("PORT")
	Env.DatabaseUrl = os.Getenv("DATABASE_URL")
	Env.PostgresUser = os.Getenv("POSTGRES_USER")
	Env.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	Env.PostgresDB = os.Getenv("POSTGRES_DB")
	Env.PostgresPort = os.Getenv("POSTGRES_PORT")
	Env.PostgresHost = os.Getenv("POSTGRES_HOST")
	Env.GinMode = os.Getenv("GIN_MODE")
	Env.JwtKey = os.Getenv("JWT_KEY")
	Env.RedisAddr = os.Getenv("REDIS_ADDR")
	Env.MinioHost = os.Getenv("MINIO_HOST")
	Env.MinioPort = os.Getenv("MINIO_PORT")
	Env.MinioRootUser = os.Getenv("MINIO_ROOT_USER")
	Env.MinioRootPassword = os.Getenv("MINIO_ROOT_PASSWORD")
	Env.StaticAuthentication = os.Getenv("STATIC_AUTHENTICATION")

	Env.Aws.AccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	Env.Aws.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	Env.Aws.Region = os.Getenv("AWS_REGION")

	Env.StabilityAIApiKey = os.Getenv("STABILITY_AI_API_KEY")
}
