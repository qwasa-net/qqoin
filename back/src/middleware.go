package main

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		realIp := req.Header.Get("X-Real-IP")
		if realIp == "" {
			realIp = req.RemoteAddr
		}
		start := time.Now()
		next.ServeHTTP(rsp, req)
		log.Printf("[%s] %s %s %s", realIp, req.Method, req.RequestURI, time.Since(start))
	})
}
