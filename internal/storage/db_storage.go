package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/url-shortener/internal/config"
)

type DBStorage struct {
	Cfg    config.Config
	DB     *sql.DB
	LastID int
}

func (s *DBStorage) GetConfig() config.Config {
	return s.Cfg
}

func (s *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}

func NewDBStorage(cfg config.Config) (*DBStorage, error) {
	s := &DBStorage{Cfg: cfg}

	db, err := sql.Open("pgx", cfg.BasePath)

	if err != nil {
		return s, err
	}

	s.DB = db

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS links (id varchar(255), url varchar(255), cookie varchar(255), deleted boolean)")
	if err != nil {
		return s, err
	}

	rows, err := db.QueryContext(ctx, "SELECT COUNT(*) FROM links")

	if err != nil {
		return s, err
	}

	defer rows.Close()

	rows.Next()
	err = rows.Scan(&s.LastID)
	if err != nil {
		return s, err
	}

	err = rows.Err()
	if err != nil {
		return s, err
	}

	return s, nil
}

func (s *DBStorage) CreateShort(userID string, urls ...string) ([]string, error) {
	var isErr409 error
	var result []string

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

	for _, url := range urls {
		var isAdded bool
		rows, err := s.DB.QueryContext(ctx, "SELECT id FROM links WHERE url = $1 LIMIT 1", url)
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
			isAdded = true
			result = append(result, id)
		}

		if err := rows.Err(); err != nil {
			return result, err
		}
		if !isAdded {
			s.LastID++
			newID := fmt.Sprint(s.LastID)
			if _, err := stmt.ExecContext(ctx, newID, url, userID, false); err != nil {
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

func (s *DBStorage) GetLong(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := s.DB.QueryContext(ctx, "SELECT url, deleted FROM links WHERE id=$1 LIMIT 1", id)

	if err != nil {
		return "", err
	}

	defer rows.Close()

	var long string
	var deleted bool

	rows.Next()
	err = rows.Scan(&long, &deleted)

	if err != nil {
		return "", Err404
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	if deleted {
		return "", Err410
	}

	return long, nil
}

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
		if _, err := stmt.ExecContext(ctx, id, userID); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

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
