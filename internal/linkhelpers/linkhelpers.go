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

type RequestJSON struct {
	URL string `json:"url"`
}

type ResponseJSON struct {
	Result string `json:"result"`
}

func NewStorage(cfg config.Config) (URLLinks, error) {
	loc := make(map[string]string)
	users := make(map[string][]string)
	db, err := sql.Open("pgx", cfg.BasePath)
	if err == nil && cfg.BasePath != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		rows, err := db.QueryContext(ctx, "SELECT * FROM links")
		if err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, err := db.ExecContext(ctx, "CREATE TABLE links (short varchar(255), url varchar(255), cookie varchar(255))")
			if err != nil {
				fmt.Println("Can't create new table:", err.Error())
			}
		} else {
			for rows.Next() {
				var short string
				var url string
				var cookie string
				err = rows.Scan(&short, &url, &cookie)
				loc[short] = url
				users[cookie] = append(users[cookie], short)
				if err != nil {
					break
				}
			}

			err = rows.Err()
			if err != nil {
				return URLLinks{}, err
			}
		}
		defer rows.Close()
		fmt.Println(loc, users)
	}

	if cfg.StoragePath != "" {
		file, err := os.OpenFile(cfg.StoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0777)
		if err != nil {
			return URLLinks{}, err
		}
		scanner := bufio.NewScanner(file)
		i := 1
		for scanner.Scan() {
			loc[fmt.Sprint(i)] = scanner.Text()
			i += 1
		}
		if err := scanner.Err(); err != nil {
			return URLLinks{}, err
		}
		return URLLinks{Locations: loc, Mutex: &sync.Mutex{}, File: file, Cfg: cfg, Users: users, DB: db}, nil
	}
	return URLLinks{Locations: loc, Mutex: &sync.Mutex{}, Cfg: cfg, Users: users, DB: db}, nil
}

func (Links *URLLinks) NewShortURL(longURL string) (string, error) {
	if _, err := url.ParseRequestURI(longURL); err != nil {
		return "", errors.New("wrong link " + longURL) //checks if url valid
	}
	Links.Lock()
	defer Links.Unlock()
	lastID := len(Links.Locations)
	newID := fmt.Sprint(lastID + 1)
	Links.Locations[newID] = longURL
	if Links.File != nil {
		_, err := Links.File.Write([]byte(longURL + "\n"))
		if err != nil {
			return "", err
		}
		err = Links.File.Sync()
		if err != nil {
			return "", err
		}
	}
	return newID, nil
}

func (Links *URLLinks) GetFullURL(id string) (string, error) {
	Links.Lock()
	defer Links.Unlock()
	if el, ok := Links.Locations[id]; ok {
		return el, nil
	}
	return "", errors.New("no such id")
}
