package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/size12/url-shortener/internal/config"
)

type FileStorage struct {
	Cfg    config.Config
	File   *os.File
	LastID int
	*sync.Mutex
}

func (s *FileStorage) GetConfig() config.Config {
	return s.Cfg
}

func (s *FileStorage) Ping() error {
	return nil
}

func NewFileStorage(cfg config.Config) (*FileStorage, error) {
	s := &FileStorage{Cfg: cfg, Mutex: &sync.Mutex{}}

	if cfg.StoragePath == "" {
		return s, errors.New("empty file path")
	}

	file, err := os.OpenFile(cfg.StoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0777)
	if err != nil {
		return s, err
	}

	s.File = file

	scanner := bufio.NewScanner(file)
	id := 0
	for scanner.Scan() {
		id++
	}

	if err := scanner.Err(); err != nil {
		return s, err
	}

	s.LastID = id

	return s, nil
}

func (s *FileStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	s.Lock()
	defer s.Unlock()

	s.File.Seek(2, io.SeekEnd)

	var buffer string
	var result []string

	for _, long := range urls {
		buffer += long + "\n"
		s.LastID++
		result = append(result, fmt.Sprint(s.LastID))
	}

	_, err := s.File.Write([]byte(buffer))
	if err != nil {
		return nil, err
	}
	err = s.File.Sync()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *FileStorage) GetLong(id string) (string, error) {
	s.Lock()
	defer s.Unlock()

	s.File.Seek(0, io.SeekStart)

	scanner := bufio.NewScanner(s.File)
	i := 0
	for scanner.Scan() {
		i++
		long := scanner.Text()
		if fmt.Sprint(i) == id {
			return long, scanner.Err()
		}
	}

	return "", Err404
}

func (s *FileStorage) Delete(userID string, ids ...string) error {
	//do nothing for file storage
	return nil
}

func (s *FileStorage) GetHistory(userID string) ([]LinkJSON, error) {
	//return all links
	var history []LinkJSON

	scanner := bufio.NewScanner(s.File)
	id := 0
	for scanner.Scan() {
		id++
		long := scanner.Text()
		history = append(history, LinkJSON{ShortURL: s.Cfg.BaseURL + "/" + fmt.Sprint(id), LongURL: long})
	}

	if err := scanner.Err(); err != nil {
		return history, err
	}

	return history, nil
}