package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	twitterscraper "github.com/imperatrona/twitter-scraper"
	"golang.org/x/term"
)

/*
type ScrapeRunner interface {
	GetTweets(ctx context.Context, username string, count int) ([]Tweet, error)
	GetTweet(ctx context.Context, id string) (*Tweet, error)
}
*/

type ScrapeRunner struct {
	cfg        *Config
	httpClient HTTPClient
	scraper    *twitterscraper.Scraper
	downloader *Downloader
}

func NewScraper(config *Config, httpClient HTTPClient) *ScrapeRunner {
	scraper := twitterscraper.New()
	scraper.WithReplies(true)
	// scraper.SetProxy(proxy)

	downloader := NewDownloader(config, httpClient)
	return &ScrapeRunner{
		cfg:        config,
		httpClient: httpClient,
		scraper:    scraper,
		downloader: downloader,
	}
}

func (s *ScrapeRunner) Run() error {

	if s.cfg.Login != "" || s.cfg.Loginp != "" {
		if err := s.login(); err != nil {
			return err
		}
	}

	if s.cfg.TweetID != "" {
		return s.RunSingleTweet(s.cfg.TweetID)
	}

	if s.cfg.User != "" {
		return s.RunUserTweets()
	}

	return errors.New("no user or tweet id specified")
}

func (s *ScrapeRunner) RunSingleTweet(id string) error {
	tweet, err := s.scraper.GetTweet(id)
	if err != nil {
		return err
	}
	if tweet == nil {
		return errors.New("error retrieving tweet")
	}

	return s.DownloadTweet(tweet)
}

func (s *ScrapeRunner) RunUserTweets() error {
	wg := sync.WaitGroup{}
	for tweet := range s.scraper.GetTweets(context.Background(), s.cfg.User, s.cfg.NumberOfTweets) {
		if tweet.Error != nil {
			return tweet.Error
		}
		go s.DownloadTweet(&tweet.Tweet)
	}
	wg.Wait()
	return nil
}

func (s *ScrapeRunner) DownloadTweet(tweet *twitterscraper.Tweet) error {
	if tweet.IsRetweet && (!s.cfg.Retweets) {
		return nil
	}
	if s.cfg.RetweetOnly && !tweet.IsRetweet {
		return nil
	}

	if s.cfg.Videos {
		err := s.downloader.downloadVideos(tweet)
		if err != nil {
			return err
		}
	}
	if s.cfg.Images {
		err := s.downloader.downloadPhotos(tweet)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ScrapeRunner) login() error {
	if _, err := os.Stat("twmd_cookies.json"); errors.Is(err, os.ErrNotExist) {
		return s.askPass()
	}

	f, err := os.Open("twmd_cookies.json")
	if err != nil {
		return err
	}
	defer f.Close()

	var cookies []*http.Cookie
	if err := json.NewDecoder(f).Decode(&cookies); err != nil {
		return err
	}

	s.scraper.SetCookies(cookies)

	if !s.scraper.IsLoggedIn() {
		return s.askPass()
	}

	fmt.Println("Logged in")
	return nil
}

func (s *ScrapeRunner) askPass() error {
	for {
		var username, pass string
		fmt.Print("username: ")
		fmt.Scanln(&username)
		fmt.Print("password: ")
		if s.cfg.Loginp != "" {
			fmt.Scanln(&pass)
		} else {
			password, _ := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			pass = string(password)
		}

		var code string
		if s.cfg.Twofa {
			fmt.Print("two-factor: ")
			fmt.Scanln(&code)
			fmt.Println()
		}

		if err := s.scraper.Login(username, pass, code); err != nil {
			return err
		}

		if !s.scraper.IsLoggedIn() {
			fmt.Println("Bad user/pass")
			continue
		}

		cookies := s.scraper.GetCookies()
		js, _ := json.Marshal(cookies)
		f, err := os.OpenFile("twmd_cookies.json", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.Write(js); err != nil {
			return err
		}
		break
	}
	return nil
}
