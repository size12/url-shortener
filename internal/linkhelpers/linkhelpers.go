package linkhelpers

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/url-shortener/internal/config"
	"net/url"
	"os"
	"sync"
	"time"
)

var Err409 = errors.New("link is already in storage")

type URLLinks struct {
	Cfg       config.Config
	Locations map[string]string
	Users     map[string][]string
	*sync.Mutex
	DB   *sql.DB
	File *os.File
}

type LinkJSON struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

type BatchJSON struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	URL           string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
}

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

func (links *URLLinks) OpenDB() error {
	if links.Cfg.BasePath == "" {
		links.DB = nil
		return errors.New("empty path for database")
	}

	db, err := sql.Open("pgx", links.Cfg.BasePath)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS links (id varchar(255), url varchar(255), cookie varchar(255))")
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT * FROM links")

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var short string
		var url string
		var cookie string
		err = rows.Scan(&short, &url, &cookie)
		if err != nil {
			return err
		}
		links.Locations[short] = url
		links.Users[cookie] = append(links.Users[cookie], short)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	links.DB = db
	return nil
}
func (links *URLLinks) OpenFile() error {
	if links.Cfg.StoragePath == "" {
		return errors.New("empty path for file")
	}

	file, err := os.OpenFile(links.Cfg.StoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0777)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	i := 1
	for scanner.Scan() {
		links.Locations[fmt.Sprint(i)] = scanner.Text()
		i++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	links.File = file
	return nil
}

func NewStorage(cfg config.Config) (URLLinks, error) {
	loc := make(map[string]string)
	users := make(map[string][]string)
	links := URLLinks{Locations: loc, Users: users, Cfg: cfg, Mutex: &sync.Mutex{}}
	err := links.OpenDB()
	if err != nil {
		fmt.Printf("Failed connect db: %v\n", err)
	}
	err = links.OpenFile()
	if err != nil {
		fmt.Printf("Failed connect to file: %v\n", err)
	}

	return links, nil
}

func (links *URLLinks) NewShortURL(cookie string, urls ...string) ([]string, error) {
	result := make([]string, 0)
	var buf string
	links.Lock()
	defer links.Unlock()
	var isErr409 error
	var tx *sql.Tx
	var stmt *sql.Stmt
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if links.DB != nil {
		var err error
		tx, err = links.DB.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
		stmt, err = tx.PrepareContext(ctx, "INSERT INTO links (id, url, cookie) VALUES ($1, $2, $3)")
		if err != nil {
			return nil, err
		}
		defer stmt.Close()
	}

	for _, longURL := range urls {
		if _, err := url.ParseRequestURI(longURL); err != nil {
			return nil, errors.New("wrong link " + longURL) //checks if url valid
		}

		lastID := len(links.Locations)
		newID := fmt.Sprint(lastID + 1)
		for id, link := range links.Locations {
			if link == longURL {
				newID = fmt.Sprint(id)
				isErr409 = Err409
				break
			}
		}
		links.Locations[newID] = longURL
		links.Users[cookie] = append(links.Users[cookie], newID)
		result = append(result, newID)

		if links.File != nil {
			buf += longURL + "\n"
		}

		if links.DB != nil {
			if _, err := stmt.ExecContext(ctx, newID, longURL, cookie); err != nil {
				return nil, err
			}
		}

	}

	if links.DB != nil {
		err := tx.Commit()
		if err != nil {
			return nil, err
		}
	}

	if links.File != nil {
		_, err := links.File.Write([]byte(buf))
		if err != nil {
			return nil, err
		}
		err = links.File.Sync()
		if err != nil {
			return nil, err
		}
	}

	return result, isErr409
}

func (links *URLLinks) GetFullURL(id string) (string, error) {
	links.Lock()
	defer links.Unlock()
	if el, ok := links.Locations[id]; ok {
		return el, nil
	}
	return "", errors.New("no such id")
}
