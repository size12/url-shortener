package storage

import (
	"os"
	"reflect"
	"testing"

	"github.com/size12/url-shortener/internal/config"
	"github.com/stretchr/testify/assert"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestNewStorage(t *testing.T) {
	// new map storage
	cfg := config.Config{}
	s, err := NewStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(&MapStorage{}), reflect.TypeOf(s))

	// new file storage
	cfg = config.Config{StoragePath: "1.txt"}
	s, err = NewStorage(cfg)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(&FileStorage{}), reflect.TypeOf(s))
	err = os.RemoveAll("1.txt")
	assert.NoError(t, err)
}
