package common

import "os"

type Environment int

const (
	Production Environment = iota
	Development
)

var Config struct {
	Env                              Environment
	OpenaiImageGenerationApiEndpoint string
	OpenaiChatCompletionApiEndpoint  string
	OpenaiApiKey                     string
	Port                             string
	PostgresUser                     string
	PostgresPassword                 string
	PostgresDB                       string
	PostgresPort                     string
	PostgresHost                     string
	GinMode                          string
	JwtKey                           string
	RedisAddr                        string
	MinioHost                        string
	MinioPort                        string
	MinioRootUser                    string
	MinioRootPassword                string
}

func init() {
	if os.Getenv("ENV") == "production" {
		Config.Env = Production
	} else {
		Config.Env = Development
	}

	Config.OpenaiImageGenerationApiEndpoint = os.Getenv("OPENAI_IMAGE_GENERATION_API_ENDPOINT")
	Config.OpenaiChatCompletionApiEndpoint = os.Getenv("OPENAI_CHAT_COMPLETION_API_ENDPOINT")
	Config.OpenaiApiKey = os.Getenv("OPENAI_API_KEY")
	Config.Port = os.Getenv("PORT")
	Config.PostgresUser = os.Getenv("POSTGRES_USER")
	Config.PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	Config.PostgresDB = os.Getenv("POSTGRES_DB")
	Config.PostgresPort = os.Getenv("POSTGRES_PORT")
	Config.PostgresHost = os.Getenv("POSTGRES_HOST")
	Config.GinMode = os.Getenv("GIN_MODE")
	Config.JwtKey = os.Getenv("JWT_KEY")
	Config.RedisAddr = os.Getenv("REDIS_ADDR")
	Config.MinioHost = os.Getenv("MINIO_HOST")
	Config.MinioPort = os.Getenv("MINIO_PORT")
	Config.MinioRootUser = os.Getenv("MINIO_ROOT_USER")
	Config.MinioRootPassword = os.Getenv("MINIO_ROOT_PASSWORD")
}
