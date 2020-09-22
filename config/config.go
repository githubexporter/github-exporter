package config

import (
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"

	"os"

	cfg "github.com/infinityworks/go-common/config"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	Config        *cfg.BaseConfig
	APIURL        string
	Repositories  []string
	Organisations []string
	Users         []string
	APITokenEnv   string
	APITokenFile  string
	APIToken      string
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	ac := cfg.Init()
	tokenEnv := os.Getenv("GITHUB_TOKEN")
	tokenFile := os.Getenv("GITHUB_TOKEN_FILE")
	token, err := getAuth(tokenEnv, tokenFile)

	if err != nil {
		log.Errorf("Error initialising Configuration, Error: %v", err)
	}

	appConfig := Config{
		Config:        &ac,
		APIURL:        cfg.GetEnv("API_URL", "https://api.github.com"),
		Repositories:  strings.Split(os.Getenv("REPOS"), ","),
		Organisations: strings.Split(os.Getenv("ORGS"), ","),
		Users:         strings.Split(os.Getenv("USERS"), ","),
		APITokenEnv:   os.Getenv("GITHUB_TOKEN"),
		APITokenFile:  tokenFile,
		APIToken:      token,
	}

	return appConfig
}

// getAuth returns oauth2 token as string for usage in http.request
func getAuth(token string, tokenFile string) (string, error) {

	if token != "" {
		return token, nil
	} else if tokenFile != "" {
		b, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), err

	}

	return "", nil
}
