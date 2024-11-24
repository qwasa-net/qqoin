package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func runServer(opts *QQOptions, hooker *qTGHooker, backer *qWebAppBack, qqoken *qQoken) error {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping/", pingHandler)
	mux.HandleFunc("POST /tghook/", hooker.tgHookHandler)
	mux.HandleFunc("GET /taps/", backer.tapsHandler)
	mux.HandleFunc("POST /taps/", backer.tapsHandler)
	mux.HandleFunc("GET /qqoken/", qqoken.qQokenHandler)

	handler := Logging(mux, opts)

	server := http.Server{
		Addr:    opts.listen,
		Handler: handler,
	}

	shutdownBySignal(&server, []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP})

	log.Printf("listening to: %v", opts.listen)
	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}

	return err

}

func shutdownBySignal(server *http.Server, sigs []os.Signal) {
	sigsChan := make(chan os.Signal, 1)
	go func() {
		s := <-sigsChan
		log.Printf("got signal: [%s], closing http server …\n", s)
		server.Shutdown(context.Background())
	}()
	for _, sig := range sigs {
		signal.Notify(sigsChan, sig)
	}
}
