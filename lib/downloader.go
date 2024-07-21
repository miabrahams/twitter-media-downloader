package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

type Downloader struct {
	config     *Config
	httpClient HTTPClient
}

func NewDownloader(cfg *Config, httpClient HTTPClient) *Downloader {
	return &Downloader{
		httpClient: httpClient,
		config:     cfg,
	}
}

func (d *Downloader) downloadVideos(tweet *twitterscraper.Tweet) error {
	wg := sync.WaitGroup{}
	for _, video := range tweet.Videos {
		wg.Add(1)
		go func(v twitterscraper.Video) {
			defer wg.Done()
			url := strings.Split(v.URL, "?")[0]
			d.download(tweet, url, "video", d.config.OutputDir, "user")
		}(video)
	}
	wg.Wait()
	return nil
}

func (d *Downloader) downloadPhotos(tweet *twitterscraper.Tweet) error {
	var wg sync.WaitGroup
	for _, photo := range tweet.Photos {
		wg.Add(1)
		go func(p twitterscraper.Photo) {
			defer wg.Done()
			if !strings.Contains(p.URL, "video_thumb/") {
				url := p.URL
				if d.config.Size == "orig" || d.config.Size == "small" {
					url += "?name=" + d.config.Size
				}
				d.download(tweet, url, "img", d.config.OutputDir, "user")
			}
		}(photo)
	}
	wg.Wait()
	return nil
}

func (d *Downloader) download(tweet *twitterscraper.Tweet, url, fileType, output, dwnType string) error {
	name := d.generateFileName(tweet, url)

	if d.config.UrlOnly {
		fmt.Println(url)
		return nil
	}

	resp, err := d.makeRequest(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	filePath, err := d.determineFilePath(output, fileType, name, dwnType)
	if err != nil {
		return err
	}

	if err := d.saveFile(filePath, resp.Body); err != nil {
		return err
	}

	fmt.Println("Downloaded " + name)
	return nil
}

func (d *Downloader) generateFileName(tweet *twitterscraper.Tweet, url string) string {
	segments := strings.Split(url, "/")
	name := segments[len(segments)-1]

	re := regexp.MustCompile(`name=`)
	if re.MatchString(name) {
		segments = strings.Split(name, "?")
		name = segments[len(segments)-2]
	}

	if d.config.Format != "" {
		nameFormat, _ := FormatFileName(tweet, d.config.Format, d.config.Datefmt)
		name = nameFormat + "_" + name
	}

	return name
}

func (d *Downloader) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	return resp, nil
}

func (d *Downloader) determineFilePath(output, fileType, name, dwnType string) (string, error) {
	var filePath string
	if dwnType == "user" {
		filePath = filepath.Join(output, fileType, name)
		if d.config.Update {
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				return "", fmt.Errorf("%s: already exists", name)
			}
		}
		if fileType == "rtimg" || fileType == "rtvideo" {
			filePath = filepath.Join(output, strings.TrimPrefix(fileType, "rt"), "RE-"+name)
		}
	} else {
		filePath = filepath.Join(output, name)
		if d.config.Update {
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				return "", fmt.Errorf("file exists")
			}
		}
	}
	return filePath, nil
}

func (d *Downloader) saveFile(filePath string, content io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	if _, err = io.Copy(f, content); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
