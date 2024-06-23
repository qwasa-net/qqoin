package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"qqoin.backend/ledger"
	"qqoin.backend/storage"
)

type QQOptions struct {
	debug            bool
	validationIgnore bool

	storageOpts storage.QSOptions

	ledgerOpts ledger.LedgerOptions

	botToken       string
	botName        string
	botSecretToken string
	botAdminUser   int64

	webappURL string

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

func getEnvi(name string, fallback int64) int64 {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return i
}

func parseArgs() QQOptions {

	opts := QQOptions{

		debug: getEnvb("QQOIN_DEBUG", false),

		storageOpts: storage.QSOptions{
			StoragePath:   getEnvs("QQOIN_STORAGE_PATH", "qqoin.db"),
			StorageEngine: getEnvs("QQOIN_STORAGE_ENGINE", "sqlite"),
		},

		ledgerOpts: ledger.LedgerOptions{
			Path:       getEnvs("QQOIN_LEDGER_PATH", ""),
			PathTs:     getEnvb("QQOIN_LEDGER_PATHTS", false),
			FlushCount: getEnvi("QQOIN_LEDGER_FLUSH_COUNT", 100),
			MaxRecords: getEnvi("QQOIN_LEDGER_RECORDS_MAX", 10000),
		},

		botToken:       getEnvs("QQOIN_BOT_TOKEN", ""),
		botName:        getEnvs("QQOIN_BOT_NAME", ""),
		botSecretToken: getEnvs("QQOIN_BOT_SECRET_TOKEN", ""),
		botAdminUser:   getEnvi("QQOIN_BOT_ADMIN_USER", 0),

		webappURL:        getEnvs("QQOIN_WEBAPP_URL", "https://qqoin.qq/"),
		validationIgnore: getEnvb("QQOIN_VALIDATION_IGNORE", false),

		listen: getEnvs("QQOIN_LISTEN", "localhost:8765"),
	}

	flag.StringVar(&opts.storageOpts.StoragePath, "storage-path", opts.storageOpts.StoragePath,
		"database connect string")
	flag.StringVar(&opts.storageOpts.StorageEngine, "storage-engine", opts.storageOpts.StorageEngine,
		"database type -- SQLITE is the only one supported at tme moment")
	flag.StringVar(&opts.botToken, "bot-token", opts.botToken,
		"TG bot access token")
	flag.StringVar(&opts.botName, "bot-name", opts.botName,
		"TG bot name")
	flag.StringVar(&opts.botSecretToken, "bot-secret", opts.botSecretToken,
		"TG bot secret key (used in input data validation)")
	flag.Int64Var(&opts.botAdminUser, "bot-admin", opts.botAdminUser,
		"TG bot admin user UID")
	flag.StringVar(&opts.webappURL, "webapp-url", opts.webappURL,
		"TMA URL")
	flag.StringVar(&opts.listen, "listen", opts.listen,
		"Back+Bot listen address")
	flag.BoolVar(&opts.validationIgnore, "ignore-validation", opts.validationIgnore,
		"")
	flag.StringVar(&opts.ledgerOpts.Path, "ledger-file", opts.ledgerOpts.Path,
		"ledger dump file name")
	flag.BoolVar(&opts.ledgerOpts.PathTs, "ledger-file-ts", opts.ledgerOpts.PathTs,
		"add timestamp to ledger dump file name")
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

	// db storage
	strg := storage.NewQStorage(&opts.storageOpts)

	// ledger logger
	var ldgr *ledger.Ledger
	if opts.ledgerOpts.Path != "" {
		ldgr = ledger.NewLedger(&opts.ledgerOpts)
		if ldgr != nil {
			go ldgr.Start()
		}
	}

	// tg hook handler
	hooker := qTGHooker{
		Opts:    &opts,
		storage: strg,
	}

	// webapp handler
	backer := qWebAppBack{
		Opts:    &opts,
		storage: strg,
		ledger:  ldgr,
	}

	err := run(&opts, &hooker, &backer)
	log.Printf("error: %s\n", err.Error())

	log.Println("qqoin is shutting down...")
	strg.Close()
	if ldgr != nil {
		ldgr.Close()
	}

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
