package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/epix/dns-bh/lib"

	_ "github.com/lib/pq"
)

type registryPosition struct {
	Domain   string `xml:"AdresDomeny"`
	EntryPos int    `xml:"Lp,attr"`
	EntryAdd string `xml:"DataWpisu"`
	EntryDel string `xml:"DataWykreslenia"`
}

type registry struct {
	XMLName   xml.Name           `xml:"Rejestr"`
	Positions []registryPosition `xml:"PozycjaRejestru"`
}

func fetchBody(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func updateItems(db *sql.DB, items []registryPosition) []string {
	var result []string

	inserSQL := `
	INSERT INTO hazard (
		entry_pos,
		entry_add,
		domain,
		created_at
	) values($1, $2, $3, CURRENT_TIMESTAMP)
	`

	updateSQL := "UPDATE hazard SET updated_at = CURRENT_TIMESTAMP WHERE entry_pos = $1 AND deleted_at IS NULL"
	deleteSQL := "UPDATE hazard SET entry_del = $1, deleted_at = CURRENT_TIMESTAMP WHERE deleted_at IS NULL AND entry_pos = $2"

	insertStmt, err := db.Prepare(inserSQL)
	lib.CheckError(err)
	defer insertStmt.Close()

	updateStmt, err := db.Prepare(updateSQL)
	lib.CheckError(err)
	defer updateStmt.Close()

	deleteStmt, err := db.Prepare(deleteSQL)
	lib.CheckError(err)
	defer deleteStmt.Close()

	for _, item := range items {
		if item.EntryDel == "" {
			updateRes, err := updateStmt.Exec(item.EntryPos)
			lib.CheckError(err)

			if affect, _ := updateRes.RowsAffected(); affect == 0 {
				_, err := insertStmt.Exec(item.EntryPos, item.EntryAdd, item.Domain)
				lib.CheckError(err)
				log.Printf("added %d:%s", item.EntryPos, item.Domain)
				result = append(result, item.Domain)
			}
		} else {
			updateRes, err := deleteStmt.Exec(item.EntryDel, item.EntryPos)
			lib.CheckError(err)

			if affect, _ := updateRes.RowsAffected(); affect == 1 {
				log.Printf("deleted %d:%s", item.EntryPos, item.Domain)
				result = append(result, item.Domain)
			}
		}
	}

	return result
}

func deleteItems(db *sql.DB, items []registryPosition) []string {
	var result []string

	var positions []int

	for _, item := range items {
		positions = append(positions, item.EntryPos)
	}

	selectSQL := "SELECT entry_pos, domain FROM hazard WHERE deleted_at IS NULL"
	updateSQL := "UPDATE hazard SET deleted_at = CURRENT_TIMESTAMP WHERE deleted_at IS NULL AND entry_pos = $1"

	updateStmt, err := db.Prepare(updateSQL)
	lib.CheckError(err)
	defer updateStmt.Close()

	rows, err := db.Query(selectSQL)
	lib.CheckError(err)

	for rows.Next() {
		var entryPos int
		var domain string

		err = rows.Scan(&entryPos, &domain)
		lib.CheckError(err)

		if !lib.ArrayContaintsInt(entryPos, positions) {
			_, err := updateStmt.Exec(entryPos)
			lib.CheckError(err)
			log.Printf("marked %d:%s as deleted", entryPos, domain)
			result = append(result, domain)
		}
	}

	return result
}

type server struct {
	cfg     lib.Config
	address string
	port    int
	closer  io.Closer
}

func (s *server) start(h http.Handler) error {
	var err error
	var addr string
	var listener net.Listener
	srv := &http.Server{Addr: addr, Handler: h}

	addr = fmt.Sprintf("%s:%d", s.address, s.port)

	listener, err = net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	go func() {
		err := srv.Serve(listener)
		if err != nil {
			log.Printf("HTTP Server Error: %v ", err)
		}
	}()

	s.closer = io.Closer(listener)

	return nil
}

func (s *server) loadRegistry(body []byte, deleteOld bool) {
	var err error
	var db *sql.DB
	var reg registry

	db, err = lib.ConnectDb(s.cfg.DB.Host, s.cfg.DB.Port, s.cfg.DB.User, s.cfg.DB.Password, s.cfg.DB.Name)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := xml.Unmarshal(body, &reg); err != nil {
		log.Println(err)
	}

	log.Printf("loaded %d entries\n", len(reg.Positions))

	lib.ReportChanges(&s.cfg, updateItems(db, reg.Positions), "Added hazard domains")

	if deleteOld == true {
		lib.ReportChanges(&s.cfg, deleteItems(db, reg.Positions), "Deleted hazard domains")
	}
}

func (s *server) pull() {
	if body, err := fetchBody("https://hazard.mf.gov.pl/api/Register"); err == nil {
		s.loadRegistry(body, true)
	}
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Header().Set("Rsh-Push", "accepted")

		if body, err := ioutil.ReadAll(r.Body); err != nil {
			log.Print("handled PUSH request: StatusInternalServerError")
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Print("handled PUSH request: StatusOK")
			s.loadRegistry(body, false)
			w.WriteHeader(http.StatusOK)
		}
		defer r.Body.Close()
	} else {
		log.Print("handled PUSH request: StatusMethodNotAllowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func signalHandler(signalChan chan os.Signal, exitChan chan bool, srv *server) {
	for {
		sig := <-signalChan
		log.Printf("Handled signal: %s", sig)

		switch sig {
		case syscall.SIGHUP:
			srv.pull()
		default:
			exitChan <- true
		}
	}
}

func main() {
	var cfgDir string
	var srv server

	flag.StringVar(&cfgDir, "cfg-dir", "/opt/dns-bh/etc", "Config dir path")
	flag.StringVar(&srv.address, "address", "127.0.0.1", "Listen address")
	flag.IntVar(&srv.port, "port", 8080, "Listen port")
	flag.Parse()

	lib.ConfigInit(cfgDir)
	if !lib.ConfigLoad(&srv.cfg) {
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	exitChan := make(chan bool, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go signalHandler(signalChan, exitChan, &srv)

	http.HandleFunc("/Register", srv.handle)

	log.Printf("Starting server")

	if err := srv.start(nil); err != nil {
		log.Fatalf("Error on server.start: %s", err)
	}

	srv.pull()

	for {
		// Waiting for signal
		if status := <-exitChan; status == true {
			break
		}
	}

	log.Printf("Stopping server")

	// Close HTTP Server
	if err := srv.closer.Close(); err != nil {
		log.Fatalf("Error on server close %v", err)
	}
}
