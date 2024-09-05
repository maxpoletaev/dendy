package main

import (
	"flag"
	"log"
	"os"

	"github.com/maxpoletaev/dendy/internal/loglevel"
	"github.com/maxpoletaev/dendy/relay"
)

type opts struct {
	addr string
}

func parseOpts() opts {
	opts := opts{}

	flag.StringVar(&opts.addr, "addr", ":1234", "address to listen on")

	flag.Parse()

	return opts
}

func main() {
	store := relay.NewInMemoryStore()
	srv := relay.NewServer(store, relay.NewIPLimiter(10))
	args := parseOpts()

	log.Default().SetFlags(0)
	log.Default().SetOutput(loglevel.New(os.Stderr, loglevel.LevelInfo))

	log.Printf("[INFO] starting relay server on %s", args.addr)

	if err := srv.Listen(args.addr); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}
}
