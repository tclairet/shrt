package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const port = 4000

func main() {
	db, err := newPostgres(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	api := newAPI(db)

	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port), Handler: api.Routes()}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
