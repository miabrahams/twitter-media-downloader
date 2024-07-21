package main

import (
	"log"
	"twmd/lib"
)

func main() {
	cfg := lib.Configure()

	httpClient := lib.NewHTTPClient(cfg.Proxy)
	twitterScraper := lib.NewScraper(cfg, httpClient)

	if err := twitterScraper.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
