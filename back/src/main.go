package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"qqoin.backend/storage"
)

type QQOptions struct {
	debug bool

	storageOpts storage.QSOptions

	botToken       string
	botName        string
	botSecretToken string

	webappURL        string
	webappIgnoreHash bool

	listen string
}

func getEnvs(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvb(name string, fallback bool) bool {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	value = strings.ToLower(value)
	if value == "true" || value == "1" || value == "yes" {
		return true
	}
	return false
}

func parseArgs() QQOptions {

	opts := QQOptions{

		debug: getEnvb("QQOIN_DEBUG", false),

		storageOpts: storage.QSOptions{
			StoragePath:   getEnvs("QQOIN_STORAGE_PATH", "qqoin.db"),
			StorageEngine: getEnvs("QQOIN_STORAGE_ENGINE", "sqlite"),
		},

		botToken:       getEnvs("QQOIN_BOT_TOKEN", ""),
		botName:        getEnvs("QQOIN_BOT_NAME", ""),
		botSecretToken: getEnvs("QQOIN_BOT_SECRET_TOKEN", ""),

		webappURL:        getEnvs("QQOIN_WEBAPP_URL", "https://qqoin.qq/"),
		webappIgnoreHash: getEnvb("QQOIN_WEBAPP_IGNORE_HASH", false),

		listen: getEnvs("QQOIN_LISTEN", "localhost:8765"),
	}

	flag.StringVar(&opts.storageOpts.StoragePath, "storage-path", opts.storageOpts.StoragePath,
		"database connect string")
	flag.StringVar(&opts.storageOpts.StorageEngine, "storage-engine", opts.storageOpts.StorageEngine,
		"database type -- SQLITE is the only one suppoerted at tme moment")
	flag.StringVar(&opts.botToken, "bot-token", opts.botToken,
		"TG bot access token")
	flag.StringVar(&opts.botName, "bot-name", opts.botName,
		"TG bot name")
	flag.StringVar(&opts.botSecretToken, "bot-secret", opts.botSecretToken,
		"TG bot secret key (used in input data validation)")
	flag.StringVar(&opts.webappURL, "webapp-url", opts.webappURL,
		"TMA URL")
	flag.StringVar(&opts.listen, "listen", opts.listen,
		"Back+Bot listen address")
	flag.BoolVar(&opts.webappIgnoreHash, "webapp-ignore-hash", opts.webappIgnoreHash,
		"")
	flag.BoolVar(&opts.debug, "debug", opts.debug,
		"")

	flag.Parse()

	if opts.debug {
		log.Printf("Configuration: %v\n", opts)
	}

	return opts
}

func main() {
	log.Println("qqoin is starting...")

	opts := parseArgs()

	storage := storage.QStorage{Opts: &opts.storageOpts}
	storage.Open()
	storage.Migrate()
	storage.Prepare()

	hooker := qTGHooker{
		Opts:    &opts,
		storage: &storage,
	}

	backer := qWebAppBack{
		Opts:    &opts,
		storage: &storage,
	}

	err := run(&opts, &hooker, &backer)
	log.Printf("Error: %s\n", err.Error())

	storage.Close()

}

func run(opts *QQOptions, hooker *qTGHooker, backer *qWebAppBack) error {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping/", pingHandler)
	mux.HandleFunc("POST /tghook/", hooker.tgHookHandler)
	mux.HandleFunc("GET /taps/", backer.tapsHandler)
	mux.HandleFunc("POST /taps/", backer.tapsHandler)

	handler := Logging(mux, opts)

	log.Printf("listening to: %v", opts.listen)
	err := http.ListenAndServe(opts.listen, handler)

	return err

}
