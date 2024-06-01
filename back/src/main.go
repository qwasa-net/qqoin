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
	storage.Open(os.Getenv("QQOIN_STORAGE_PATH"), os.Getenv("QQOIN_STORAGE_ENGINE"))
	storage.Migrate()
	storage.Prepare()

	hooker := qTGHooker{
		botToken:       os.Getenv("QQOIN_BOT_TOKEN"),
		Name:           os.Getenv("QQOIN_BOT_NAME"),
		WebAppUrl:      os.Getenv("QQOIN_WEBAPP_URL"),
		botSecretToken: os.Getenv("QQOIN_BOT_SECRET_TOKEN"),
		storage:        &storage,
	}

	backer := qWebAppBack{
		botToken: os.Getenv("QQOIN_BOT_TOKEN"),
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

	storage.Close()
}
