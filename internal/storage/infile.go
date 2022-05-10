package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
	"log"
	"os"
)

type InFile struct {
	file    *os.File
	cache   map[string]string
	encoder *json.Encoder
}

func NewInFile(fileName string) (interfaces.Storage, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)

	if stat, _ := file.Stat(); stat.Size() != 0 {
		scanner := bufio.NewScanner(file)
		// optionally, resize scanner's capacity for lines over 64K, see next example
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			err := json.Unmarshal(scanner.Bytes(), &data)
			if err != nil {
				log.Fatal("DB file is damaged.")
			}
		}
	}

	return &InFile{
		file:    file,
		cache:   data,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *InFile) Close() error {
	s.cache = nil
	return s.file.Close()
}

func (s *InFile) GetURL(key string) (string, error) {
	if v, ok := s.cache[key]; ok {
		return v, nil
	}
	return "", interfaces.ErrNotFound
}

func (s *InFile) SetShortURL(key string, value string) error {
	if _, ok := s.cache[key]; ok {
		return interfaces.ErrAlreadyExists
	}
	s.cache[key] = value
	data := make(map[string]string)
	data[key] = value
	return s.encoder.Encode(&data)
}
