package main

import (
	"github.com/tonradar/ton-dice-web-fetcher/config"
	"github.com/tonradar/ton-dice-web-fetcher/fetcher"
)

func main() {
	cfg := config.GetConfig()

	service := fetcher.NewFetcher(&cfg)
	service.Start()
}
