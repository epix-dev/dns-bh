package main

import (
	"crypto/tls"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/epix-dev/dns-bh/lib"

	_ "github.com/lib/pq"
)

var (
	version = "dev"
	build   = "none"
	author  = "undefined"
)

type TimeWithoutTZ struct {
	time.Time
	Valid bool
}

func (m *TimeWithoutTZ) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// Parse with time format which omit TZ
	tt, err := time.Parse(`"2006-01-02T15:04:05"`, string(data))
	*m = TimeWithoutTZ{tt, true}
	return err
}

func (m TimeWithoutTZ) Value() (driver.Value, error) {
	if m.Valid == false {
		return nil, nil
	}
	return driver.Value(m.Time.Format(`"2006-01-02T15:04:05"`)), nil
}

type domain struct {
	RegisterPositionID int           `json:"RegisterPositionId"`
	Domain             string        `json:"DomainAddress"`
	InsertTime         TimeWithoutTZ `json:"InsertDate"`
	DeleteTime         TimeWithoutTZ `json:"DeleteDate"`
}

func arrayContaintsMalware(a domain, list []domain) bool {
	for _, b := range list {
		if b.Domain == a.Domain && b.RegisterPositionID == a.RegisterPositionID {
			return true
		}
	}

	return false
}

func loadPositions(url string, items *[]domain) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		log.Panicln(err)
	}

	return nil
}

func updateItems(db *sql.DB, items []domain) []string {
	var result []string

	inserSQL := `
	INSERT INTO cert_hole (
		remote_id,
		domain,
		insert_time,
		delete_time,
		created_at,
		deleted_at
	) values($1, $2, $3, $4, CURRENT_TIMESTAMP, $4)
	`

	updateSQL := "UPDATE cert_hole SET updated_at = CURRENT_TIMESTAMP, delete_time = $2, deleted_at=$2 WHERE remote_id = $1"

	insertStmt, err := db.Prepare(inserSQL)
	lib.CheckError(err)
	defer insertStmt.Close()

	updateStmt, err := db.Prepare(updateSQL)
	lib.CheckError(err)
	defer updateStmt.Close()

	for _, item := range items {

		updateRes, err := updateStmt.Exec(item.RegisterPositionID, item.DeleteTime)
		lib.CheckError(err)

		if affect, _ := updateRes.RowsAffected(); affect == 0 {
			_, err := insertStmt.Exec(
				item.RegisterPositionID,
				item.Domain,
				item.InsertTime.Time,
				item.DeleteTime)
			lib.CheckError(err)

			log.Printf("added %d:%s", item.RegisterPositionID, item.Domain)
			result = append(result, item.Domain)
		}

	}

	return result
}

func main() {
	var err error
	var db *sql.DB

	var cfgDir string
	var cfg lib.Config
	var domains []domain

	program := filepath.Base(os.Args[0])

	log.Printf("%s started, version: %s+%s, author: %s\n", program, version, build, author)

	flag.StringVar(&cfgDir, "cfg-dir", "/opt/dns-bh/etc", "Config dir path")
	flag.Parse()

	lib.ConfigInit(cfgDir)
	if !lib.ConfigLoad(&cfg) {
		os.Exit(1)
	}

	db, err = lib.ConnectDb(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSL)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := loadPositions("https://hole.cert.pl/domains/domains.json", &domains); err != nil {
		log.Fatalln(err)
	}

	lib.ReportChanges(&cfg, updateItems(db, domains), "Added malware domains from hole.cert.pl")
	// TODO: implements deleteItems() for report use or something usable.
	// lib.ReportChanges(&cfg, deleteItems(db, domains), "Deleted malware domains from hole.cert.pl")
}
