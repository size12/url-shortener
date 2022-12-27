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
		i += 1
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

func (links *URLLinks) NewShortURL(longURL, cookie string) (string, error) {
	if _, err := url.ParseRequestURI(longURL); err != nil {
		return "", errors.New("wrong link " + longURL) //checks if url valid
	}
	links.Lock()
	defer links.Unlock()
	lastID := len(links.Locations)
	newID := fmt.Sprint(lastID + 1)
	links.Locations[newID] = longURL
	if links.File != nil {
		_, err := links.File.Write([]byte(longURL + "\n"))
		if err != nil {
			return "", err
		}
		err = links.File.Sync()
		if err != nil {
			return "", err
		}
	}
	if links.DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_, err := links.DB.ExecContext(ctx, "INSERT INTO links (id, url, cookie) VALUES ($1, $2,  $3)", newID, longURL, cookie)

		if err != nil {
			return "", err
		}
	}

	links.Users[cookie] = append(links.Users[cookie], newID)

	return newID, nil
}

func (links *URLLinks) GetFullURL(id string) (string, error) {
	links.Lock()
	defer links.Unlock()
	if el, ok := links.Locations[id]; ok {
		return el, nil
	}
	return "", errors.New("no such id")
}
