package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type IPList struct {
	dbFile     string
	outputFile string
	quarantine int
	list       []string
}

func (acl *IPList) readSTDIN() {
	var expr = "^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/[0-9]{1,2}$"
	var validID = regexp.MustCompile(expr)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if validID.MatchString(line) {
			acl.list = append(acl.list, line)
		}
	}
}

func (acl *IPList) updateDB(db *sql.DB) {
	insertSQL := "INSERT OR REPLACE INTO networks (ip, seen) VALUES ($1, $2)"

	updateStm, err := db.Prepare(insertSQL)

	for _, ip := range acl.list {
		_, err = updateStm.Exec(ip, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Printf("%q: %s\n", err, insertSQL)
			return
		}
	}
}

func (acl *IPList) generateACL(db *sql.DB) error {
	var ip string
	var seen string

	file, err := os.Create(acl.outputFile)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	oldest := time.Now().AddDate(0, 0, acl.quarantine*-1).Format("2006-01-02 15:04:05")

	selectSQL := "SELECT ip, seen FROM networks WHERE date(seen) > date($1) ORDER BY 2,1"

	if rows, err := db.Query(selectSQL, oldest); err == nil {

		for rows.Next() {
			if err = rows.Scan(&ip, &seen); err == nil {
				file.WriteString(fmt.Sprintf("%s\t#%s\n", ip, seen))
				// fmt.Println(ip)
			}
		}

		rows.Close()
	} else {
		fmt.Println(err)
		return err
	}

	return nil
}

func main() {
	var acl IPList

	flag.StringVar(&acl.dbFile, "db-file", "/opt/dns-bh/etc/acl.db", "Path to ACL database file")
	flag.StringVar(&acl.outputFile, "acl-file", "/opt/dns-bh/etc/acl.txt", "Path to ACL output file")
	flag.IntVar(&acl.quarantine, "quarantine", 30, "Number of days of quarantine for networks that doesn't exists")
	flag.Parse()

	db, err := sql.Open("sqlite3", acl.dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	createSQL := "CREATE TABLE IF NOT EXISTS networks (ip TEXT, seen TEXT)"

	_, err = db.Exec(createSQL)
	if err != nil {
		log.Printf("%q: %s\n", err, createSQL)
		return
	}

	indexSQL := "CREATE UNIQUE INDEX IF NOT EXISTS networks_idx ON networks(ip)"

	_, err = db.Exec(indexSQL)
	if err != nil {
		log.Printf("%q: %s\n", err, indexSQL)
		return
	}

	acl.readSTDIN()
	acl.updateDB(db)
	acl.generateACL(db)
}
