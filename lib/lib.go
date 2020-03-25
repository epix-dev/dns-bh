package lib

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"

	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Name     string
		Password string
		SSL      bool
	}

	SMTP struct {
		Host      string
		Port      int
		User      string
		Password  string
		Recipient string
		Sender    string
	}
}

func ConnectDb(host string, port int, user string, password string, name string, ssl bool) (*sql.DB, error) {
	var err error
	var db *sql.DB
	var sslMode string = "require"

	if !ssl {
		sslMode = "disable"
	}

	dbString := fmt.Sprintf(
		"host='%s' port=%d user='%s' password='%s' dbname='%s' sslmode='%s'",
		host, port, user, password, name, sslMode)

	if db, err = sql.Open("postgres", dbString); err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, err
}

func CheckError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func ArrayContaintsInt(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func ConfigInit(dir string) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config") // no need to include file extension
	viper.AddConfigPath(dir)      // set the path of your config file

	viper.SetDefault("db.host", "127.0.0.1")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("db.user", "dns-bh")
	viper.SetDefault("db.name", "dns-bh_development")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.ssl", true)

	viper.SetDefault("smtp.host", "127.0.0.1")
	viper.SetDefault("smtp.port", "25")
	viper.SetDefault("smtp.user", "")
	viper.SetDefault("smtp.password", "")
	viper.SetDefault("smtp.recipient", "")
	viper.SetDefault("smtp.sender", "")
}

func ConfigLoad(cfg *Config) bool {
	result := true

	err := viper.ReadInConfig()
	if err != nil {
		result = false
		fmt.Printf("Config file: %s\n", err)
	} else {
		cfg.DB.Host = viper.GetString("db.host")
		cfg.DB.Port = viper.GetInt("db.port")
		cfg.DB.User = viper.GetString("db.user")
		cfg.DB.Name = viper.GetString("db.name")
		cfg.DB.Password = viper.GetString("db.password")
		cfg.DB.SSL = viper.GetBool("db.ssl")

		cfg.SMTP.Host = viper.GetString("smtp.host")
		cfg.SMTP.Port = viper.GetInt("smtp.port")
		cfg.SMTP.User = viper.GetString("smtp.user")
		cfg.SMTP.Password = viper.GetString("smtp.password")
		cfg.SMTP.Recipient = viper.GetString("smtp.recipient")
		cfg.SMTP.Sender = viper.GetString("smtp.sender")
	}

	return result
}

func ReportChanges(cfg *Config, domains []string, subject string) {
	log.Printf("%s: %d", subject, len(domains))

	if len(domains) == 0 {
		return
	}

	mailBody := fmt.Sprintf("To: %s\r\n", cfg.SMTP.Recipient) +
		fmt.Sprintf("From: %s\r\n", cfg.SMTP.Sender) +
		fmt.Sprintf("Subject: [DNS-BH] %s\r\n", subject) +
		"\r\n" +
		fmt.Sprintf("%s:\r\n - %s \r\n", subject, strings.Join(domains, "\r\n - "))

	auth := smtp.PlainAuth(
		"",
		cfg.SMTP.User,
		cfg.SMTP.Password,
		cfg.SMTP.Host,
	)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port),
		auth,
		cfg.SMTP.Sender,
		[]string{cfg.SMTP.Recipient},
		[]byte(mailBody),
	)

	if err != nil {
		log.Printf("SMTP: %v", err)
	}
}
