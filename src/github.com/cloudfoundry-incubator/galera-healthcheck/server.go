package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/galera-healthcheck/healthcheck"
	"github.com/imdario/mergo"
	"github.com/pivotal-cf-experimental/service-config"
	"github.com/pivotal-golang/lager"

	_ "github.com/go-sql-driver/mysql"
)

var healthchecker *healthcheck.Healthchecker

func handler(w http.ResponseWriter, r *http.Request, logger lager.Logger) {
	result, msg := healthchecker.Check()
	if result {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	body := fmt.Sprintf("Galera Cluster Node Status: %s", msg)
	fmt.Fprint(w, body)

	logger.Debug(fmt.Sprintf("Healhcheck Response Body: %s", body))
}

func main() {
	serviceConfig := service_config.New()

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	pidFile := flags.String("pidFile", "", "Path to create a pid file when the healthcheck server has started")
	serviceConfig.AddFlags(flags)
	cf_lager.AddFlags(flags)
	flags.Parse(os.Args[1:])
	logger, _ := cf_lager.New("Galera Healthcheck")

	var config healthcheck.Config
	err := serviceConfig.Read(&config)
	if err != nil && err != service_config.NoConfigError {
		logger.Fatal("Failed to read config", err)
	}

	var defaults = healthcheck.Config{
		Host: "0.0.0.0",
		Port: 8080,
		DB: healthcheck.DBConfig{
			Host:     "0.0.0.0",
			Port:     3306,
			User:     "root",
			Password: "",
		},
		AvailableWhenDonor:    true,
		AvailableWhenReadOnly: false,
	}

	err = mergo.Merge(&config, defaults)
	if err != nil {
		logger.Fatal("Failed to merge user config with defaults", err)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port))
	if err != nil {
		logger.Fatal("Failed to open DB connection", err, lager.Data{
			"dbHost": config.DB.Host,
			"dbPort": config.DB.Port,
			"dbUser": config.DB.User,
		})
	}

	healthchecker = healthcheck.New(db, config)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, logger)
	})

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	url := fmt.Sprintf("http://%s/", address)

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			logger.Fatal("Initialization failed: GET endpoint", err, lager.Data{"url": url})
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Fatal("Initialization failed: reading response body", err, lager.Data{
				"url":         url,
				"status-code": resp.StatusCode,
			})
		}
		logger.Info(fmt.Sprintf("Initial Response: %s", body))

		if *pidFile != "" {
			// existence of pid file means the server is running
			pid := os.Getpid()
			err = ioutil.WriteFile(*pidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
			if err != nil {
				logger.Fatal("Failed to write pid file", err, lager.Data{
					"pid":     pid,
					"pidFile": *pidFile,
				})
			}
		}

		// Used by tests to deterministically know that the healthcheck is accepting incoming connections
		logger.Info("Healthcheck Started")
	}()

	server := &http.Server{Addr: address}
	server.ListenAndServe()
}
