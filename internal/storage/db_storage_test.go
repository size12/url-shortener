package storage

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/size12/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewDBStorage(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewDBStorage(cfg)

	assert.NoError(t, err)

	// changing DB to mock DB.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.NoError(t, err, "Create new mock DB storage.")
	defer db.Close()

	s.DB = db
	_ = mock
}

func TestDBStorage_GetConfig(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewDBStorage(cfg)

	assert.NoError(t, err)
	assert.Equal(t, cfg, s.GetConfig())
}

func TestDBStorage(t *testing.T) {
	cfg := config.GetTestConfig()
	s, err := NewDBStorage(cfg)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual), sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err, "Create new mock DB storage.")
	defer db.Close()

	s.DB = db

	// add new links.
	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)").
		WithArgs("1", "https://yandex.ru", "user12", false).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://google.com").
		WillReturnRows(sqlmock.NewRows(nil))

	mock.ExpectExec("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)").
		WithArgs("2", "https://google.com", "user12", false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	result, err := s.CreateShort("user12", "https://yandex.ru", "https://google.com")
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, result)

	assert.NoError(t, mock.ExpectationsWereMet())

	// add existed link.

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))

	mock.ExpectCommit()

	result, err = s.CreateShort("user12", "https://yandex.ru")
	assert.Equal(t, Err409, err)
	assert.Equal(t, []string{"1"}, result)

	assert.NoError(t, mock.ExpectationsWereMet())

	// expecting error and rollback.

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	mock.ExpectQuery("SELECT id FROM links WHERE url = $1 LIMIT 1").WithArgs("https://yandex.ru").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1").RowError(0, errors.New("row error")))

	mock.ExpectRollback()

	_, err = s.CreateShort("user12", "https://yandex.ru")
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())

	// ping DB (no error).

	mock.ExpectPing()
	err = s.Ping()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// ping DB (with error).

	ErrPing := errors.New("failed ping DB")
	mock.ExpectPing().WillReturnError(ErrPing)
	err = s.Ping()
	assert.Equal(t, ErrPing, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

//
//func TestDBStorage(t *testing.T) {
//	cfg := config.GetTestConfig()
//	s, err := NewDBStorage(cfg)
//	if err != nil {
//		t.Log("Please run DB. Can't run tests.")
//		return
//	}
//	// Get config.
//	assert.Equal(t, cfg, s.GetConfig())
//
//	// Ping DB.
//	err = s.Ping()
//	assert.NoError(t, err)
//
//	// Creating short urls.
//	res, err := s.CreateShort("user12", "https://yandex.ru", "https://google.com")
//	if err != nil && err != Err409 {
//		t.Error("Failed create short urls: ", err)
//		return
//	}
//
//	// Getting long url.
//	long, err := s.GetLong(res[0])
//	if err != nil && err != Err404 {
//		t.Error("Failed create short urls: ", err)
//		return
//	}
//
//	assert.Equal(t, "https://yandex.ru", long)
//}
