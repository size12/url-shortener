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
	cfg := config.Config{StoragePath: "file.txt"}
	_, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	// new file storage with empty file name
	cfg = config.Config{StoragePath: ""}
	_, err = NewFileStorage(cfg)
	assert.Equal(t, errors.New("empty file path"), err)
}

func TestFileStorage_GetConfig(t *testing.T) {
	cfg := config.Config{StoragePath: "file.txt"}
	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_Ping(t *testing.T) {
	cfg := config.Config{StoragePath: "file.txt"}
	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)
	assert.NoError(t, s.Ping())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_CreateShort(t *testing.T) {
	cfg := config.Config{StoragePath: "file.txt"}

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
	cfg := config.Config{StoragePath: "file.txt"}

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

		var longUrl string
		longUrl, err = s.GetLong(test.id)
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.want, longUrl)
	}

	assert.NoError(t, s.Ping())
	err = os.RemoveAll(cfg.StoragePath)
	assert.NoError(t, err)
}

func TestFileStorage_Delete(t *testing.T) {
	// can't delete from file storage
	cfg := config.Config{StoragePath: "1.txt"}
	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	err = s.Delete("user12", "1234") // do nothing.
	assert.NoError(t, err)

	os.RemoveAll("1.txt")
}

func TestFileStorage_GetHistory(t *testing.T) {
	cfg := config.Config{StoragePath: "1.txt", BaseURL: "http://127.0.0.1"}
	s, err := NewFileStorage(cfg)
	assert.NoError(t, err)

	_, err = s.CreateShort("user12", "https://yandex.ru", "https://google.com", "https://youtube.com")
	assert.NoError(t, err)

	res, err := s.GetHistory("user12")
	assert.NoError(t, err)

	assert.Equal(t, []LinkJSON{
		{
			ShortURL: "http://127.0.0.1/1",
			LongURL:  "https://yandex.ru",
		},
		{
			ShortURL: "http://127.0.0.1/2",
			LongURL:  "https://google.com",
		},
		{
			ShortURL: "http://127.0.0.1/3",
			LongURL:  "https://youtube.com",
		},
	}, res)

	err = os.RemoveAll("1.txt")
	assert.NoError(t, err)
}
