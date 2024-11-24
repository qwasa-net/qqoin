package main

import (
	"log"
	"net/http"
	"text/template"

	"qqoin.backend/storage"
)

type qQoken struct {
	Opts    *QQOptions
	storage *storage.QStorage
}

var qqokenMetaTmplt = `{
  "name": "qqoken",
  "description": "qqoin qqoken"
}
`

func (s *qQoken) qQokenHandler(rsp http.ResponseWriter, req *http.Request) {

	type tmplData struct {
	}

	tmpl, _ := template.New("").Parse(qqokenMetaTmplt)
	rsp.Header().Set("Content-Type", "application/json")
	err := tmpl.Execute(rsp, tmplData{})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}

}
