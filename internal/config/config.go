package config

type Config struct {
	serverAddress   string
	baseURL         string
	fileStoragePath string
	DatabasePath    string
}

const Secret = "vfktymrjqtkjxrt[jkjlyjpb"

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

func (c Config) Database() string {
	return c.DatabasePath
}

//constructor

func NewConfig(srvAddr, hostName string, filePath string, database string) *Config {
	return &Config{
		serverAddress:   srvAddr,
		baseURL:         hostName,
		fileStoragePath: filePath,
		DatabasePath:    database,
	}
}
