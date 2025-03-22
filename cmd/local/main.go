package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/david-kalmakoff/gotify-smtp-emailer/testlib"
)

func main() {
	ctx := context.Background()

	log.Println("starting docker")
	filename := fmt.Sprintf("gotify-smtp-emailer-linux-amd64%s.so", os.Getenv("FILE_SUFFIX"))
	binPath, err := filepath.Abs(filepath.Join("build", filename))
	if err != nil {
		log.Fatal(err)
	}

	s, err := testlib.NewDockerService(ctx, binPath, "ENV", "development")
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
