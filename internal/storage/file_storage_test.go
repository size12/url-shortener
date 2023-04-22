package storage

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/size12/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewFileStorage(t *testing.T) {
	// new file storage with good file name
	cfg := config.GetTestConfig()
	_, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)

	// new file storage with empty file name
	cfg = config.GetTestConfig()
	cfg.StoragePath = ""
	_, err = NewFileStorage(cfg)
	assert.Equal(t, errors.New("empty file path"), err)
}

func TestFileStorage_GetConfig(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_Ping(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)
	assert.NoError(t, s.Ping())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_CreateShort(t *testing.T) {
	cfg := config.GetTestConfig()

	tc := []struct {
		name       string
		urls       []string
		want       []string
		wantInFile []string
		err        error
	}{
		{
			"add single good link to file storage",
			[]string{"https://yandex.ru"},
			[]string{"1"},
			[]string{"https://yandex.ru"},
			nil,
		},
		{
			"add multiple good links to file storage",
			[]string{"https://youtube.com", "https://google.com"},
			[]string{"2", "3"},
			[]string{"https://youtube.com", "https://google.com"},
			nil,
		},
	}

	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	for _, test := range tc {
		s, err = NewFileStorage(cfg)
		assert.NoError(t, err)

		var res []string
		res, err = s.CreateShort("user12", test.urls...)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.want, res, test.name)
		_, err = s.File.Seek(0, io.SeekStart)
		assert.NoError(t, err)
		var text []byte
		text, err = io.ReadAll(s.File)
		assert.NoError(t, err)

		for _, url := range test.wantInFile {
			assert.Contains(t, string(text), url, test.name)
		}
	}

	assert.NoError(t, s.Ping())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_GetLong(t *testing.T) {
	cfg := config.GetTestConfig()

	tc := []struct {
		name string
		urls []string
		id   string
		want string
		err  error
	}{
		{
			"add urls and get long by id",
			[]string{"https://yandex.ru", "https://google.com"},
			"1",
			"https://yandex.ru",
			nil,
		},
		{
			"add urls and get long by non exists id",
			[]string{"https://yandex.ru", "https://google.com"},
			"9",
			"",
			Err404,
		},
	}

	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	for _, test := range tc {
		s, err = NewFileStorage(cfg)
		assert.NoError(t, err)

		_, err = s.CreateShort("user12", test.urls...)
		assert.NoError(t, err)

		var longURL string
		longURL, err = s.GetLong(test.id)
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.want, longURL)
	}

	assert.NoError(t, s.Ping())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_Delete(t *testing.T) {
	// can't delete from file storage
	cfg := config.GetTestConfig()

	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	err = s.Delete("user12", "1234") // do nothing.
	assert.NoError(t, err)

	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_GetHistory(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	// get history from empty file.

	history, err := s.GetHistory("user12")

	assert.NoError(t, err)
	assert.Empty(t, history)

	// get history from non-empty file.
	_, err = s.CreateShort("user12", "https://yandex.ru", "https://google.com", "https://youtube.com")
	assert.NoError(t, err)

	history, err = s.GetHistory("user12")
	assert.NoError(t, err)

	assert.Equal(t, []LinkJSON{
		{
			ShortURL: cfg.BaseURL + "/1",
			LongURL:  "https://yandex.ru",
		},
		{
			ShortURL: cfg.BaseURL + "/2",
			LongURL:  "https://google.com",
		},
		{
			ShortURL: cfg.BaseURL + "/3",
			LongURL:  "https://youtube.com",
		},
	}, history)

	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_GetStatistic(t *testing.T) {
	cfg := config.GetTestConfig()

	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	stat, err := s.GetStatistic()
	assert.NoError(t, err)
	assert.Equal(t, Statistic{
		Urls:  0,
		Users: 0,
	}, stat)

	// add url and get statistic again.
	_, err = s.CreateShort("user12", "https:/yandex.ru")
	assert.NoError(t, err)
	stat, err = s.GetStatistic()
	assert.NoError(t, err)
	assert.Equal(t, Statistic{
		Urls:  1,
		Users: 0,
	}, stat)
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}
