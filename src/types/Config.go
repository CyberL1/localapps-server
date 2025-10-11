package types

type ServerConfig struct {
	AccessUrl string `default:"\"http://localhost:8080\""`
	ApiKey    string `default:"\"\""`
}
