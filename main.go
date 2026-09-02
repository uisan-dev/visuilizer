package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"visuilizer/anilist"
	"visuilizer/api"
	"visuilizer/store"
)

func main() {
	// bakemonogatariID := 5081

	client := anilist.NewClient()
	st, err := store.Open("visuilizer.db")
	if err != nil {
		log.Fatalf("Error opening database: %s", err.Error())
	}
	server := api.NewServer(client, st)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.Router(),
	}

	go func() {
		log.Println("Listening on port 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Visuilizer server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown: %v", err)
	}

	log.Println("Gracefully shut down")
}
