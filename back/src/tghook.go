package main

import (
	"encoding/json"
	"log"
	"net/http"
	"text/template"

	"qqoin.backend/storage"
)

type qTGHooker struct {
	botToken       string
	botSecretToken string
	Name           string
	WebAppUrl      string
	storage        *storage.QStorage
}

type tgHookPayload struct {
	UpdateId int64     `json:"update_id"`
	Message  tgMessage `json:"message"`
}

type tgMessage struct {
	Id   int64  `json:"message_id"`
	Text string `json:"text"`
	Chat tgChat `json:"chat"`
	User tgUser `json:"from"`
}

type tgChat struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type tgUser struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

var helloReplyTemplate = `{
"method": "sendMessage",
"chat_id": "{{.Message.Chat.Id}}",
{{if .Tap.Score}}
"text": "welcome back, {{.Message.User.Username}}!\nyou have {{.Tap.Score}} points after {{.Tap.Count}} rounds.\nlet's play more!",
{{ else }}
"text": "hello! I'm qQoin bot. Let's play a game!",
{{end}}
"reply_markup": "{\"inline_keyboard\": [[{\"text\": \"play qQoin\", \"web_app\": {\"url\": \"{{.Server.WebAppUrl}}\"}}]]}"
}`

func (s *qTGHooker) tgHookHandler(rsp http.ResponseWriter, req *http.Request) {

	// validate bot secret token
	valid := s.validateSecretToken(req)
	if !valid {
		log.Printf("bot secret token mismatch\n")
		http.Error(rsp, "", http.StatusForbidden)
		return
	}

	// decode incoming message
	var payload tgHookPayload
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		log.Printf("Error decoding incoming message: %v\n", err)
		http.Error(rsp, "invalid payload", http.StatusBadRequest)
		return
	}
	js, _ := json.Marshal(payload)
	log.Printf("Received message: %s\n", js)

	// select action based on message
	switch payload.Message.Text {
	case "/start":
		s.tgHookStartHandler(rsp, payload.Message)
	case "/hello":
		s.tgHookStartHandler(rsp, payload.Message)
	default:
		s.tgHookDefaultHandler(rsp, payload.Message)
	}
}

func (s *qTGHooker) validateSecretToken(req *http.Request) bool {
	if s.botSecretToken != "" {
		botSecretToken := req.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if botSecretToken != s.botSecretToken {
			return false
		}
	}
	return true
}

func (s *qTGHooker) tgHookStartHandler(rsp http.ResponseWriter, msg tgMessage) {
	_, dbtap := s.getUserTap(msg)
	type tmplData struct {
		Message tgMessage
		Server  qTGHooker
		Tap     storage.Tap
	}
	tmpl, _ := template.New("").Parse(helloReplyTemplate)
	rsp.Header().Set("Content-Type", "application/json")
	err := tmpl.Execute(rsp, tmplData{Message: msg, Server: *s, Tap: *dbtap})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}
}

func (s *qTGHooker) tgHookDefaultHandler(rsp http.ResponseWriter, msg tgMessage) {
	s.tgHookStartHandler(rsp, msg)
}

func (s *qTGHooker) getUserTap(msg tgMessage) (*storage.User, *storage.Tap) {
	user := storage.User{
		UID:      msg.User.Id,
		Username: msg.User.Username,
		Name:     msg.User.FirstName + " " + msg.User.LastName,
	}
	err := s.storage.CreateUpdateUser(&user)
	if err != nil {
		log.Printf("error updating user: %v\n", err)
	}
	dbtap, err := s.storage.GetTap(user.UID)
	if err != nil || dbtap == nil {
		log.Printf("no taps for user: %v\n", err)
		dbtap = &storage.Tap{Score: 0, Count: 0}
	} else {
		log.Printf("found tap: %v\n", dbtap)
	}

	return &user, dbtap

}
