package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rverst/bwof-backend/pkg/server"
	"os"
)

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	server.Run()
}
