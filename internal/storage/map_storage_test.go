package storage

import (
	"errors"
	"testing"

	"github.com/size12/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMapStorage_CreateShort(t *testing.T) {
	cfg := config.Config{}

	tc := []struct {
		name string
		urls []string
		want []string
		loc  map[string]string
		err  error
	}{
		{
			"Add single good link",
			[]string{"https://yandex.ru"},
			[]string{"1"},
			map[string]string{"1": "https://yandex.ru"},
			nil,
		},
		{
			"Add multiple good links",
			[]string{"https://yandex.ru", "https://google.com"},
			[]string{"1", "2"},
			map[string]string{"1": "https://yandex.ru", "2": "https://google.com"},
			nil,
		},
		{
			"Add multiple same links",
			[]string{"https://yandex.ru", "https://yandex.ru"},
			[]string{"1", "1"},
			map[string]string{"1": "https://yandex.ru"},
			Err409,
		},
		{
			"Add bad single link",
			[]string{"not_url"},
			nil,
			map[string]string{},
			errors.New("wrong link not_url"),
		},
		{
			"Add bad multiple links",
			[]string{"not_url", "not_url2"},
			nil,
			map[string]string{},
			errors.New("wrong link not_url"),
		},
		{
			"Don't pass links",
			nil,
			[]string{},
			map[string]string{},
			nil,
		},
	}

	for _, test := range tc {
		s, err := NewMapStorage(cfg)
		assert.NoError(t, err, test.name)
		res, err := s.CreateShort("user12", test.urls...)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.want, res, test.name)
		assert.Equal(t, test.loc, s.Locations)
	}

}

func TestMapStorage_GetConfig(t *testing.T) {
	cfg := config.GetDefaultConfig()
	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
}

func TestMapStorage_GetLong(t *testing.T) {
	cfg := config.Config{}

	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	tc := []struct {
		name   string
		loc    map[string]string
		id     string
		result string
		err    error
	}{
		{
			"Get long exists url",
			map[string]string{"1": "https://yandex.ru"},
			"1",
			"https://yandex.ru",
			nil,
		},
		{
			"Get long non-exists url",
			map[string]string{"1": "https://yandex.ru"},
			"2",
			"",
			Err404,
		},
	}

	for _, test := range tc {
		s.Locations = test.loc
		res, err := s.GetLong(test.id)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.result, res, test.name)
	}
}

func TestMapStorage_Delete(t *testing.T) {
	cfg := config.GetDefaultConfig()
	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	tc := []struct {
		name        string
		loc         map[string]string
		users       map[string][]string
		deleted     map[string]bool
		id          string
		wantDeleted map[string]bool
		err         error
	}{
		{
			"user can delete link",
			map[string]string{"1": "https://yandex.ru"},
			map[string][]string{
				"user1": {"1"},
			},
			map[string]bool{"1": false},
			"1",
			map[string]bool{"1": true},
			nil,
		},
		{
			"user can't delete link",
			map[string]string{"1": "https://yandex.ru", "2": "https://google.com"},
			map[string][]string{
				"user1": {"2"},
			},
			map[string]bool{"1": false, "2": false},
			"1",
			map[string]bool{"1": false, "2": false},
			nil,
		},
	}

	for _, test := range tc {
		s.Locations = test.loc
		s.Users = test.users
		s.Deleted = test.deleted

		err := s.Delete("user1", test.id)
		assert.Equal(t, test.err, err, test.name)
		assert.Equal(t, test.wantDeleted, s.Deleted, test.name)
	}

}

func TestMapStorage_GetHistory(t *testing.T) {
	cfg := config.Config{BaseURL: "http://127.0.0.1"}
	s, err := NewMapStorage(cfg)
	assert.NoError(t, err)

	tc := []struct {
		name   string
		loc    map[string]string
		users  map[string][]string
		cookie string
		want   []LinkJSON
		err    error
	}{

		{
			"no history",
			map[string]string{"1": "https://yandex.ru", "2": "https://google.com"},
			map[string][]string{
				"user1": {"1"}, "user2": {"2"},
			},
			"user_unknown",
			[]LinkJSON{},
			nil,
		},
		{
			"get history first user",
			map[string]string{"1": "https://yandex.ru", "2": "https://google.com"},
			map[string][]string{
				"user1": {"1"}, "user2": {"2"},
			},
			"user1",
			[]LinkJSON{
				{
					ShortURL: cfg.BaseURL + "/" + "1",
					LongURL:  "https://yandex.ru",
				},
			},
			nil,
		},
		{
			"get history second user",
			map[string]string{"1": "https://yandex.ru", "2": "https://google.com"},
			map[string][]string{
				"user1": {"1"}, "user2": {"2"},
			},
			"user2",
			[]LinkJSON{
				{
					ShortURL: cfg.BaseURL + "/" + "2",
					LongURL:  "https://google.com",
				},
			},
			nil,
		},
	}

	for _, test := range tc {
		s.Locations = test.loc
		s.Users = test.users

		res, err := s.GetHistory(test.cookie)
		assert.Equal(t, test.err, err)
		assert.Equal(t, test.want, res)
	}

}
