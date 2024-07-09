package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
	"log"
)

const dbPath string = "./simple.db"
const cfgPath string = "./data.csv"
const dropTable bool = true
const batchSize int = 25

func OpenDbConn() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	return conn, err
}

type MonitoringService struct {
	Db *sql.DB
}

func (s MonitoringService) CloseDbConn() error {
	return s.Db.Close()
}

func (s MonitoringService) DropTable() error {
	statement, err := s.Db.Prepare("DROP TABLE IF EXISTS responses;")
	statement.Exec()
	return err
}

func (s MonitoringService) SetupDB() error {
	statement, err := s.Db.Prepare(
		"CREATE TABLE IF NOT EXISTS responses (id INTEGER PRIMARY KEY AUTOINCREMENT, ts TEXT NOT NULL, url TEXT NOT NULL, status TEXT NOT NULL, rt INT NOT NULL, regexp INTEGER NOT NULL)")
	statement.Exec()
	return err
}

func (s MonitoringService) saveToDB(responses responses) error {
	dbEntries, _ := s.Db.Prepare("INSERT INTO responses (ts, url, status, rt, regexp) VALUES (?, ?, ?, ?, ?)")
	tx, _ := s.Db.Begin()
	for _, response := range responses {
		_, err := tx.Stmt(dbEntries).Exec(response.ts, response.url, response.status, response.rt, response.regexp)
		if err != nil {
			panic(err)
		}
	}
	err := tx.Commit()
	return err
}

func (s MonitoringService) readFromDB() (dbEntries, error) {
	rows, err := s.Db.Query("SELECT * FROM responses;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbEntries dbEntries
	for rows.Next() {
		i := dbEntry{}
		err = rows.Scan(&i.id, &i.ts, &i.url, &i.status, &i.rt, &i.regexp)
		if err != nil {
			return nil, err
		}
		dbEntries = append(dbEntries, i)
	}
	return dbEntries, nil
}

type response struct {
	ts     string
	url    string
	status int
	rt     int64 // response time in ms
	regexp bool
}

type responses []response

type dbEntry struct {
	id     int
	ts     string
	url    string
	status int
	rt     int64
	regexp int // SQLite BOOL type is INT
}

type dbEntries []dbEntry

type urlConfiguration struct {
	id       int
	interval int
	url      string
	regexp   string
}

type urlConfigurations []urlConfiguration

// LoadConfigurations takes a file path to the configuration file, reads the configurations and returns back
// the url configurations. Each URL urlConfiguration includes an id, interval, url and regexp (optional).
func (s MonitoringService) LoadConfigurations(fileName string) (urlConfigurations, error) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	csvEntries, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var cfgs urlConfigurations
	for _, csvEntry := range csvEntries {
		id := csvEntry[0]
		interval := csvEntry[1]
		url := csvEntry[2]
		regexp := csvEntry[3]
				
		idInt, err := strconv.Atoi(id)
		if err != nil {
			fmt.Printf("Could not parse id %s from csv file", id)
			continue
		}

		intervalInt, err := strconv.Atoi(interval)
		if err != nil {
			fmt.Printf("Could not parse interval %s from csv file for entry number %s", interval, id)
			continue
		}

		cfg := urlConfiguration{
			id:       idInt,
			interval: intervalInt,
			url:      url,
			regexp:   regexp,
		}

		cfgs = append(cfgs, cfg)
	}
	return cfgs, nil
}


// MakeRequest reads each urlConfiguration, makes a HTTP request at desired interval,
// reads the response and also matches a regexp pattern (optional) before
// writing the result back to the channel
func (s MonitoringService) MakeRequest(cfg urlConfiguration, ch chan<- response) {
	ticker := time.NewTicker(time.Duration(cfg.interval) * time.Second)
	for range ticker.C {
		start := time.Now()

		ts := start.UTC()

		resp, err := http.Get(cfg.url)
		if err != nil {
			log.Printf("Request failed for %s", cfg.url)
		} else {
			
			status := resp.StatusCode

			rt := time.Since(start).Milliseconds()
			
			log.Printf("\t code %d \t %d ms \t %s", status, rt, cfg.url)

			body, _ := io.ReadAll(resp.Body)
	
			isRegexpMatch := false
	
			if cfg.regexp != "" {
				r, err := regexp.Compile(cfg.regexp)
				if err != nil {
					log.Printf("Compile regexp failed for", cfg.url)
				} else {
					isRegexpMatch = r.Match(body)
				}
			}
	
			response := response{
				ts:     ts.String(),
				url:    cfg.url,
				status: status,
				rt:     rt,
				regexp: isRegexpMatch,
			}
			ch <- response
		}
		
	}

}

func main() {
	// Create service
	conn, err := OpenDbConn()
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	service := MonitoringService{
		Db: conn,
	}

	// Setup DB
	if dropTable {
		err := service.DropTable()
		if err != nil {
			panic(err)
		}
	}

	service.SetupDB()

	// Load file urlConfiguration
	cfgs, err := service.LoadConfigurations(cfgPath)
	if err != nil {
		panic(err)
	}

	// Make requests
	ch := make(chan response)
	for _, cfg := range cfgs {
		go service.MakeRequest(cfg, ch)
	}

	// Read responses
	var responses responses
	for {
		responses = append(responses, <-ch)
		if len(responses) >= batchSize {
			service.saveToDB(responses)
			responses = nil
		}
	}

}
