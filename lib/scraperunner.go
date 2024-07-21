package lib

import (
	"context"
	"errors"
	"sync"

	twitterscraper "github.com/imperatrona/twitter-scraper"
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
		auth := NewAuthenticator(s.scraper, s.cfg)
		if err := auth.Login(); err != nil {
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
