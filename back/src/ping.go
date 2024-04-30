package main

import (
	"net/http"
)

func pingHandler(rsp http.ResponseWriter, req *http.Request) {
	rsp.Write([]byte("pong"))
}
