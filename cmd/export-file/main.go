package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"path/filepath"

	"github.com/epix-dev/dns-bh/lib"
	"golang.org/x/net/idna"

	_ "github.com/lib/pq"
)

const hazardFile = "hazard_domains.txt"
const malwareFile = "malware_domains.txt"

type ByLength []string

func (s ByLength) Len() int {
	return len(s)
}
func (s ByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

type domains struct {
	hazard  []string
	malware []string
}

func (d *domains) load(db *sql.DB) error {
	var domain string
	var selectSQL string

	selectSQL = `SELECT domain FROM hazard
				WHERE deleted_at IS NULL AND domain NOT IN (SELECT domain FROM whitelist WHERE deleted_at IS NULL) ORDER BY 1`

	if rows, err := db.Query(selectSQL); err == nil {
		for rows.Next() {
			if err = rows.Scan(&domain); err == nil {
				d.hazard = append(d.hazard, domain)
			}
		}
		rows.Close()
	} else {
		return err
	}

	selectSQL = `SELECT domain FROM malware
				WHERE deleted_at IS NULL AND domain NOT IN (SELECT domain FROM whitelist WHERE deleted_at IS NULL) ORDER BY 1`

	if rows, err := db.Query(selectSQL); err == nil {
		for rows.Next() {
			if err = rows.Scan(&domain); err == nil {
				d.malware = append(d.malware, domain)
			}
		}
		rows.Close()
	} else {
		return err
	}

	return nil
}

func (d *domains) save(outputDir string) error {
	return nil
}

func fileMD5(filePath string) (string, error) {
	var MD5String string

	file, err := os.Open(filePath)
	if err != nil {
		return MD5String, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return MD5String, err
	}

	MD5String = hex.EncodeToString(hash.Sum(nil)[:16])

	return MD5String, nil
}

func fileSave(filePath string, d []string) error {
	var err error
	var md5Before string
	var md5After string

	if md5Before, err = fileMD5(filePath); err != nil {
		return err
	}

	tmpfile, err := ioutil.TempFile(filepath.Dir(filePath), "dns-bh_")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name()) // clean up

	sort.Sort(ByLength(d))

	for _, domain := range d {
		if domainIDN, err := idna.ToASCII(domain); err == nil {

			if _, err := tmpfile.WriteString(domainIDN + "\n"); err != nil {
				return err
			}
		}
	}

	tmpfile.Chmod(0444)

	if err := tmpfile.Close(); err != nil {
		return err
	}

	if md5After, err = fileMD5(tmpfile.Name()); err != nil {
		return err
	}

	if md5Before != md5After {
		var file *os.File

		if file, err = os.Create(filepath.Join(filepath.Dir(filePath), "dns-bh.reload")); err != nil {
			return err
		}
		defer file.Close()

		file.Chmod(0644)
	}

	if err := os.Rename(tmpfile.Name(), filePath); err != nil {
		return err
	}

	return nil
}

func main() {
	var err error
	var db *sql.DB
	var dom domains

	var cfgDir string
	var outputDir string
	var cfg lib.Config

	flag.StringVar(&cfgDir, "cfg-dir", "/opt/dns-bh/etc", "Config dir path")
	flag.StringVar(&outputDir, "output-dir", "/etc/powerdns", "Output dir path")
	flag.Parse()

	lib.ConfigInit(cfgDir)
	if !lib.ConfigLoad(&cfg) {
		os.Exit(1)
	}

	db, err = lib.ConnectDb(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err = dom.load(db); err != nil {
		log.Fatalln(err)
	}

	if err = fileSave(filepath.Join(outputDir, hazardFile), dom.hazard); err != nil {
		log.Fatalln(err)
	}

	log.Printf("saved hazard domains: %d", len(dom.hazard))

	if err = fileSave(filepath.Join(outputDir, malwareFile), dom.malware); err != nil {
		log.Fatalln(err)
	}

	log.Printf("saved malware domains: %d", len(dom.malware))
}
