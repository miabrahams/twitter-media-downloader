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

func (d *Downloader) download(tweet *twitterscraper.Tweet, url, fileType, output, dwnType string) {
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
	if d.config.UrlOnly {
		fmt.Println(url)
		return
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")
	resp, err := d.httpClient.Do(req)
	if err != nil {
		fmt.Println("Error downloading:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error: status code", resp.StatusCode)
		return
	}

	var filePath string
	if dwnType == "user" {
		filePath = filepath.Join(output, fileType, name)
		if d.config.Update {
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				fmt.Println(name + ": already exists")
				return
			}
		}
		if fileType == "rtimg" || fileType == "rtvideo" {
			filePath = filepath.Join(output, strings.TrimPrefix(fileType, "rt"), "RE-"+name)
		}
	} else {
		filePath = filepath.Join(output, name)
		if d.config.Update {
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				fmt.Println("File exists")
				return
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Downloaded " + name)
}
