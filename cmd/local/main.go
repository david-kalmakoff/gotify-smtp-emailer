package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/david-kalmakoff/gotify-smtp-mailer/testlib"
)

func main() {
	ctx := context.Background()

	log.Println("starting docker")
	binPath, err := filepath.Abs(filepath.Join("build", "gotify-smtp-emailer-linux-amd64.so"))
	if err != nil {
		log.Fatal(err)
	}

	s, err := testlib.NewDockerService(ctx, binPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("started docker")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown

	log.Println("stopping docker")
	err = s.Stop(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
