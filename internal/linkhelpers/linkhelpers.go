package linkhelpers

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/size12/url-shortener/internal/config"
)

var Err409 = errors.New("link is already in storage")
var Err410 = errors.New("link is deleted, sorry :(")

type URLLinks struct {
	Cfg        config.Config
	Locations  map[string]string
	Users      map[string][]string
	Deleted    map[string]bool
	willDelete chan string
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

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS links (id varchar(255), url varchar(255), cookie varchar(255), deleted boolean)")
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
		var deleted bool
		err = rows.Scan(&short, &url, &cookie, &deleted)
		if err != nil {
			return err
		}
		links.Locations[short] = url
		links.Users[cookie] = append(links.Users[cookie], short)
		links.Deleted[short] = deleted
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
	deleted := make(map[string]bool)
	willDelete := make(chan string, 100)
	links := URLLinks{Locations: loc, Users: users, Deleted: deleted, Cfg: cfg, Mutex: &sync.Mutex{}, willDelete: willDelete}
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
		stmt, err = tx.PrepareContext(ctx, "INSERT INTO links (id, url, cookie, deleted) VALUES ($1, $2, $3, $4)")
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
		foundThisLink := false
		for id, link := range links.Locations {
			if link == longURL {
				newID = id
				isErr409 = Err409
				foundThisLink = true
				break
			}
		}

		result = append(result, newID)
		if foundThisLink {
			continue //do not add to storage again
		}

		links.Locations[newID] = longURL
		links.Users[cookie] = append(links.Users[cookie], newID)

		if links.File != nil {
			buf += longURL + "\n"
		}

		if links.DB != nil {
			if _, err := stmt.ExecContext(ctx, newID, longURL, cookie, false); err != nil {
				return nil, err
			}
		}

	}

	if links.DB != nil && len(result) > 0 {
		err := tx.Commit()
		if err != nil {
			return nil, err
		}
	}

	if links.File != nil && len(result) > 0 {
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
		var isErr410 error
		if links.Deleted[id] {
			isErr410 = Err410
		}
		return el, isErr410
	}
	return "", errors.New("no such id")
}

func (links *URLLinks) DeleteURLs(userID string, ids []string) error {
	links.Lock()
	defer links.Unlock()
	canDelete := links.Users[userID]

	for _, id := range ids {
		for _, can := range canDelete {
			if id == can {
				links.Deleted[id] = true
				links.willDelete <- id
				break
			}
		}
	}
	return nil
}

// DBDeleteURLs идея такая: ждем по максимуму значения из канала links.willDelete, если они заканчиваются, то всё это добваляем в базу
func (links *URLLinks) DBDeleteURLs() {
	if links.DB == nil {
		for {
			<-links.willDelete // do nothing, if db is not working
		}
	}
	var willDelete []string
	for {
		select {
		case id := <-links.willDelete:
			{
				willDelete = append(willDelete, id)
			}
		default:
			{
				if len(willDelete) != 0 {
					tx, err := links.DB.Begin()
					if err != nil {
						log.Println("something is wrong:", err.Error())
						continue
					}
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					stmt, err := tx.PrepareContext(ctx, "UPDATE links SET deleted = TRUE WHERE id = $1")
					if err != nil {
						log.Println("something is wrong:", err.Error())
						continue
					}

					for _, id := range willDelete {
						if _, err := stmt.ExecContext(ctx, id); err != nil {
							log.Println("something is wrong:", err.Error())
							continue
						}
					}

					err = tx.Commit()
					if err != nil {
						err = tx.Rollback()
						if err != nil {
							log.Fatal(err)
						}
					}

					willDelete = []string{}
					cancel()
				}
			}
		}
	}
}
