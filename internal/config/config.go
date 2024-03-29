//	Package config - defining and getting application launch settings
package config

import (
	"encoding/json"
	"log"
	"os"
)

//	Config struct  - Structure of application settings fields
type Config struct {
	// server address
	ServerAddress string
	// server base URL
	BaseURL string
	// file storage path
	FileStoragePath string
	// database path
	DatabasePath string
	// use ssl
	EnableHTTPS bool
	// config file
	Config string
}

//	The secret word for creating a session id
const Secret = "vfktymrjqtkjxrt[jkjlyjpb"

//	Getters
//	SrvAddr is function to get server address.
func (c Config) SrvAddr() string {
	return c.ServerAddress
}

//	HostName is function to get server hostname.
func (c Config) HostName() string {
	return c.BaseURL
}

//	FilePath function to get file path.
func (c Config) FilePath() string {
	return c.FileStoragePath
}

//	Database function to get database address.
func (c Config) Database() string {
	return c.DatabasePath
}

//	Config constructor
//
//	NewConfig is function to set Application Settings values.
func NewConfig(srvAddr, hostName string, filePath string, database string, ssl bool) *Config {
	return &Config{
		ServerAddress:   srvAddr,
		BaseURL:         hostName,
		FileStoragePath: filePath,
		DatabasePath:    database,
		EnableHTTPS:     ssl,
	}
}

func ParseConfig(path string, flags interface{}) {
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Println(err)
		}

		err = json.Unmarshal(data, &flags)
		if err != nil {
			log.Println(err)
		}
	}
}
