package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/size12/url-shortener/internal/config"
)

// FileStorage struct of file storage, implements storage.Storage.
type FileStorage struct {
	Cfg    config.Config
	File   *os.File
	LastID int
	*sync.Mutex
}

// GetConfig gets config.
func (s *FileStorage) GetConfig() config.Config {
	return s.Cfg
}

// Ping does nothing.
func (s *FileStorage) Ping() error {
	return nil
}

// NewFileStorage creates new file storage.
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

// CreateShort creates short url from long.
func (s *FileStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	s.Lock()
	defer s.Unlock()

	s.File.Seek(2, io.SeekEnd)

	result := make([]string, 0, len(urls))

	var builder strings.Builder

	for _, long := range urls {
		builder.WriteString(long)
		builder.WriteRune('\n')
		s.LastID++
		result = append(result, fmt.Sprint(s.LastID))
	}

	_, err := s.File.Write([]byte(builder.String()))
	if err != nil {
		return nil, err
	}
	//err = s.File.Sync()
	//if err != nil {
	//	return nil, err
	//}

	return result, nil
}

// GetLong gets long url from short.
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

// Delete does nothing.
func (s *FileStorage) Delete(userID string, ids ...string) error {
	// do nothing for file storage.
	return nil
}

// GetHistory gets history of urls.
func (s *FileStorage) GetHistory(userID string) ([]LinkJSON, error) {
	// return all links.
	var history []LinkJSON

	_, err := s.File.Seek(0, io.SeekStart)
	if err != nil {
		return history, err
	}

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

// GetStatistic gets total count of users and urls.
func (s *FileStorage) GetStatistic() (Statistic, error) {
	return Statistic{
		Urls:  s.LastID,
		Users: 0,
	}, nil
}
