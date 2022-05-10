package config

type EnvVar struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

type Config struct {
	serverAddress   string
	baseURL         string
	fileStoragePath string
}

//getters

func (c Config) SrvAddr() string {
	return c.serverAddress
}

func (c Config) HostName() string {
	return c.baseURL
}
func (c Config) FilePath() string {
	return c.fileStoragePath
}

//constructor

func NewConfig(srvAddr, hostName string, filePath string) *Config {
	return &Config{
		serverAddress:   srvAddr,
		baseURL:         hostName,
		fileStoragePath: filePath,
	}
}
