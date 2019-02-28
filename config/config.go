package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// Config struct holds all of the runtime configuration for the application
type Config struct {
	*ExporterConfig
	*APIConfig
	Repositories  string
	Organisations string
	Users         string
	APITokenEnv   string
	APITokenFile  string
	APIToken      string
	TargetURLs    []string
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	ec := ExporterConfig{
		MetricsPath:     viper.GetString("METRICS_PATH"),
		ListenPort:      viper.GetString("LISTEN_PORT"),
		LogLevel:        viper.GetString("LOG_LEVEL"),
		ApplicationName: viper.GetString("APP_NAME"),
	}

	ac := APIConfig{
		APIURL: viper.GetString("API_URL"),
	}

	repos := viper.GetString("REPOS")
	orgs := viper.GetString("ORGS")
	users := viper.GetString("USERS")
	tokenEnv := viper.GetString("GITHUB_TOKEN")
	tokenFile := viper.GetString("GITHUB_TOKEN_FILE")

	token, err := getAuth(tokenEnv, tokenFile)
	scraped, err := getScrapeURLs(ac.APIURL, repos, orgs, users)

	if err != nil {
		log.Errorf("Error initialising Configuration, Error: %v", err)
	}

	appConfig := Config{
		&ec,
		&ac,
		repos,
		orgs,
		users,
		tokenEnv,
		tokenFile,
		token,
		scraped,
	}

	return appConfig
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func getScrapeURLs(apiURL, repos, orgs, users string) ([]string, error) {

	urls := []string{}

	opts := "?&per_page=100" // Used to set the Github API to return 100 results per page (max)

	// User input validation, check that either repositories or organisations have been passed in
	if len(repos) == 0 && len(orgs) == 0 && len(users) == 0 {
		return urls, fmt.Errorf("no targets specified")
	}

	// Append repositories to the array
	if repos != "" {
		rs := strings.Split(repos, ", ")
		for _, x := range rs {
			y := fmt.Sprintf("%s/repos/%s%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	// Append github organisations to the array
	if orgs != "" {
		o := strings.Split(orgs, ", ")
		for _, x := range o {
			y := fmt.Sprintf("%s/orgs/%s/repos%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	if users != "" {
		us := strings.Split(users, ", ")
		for _, x := range us {
			y := fmt.Sprintf("%s/users/%s/repos%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	return urls, nil
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
