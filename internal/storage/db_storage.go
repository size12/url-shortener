package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/size12/url-shortener/internal/config"
)

// DBStorage is storage that uses DB.
// Implements storage.Storage interface.
type DBStorage struct {
	Cfg    config.Config
	DB     *sql.DB
	LastID int
}

// Interface storage.Storage implementation.

// GetConfig gets config from storage.
func (s *DBStorage) GetConfig() config.Config {
	return s.Cfg
}

// Ping check connection to storage.
func (s *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}

// NewDBStorage creates new DB storage.
func NewDBStorage(cfg config.Config) (*DBStorage, error) {
	s := &DBStorage{Cfg: cfg, LastID: 0}

	if cfg.BasePath == "mockedDB" {
		return s, nil
	}

	db, err := sql.Open("pgx", cfg.BasePath)

	if err != nil {
		return s, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = migrateUP(db, cfg)

	if err != nil {
		log.Println("Failed migrate DB: ", err)
		return s, err
	}

	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM links")

	err = row.Scan(&s.LastID)

	if err != nil {
		return s, err
	}

	err = row.Err()
	if err != nil {
		return s, err
	}

	s.DB = db

	return s, nil
}

// migrateUP DB migrations.
func migrateUP(db *sql.DB, cfg config.Config) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Printf("Failed create postgres instance: %v\n", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		cfg.DBMigrationPath,
		"pgx", driver)

	if err != nil {
		log.Printf("Failed create migration instance: %v\n", err)
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Println("Failed migrate: ", err)
		return err
	}

	return nil
}

// CreateShort creates short url from long.
func (s *DBStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	var isErr409 error
	result := make([]string, 0, len(urls))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := s.DB.Begin()
	defer tx.Rollback()

	if err != nil {
		return result, err
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return result, err
	}
	defer stmt.Close()

	var rows *sql.Rows

	for _, url := range urls {
		var alreadyAdded bool
		rows, err = s.DB.QueryContext(ctx, "SELECT id FROM links WHERE url = $1 LIMIT 1", url)
		if err != nil {
			return result, err
		}
		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			if err != nil {
				return result, err
			}
			isErr409 = Err409
			alreadyAdded = true
			result = append(result, id)
		}

		if err = rows.Err(); err != nil {
			return result, err
		}
		if !alreadyAdded {
			s.LastID++
			newID := fmt.Sprint(s.LastID)
			if _, err = stmt.ExecContext(ctx, newID, url, userID, false); err != nil {
				return result, err
			}
			result = append(result, newID)
		}
	}

	err = tx.Commit()
	if err != nil {
		return result, err
	}

	return result, isErr409
}

// GetLong gets long url from short.
func (s *DBStorage) GetLong(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	row := s.DB.QueryRowContext(ctx, "SELECT url, deleted FROM links WHERE id=$1 LIMIT 1", id)

	var long string
	var deleted bool

	err := row.Scan(&long, &deleted)

	if errors.Is(err, sql.ErrNoRows) {
		return "", Err404
	}

	if err := row.Err(); err != nil {
		return "", err
	}

	if deleted {
		return "", Err410
	}

	return long, nil
}

// Delete deletes url.
func (s *DBStorage) Delete(userID string, ids ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	tx, err := s.DB.Begin()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "UPDATE links SET deleted = TRUE WHERE id = $1 AND cookie = $2")
	if err != nil {
		return err
	}

	for _, id := range ids {
		if _, err = stmt.ExecContext(ctx, id, userID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// GetHistory gets history of links.
func (s *DBStorage) GetHistory(userID string) ([]LinkJSON, error) {
	var history []LinkJSON

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT id, url FROM links WHERE cookie=$1", userID)

	if err != nil {
		return history, err
	}

	defer rows.Close()

	for rows.Next() {
		var id string
		var long string
		err = rows.Scan(&id, &long)

		if err != nil {
			return history, err
		}

		history = append(history, LinkJSON{
			ShortURL: id,
			LongURL:  long,
		})

	}

	if err := rows.Err(); err != nil {
		return history, err
	}

	return history, nil
}
