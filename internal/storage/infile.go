package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"sync"

	"github.com/ivanmyagkov/shortener.git/internal/interfaces"
)

type InFile struct {
	sync.Mutex
	file     *os.File
	DataFile ModelFile `json:"data_file"`
	Storage  map[string]string
	cache    map[string][]interfaces.ModelURL
	encoder  *json.Encoder
}

type ModelFile struct {
	UserID   string `json:"user_id"`
	ShortURL string `json:"short_url"`
	BaseURL  string `json:"base_url"`
}

func NewInFile(fileName string) (interfaces.Storage, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	var modelURL interfaces.ModelURL
	data := make(map[string][]interfaces.ModelURL)
	stor := make(map[string]string)
	var dataFile ModelFile

	if stat, _ := file.Stat(); stat.Size() != 0 {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			err := json.Unmarshal(scanner.Bytes(), &dataFile)
			if err != nil {
				return nil, err
			}
			stor[dataFile.ShortURL] = dataFile.BaseURL
		}
	}
	modelURL.ShortURL = dataFile.ShortURL
	modelURL.BaseURL = dataFile.BaseURL
	data[dataFile.UserID] = append(data[dataFile.UserID], modelURL)

	return &InFile{
		file:     file,
		DataFile: dataFile,
		Storage:  stor,
		cache:    data,
		encoder:  json.NewEncoder(file),
	}, nil
}

func (s *InFile) Close() {
	s.cache = nil
	if err := s.file.Close(); err != nil {
		log.Println(err)
	}
}

func (s *InFile) GetURL(key string) (string, error) {
	s.Lock()
	defer s.Unlock()
	if URL, ok := s.Storage[key]; ok {
		return URL, nil
	}

	return "", interfaces.ErrNotFound
}

func (s *InFile) GetAllURLsByUserID(userID string) ([]interfaces.ModelURL, error) {
	if _, ok := s.cache[userID]; ok {
		return s.cache[userID], nil
	}
	return nil, interfaces.ErrNotFound
}

func (s *InFile) SetShortURL(userID, key, value string) error {
	s.Lock()
	defer s.Unlock()
	s.DataFile.UserID = userID
	s.DataFile.ShortURL = key
	s.DataFile.BaseURL = value
	modelURL := interfaces.ModelURL{
		ShortURL: key,
		BaseURL:  value,
	}
	if _, ok := s.cache[userID]; ok {
		for _, val := range s.cache[userID] {
			if val.ShortURL == key {
				return nil
			}
		}
	}
	s.cache[userID] = append(s.cache[userID], modelURL)

	return s.encoder.Encode(&s.DataFile)

}
