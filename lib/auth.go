// auth.go

package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	twitterscraper "github.com/imperatrona/twitter-scraper"
	"golang.org/x/term"
)

const cookieFile = "twmd_cookies.json"

type Authenticator struct {
	scraper *twitterscraper.Scraper
	config  *Config
}

func NewAuthenticator(scraper *twitterscraper.Scraper, config *Config) *Authenticator {
	return &Authenticator{
		scraper: scraper,
		config:  config,
	}
}

func (a *Authenticator) Login() error {
	if a.tryLoadCookies() {
		return nil
	}
	return a.performLogin()
}

func (a *Authenticator) tryLoadCookies() bool {
	cookies, err := a.loadCookiesFromFile()
	if err != nil {
		return false
	}

	a.scraper.SetCookies(cookies)
	if a.scraper.IsLoggedIn() {
		fmt.Println("Logged in using saved cookies")
		return true
	}
	return false
}

func (a *Authenticator) loadCookiesFromFile() ([]*http.Cookie, error) {
	f, err := os.Open(cookieFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cookies []*http.Cookie
	if err := json.NewDecoder(f).Decode(&cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}

func (a *Authenticator) performLogin() error {
	for {
		credentials, err := a.getCredentials()
		if err != nil {
			return err
		}

		if err := a.scraper.Login(credentials.username, credentials.password, credentials.code); err != nil {
			return err
		}

		if !a.scraper.IsLoggedIn() {
			fmt.Println("Bad user/pass")
			continue
		}

		if err := a.saveCookies(); err != nil {
			return err
		}

		fmt.Println("Logged in successfully")
		return nil
	}
}

type credentials struct {
	username, password, code string
}

func (a *Authenticator) getCredentials() (credentials, error) {
	var cred credentials
	fmt.Print("username: ")
	fmt.Scanln(&cred.username)

	fmt.Print("password: ")
	if a.config.Loginp != "" {
		fmt.Scanln(&cred.password)
	} else {
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return credentials{}, err
		}
		fmt.Println()
		cred.password = string(password)
	}

	if a.config.Twofa {
		fmt.Print("two-factor: ")
		fmt.Scanln(&cred.code)
		fmt.Println()
	}

	return cred, nil
}

func (a *Authenticator) saveCookies() error {
	cookies := a.scraper.GetCookies()
	js, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	return os.WriteFile(cookieFile, js, 0666)
}
