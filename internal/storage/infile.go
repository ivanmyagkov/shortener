package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
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
		for scanner.Scan() {
			err := json.Unmarshal(scanner.Bytes(), &data)
			if err != nil {
				return nil, err
			}
		}
	}

	return &InFile{
		file:    file,
		cache:   data,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *InFile) Close() {
	s.cache = nil
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
