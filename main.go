package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MikeFilimonov/go-pwc/config"
	"github.com/MikeFilimonov/go-pwc/core/links"
	"github.com/MikeFilimonov/go-pwc/core/site"
)

func main() {

	params, err := config.Load(os.Args[1:], os.Stdout)

	if errors.Is(err, flag.ErrHelp) {
		os.Exit(0)
	}

	if err != nil {
		os.Exit(2)
	}

	input, err := links.FixURL(params.InitialURL)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	backgroundCtx := context.Background()
	ctx, stopSignals := signal.NotifyContext(backgroundCtx, os.Interrupt,
		syscall.SIGTERM)
	defer stopSignals()

	stats, err := site.Crawl(ctx, input, params.MaxDepth)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Println(err)
		os.Exit(1)
	}

	log.Printf("results:\n -links found : %v,\n -pages count: %v,\n -total bytes: %v,\n  params: %v",
		stats.LinksFound, stats.PageCount, stats.TotalBytes, params)

}
