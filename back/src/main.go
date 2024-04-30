package main

import (
	"log"
	"net/http"
	"os"

	"qqoin.backend/storage"
)

func main() {
	log.Println("qqoin is starting...")

	storage := storage.QStorage{}
	storage.Open(os.Getenv("STORAGE_PATH"), os.Getenv("STORAGE_ENGINE"))
	storage.Migrate()

	hooker := qTGHooker{
		botToken:       os.Getenv("BOT_TOKEN"),
		Name:           os.Getenv("BOT_NAME"),
		WebAppUrl:      os.Getenv("WEBAPP_URL"),
		botSecretToken: os.Getenv("BOT_SECRET_TOKEN"),
		storage:        &storage,
	}
	backer := qWebAppBack{
		botToken: os.Getenv("BOT_TOKEN"),
		storage:  &storage,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping/", pingHandler)
	mux.HandleFunc("POST /tghook/", hooker.tgHookHandler)
	mux.HandleFunc("GET /taps/", backer.tapsHandler)
	mux.HandleFunc("POST /taps/", backer.tapsHandler)

	handler := Logging(mux)

	listen_host := os.Getenv("QQOIN_LISTEN_HOST")
	if listen_host == "" {
		listen_host = "localhost"
	}
	listen_port := os.Getenv("QQOIN_LISTEN_PORT")
	if listen_port == "" {
		listen_port = "8765"
	}
	listen := listen_host + ":" + listen_port
	log.Printf("listening to: %v", listen)

	err := http.ListenAndServe(listen, handler)
	log.Printf("Error: %s\n", err.Error())
}
