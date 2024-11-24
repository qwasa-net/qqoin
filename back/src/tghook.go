package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"

	"qqoin.backend/storage"
)

type qTGHooker struct {
	Opts    *QQOptions
	storage *storage.QStorage
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

func (s *qTGHooker) tgHookHandler(rsp http.ResponseWriter, req *http.Request) {

	// validate bot secret token
	valid := s.validateSecretToken(req)
	if !valid {
		log.Printf("bot secret token mismatch\n")
		if !s.Opts.validationIgnore {
			http.Error(rsp, "", http.StatusForbidden)
			return
		}
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
	log.Printf("TGHook message: %s\n", js)

	// select action based on message
	if looksLikeTONAddress(payload.Message.Text) {
		log.Printf("looks like TON address: %s\n", payload.Message.Text)
		s.tgHookWalletAddressHandler(rsp, payload.Message)
	} else if payload.Message.Text == "/start" {
		s.tgHookStartHandler(rsp, payload.Message)
	} else if payload.Message.Text == "/qlaim" {
		s.tgHookQlaimHandler(rsp, payload.Message)
	} else if payload.Message.Text == "/admin" {
		s.tgHookAdminHandler(rsp, payload.Message)
	} else {
		s.tgHookDefaultHandler(rsp, payload.Message)
	}
}

func (s *qTGHooker) validateSecretToken(req *http.Request) bool {
	if s.Opts.botSecretToken != "" {
		botSecretToken := req.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if botSecretToken != s.Opts.botSecretToken {
			return false
		}
	}
	return true
}

var helloReplyTemplate = `{
	"method": "sendMessage",
	"chat_id": "{{.Message.Chat.Id}}",
	{{if .Tap.Score}}
	"text": "welcome back, {{if .Message.User.Username}}{{.Message.User.Username}}{{else}}#{{.Message.Chat.Id}}{{end}}!\nyou have {{.Tap.Score}} points after {{.Tap.Count}} rounds.\nlet's play more!",
	{{ else }}
	"text": "hello! I'm qQoin bot. Let's play a game!",
	{{end}}
	"reply_markup": "{\"inline_keyboard\": [[{\"text\": \"play qQoin\", \"web_app\": {\"url\": \"{{.WebAppUrl}}\"}}]]}"
}`

func (s *qTGHooker) tgHookStartHandler(rsp http.ResponseWriter, msg tgMessage) {
	_, dbtap := s.getUserTapData(msg)
	type tmplData struct {
		Message   tgMessage
		WebAppUrl string
		Tap       storage.Tap
	}
	tmpl, _ := template.New("").Parse(helloReplyTemplate)
	rsp.Header().Set("Content-Type", "application/json")
	err := tmpl.Execute(rsp, tmplData{Message: msg, WebAppUrl: s.Opts.webappURL, Tap: *dbtap})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}
}

// var qlaimReplyTemplate = `{
// 	"method": "sendMessage",
// 	"chat_id": "{{.Message.Chat.Id}}",
// 	"text": "{{if .Message.User.Username}}{{.Message.User.Username}}{{else}}#{{.Message.Chat.Id}}{{end}}, send your wallet address to get your personal qQoken — an item from the very limited NFT collection!!\n\n"
// }`

var qlaimReplyTemplate = `{
	"method": "sendMessage",
	"chat_id": "{{.Message.Chat.Id}}",
	"text": "{{.Message.User.Username}}, qQoken qlaim time is over!!\n\nYou can still play qQoin! /start and /play now!!\n\n"
}`

func (s *qTGHooker) tgHookQlaimHandler(rsp http.ResponseWriter, msg tgMessage) {
	s.getUserTapData(msg) // create/update user
	var qqoken *storage.QQoken
	qqoken, _ = s.storage.GetQqoken(msg.User.Id)
	if qqoken != nil && qqoken.Qqoken_id != "" {
		log.Printf("qqoken already set: %v\n", qqoken)
		s.tgHookQqokenReadyHandler(rsp, msg, qqoken)
		return
	}
	type tmplData struct {
		Message tgMessage
	}
	tmpl, _ := template.New("").Parse(qlaimReplyTemplate)
	rsp.Header().Set("Content-Type", "application/json")
	err := tmpl.Execute(rsp, tmplData{Message: msg})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}
}

func looksLikeTONAddress(line string) bool {
	// User-friendly address --
	// 36 bytes, encoded with base64 or base64url --
	// 48 non-spaced characters.

	line = strings.TrimSpace(line)
	if len(line) == 48 {
		var bline []byte
		var err error
		bline, err = base64.URLEncoding.DecodeString(line)
		if err != nil {
			bline, err = base64.StdEncoding.DecodeString(line)
		}
		if err == nil {
			if len(bline) == 32 || len(bline) == 36 {
				return true
			}
		}
	}
	return false
}

var qlaimWalletSetReplyTemplate = `{
	"method": "sendMessage",
	"chat_id": "{{.Message.Chat.Id}}",
	"text": "{{if .Message.User.Username}}{{.Message.User.Username}}{{else}}#{{.Message.Chat.Id}}{{end}}, your wallet address is set to: {{.Message.Text}}\n\nYou will be notified when your personal qQoken is ready!\n\n",
    "reply_markup": "{\"inline_keyboard\": [[{\"text\": \"play qQoin\", \"web_app\": {\"url\": \"{{.WebAppUrl}}\"}}]]}"
}`

func (s *qTGHooker) tgHookWalletAddressHandler(rsp http.ResponseWriter, msg tgMessage) {
	var err error
	var qqoken *storage.QQoken
	var tmpl *template.Template

	qqoken, _ = s.storage.GetQqoken(msg.User.Id)
	if qqoken != nil && qqoken.Qqoken_id != "" {
		log.Printf("qqoken already set: %v\n", qqoken)
		s.tgHookQqokenReadyHandler(rsp, msg, qqoken)
		return
	}

	wallet_addr := strings.TrimSpace(msg.Text)
	qqoken = &storage.QQoken{
		UID:         msg.User.Id,
		Wallet_addr: wallet_addr,
	}
	err = s.storage.CreateUpdateQqoken(qqoken)
	if err != nil {
		log.Printf("error upserting qqoken: %v\n", err)
	}

	tmpl, _ = template.New("").Parse(qlaimWalletSetReplyTemplate)
	type tmplData struct {
		Message   tgMessage
		QQoken    storage.QQoken
		WebAppUrl string
	}
	rsp.Header().Set("Content-Type", "application/json")
	err = tmpl.Execute(rsp, tmplData{Message: msg, WebAppUrl: s.Opts.webappURL, QQoken: *qqoken})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}

}

var qlaimQqokenReadyReplyTemplate = `{
	"method": "sendMessage",
	"chat_id": "{{.Message.Chat.Id}}",
	"text": "{{if .Message.User.Username}}{{.Message.User.Username}}{{else}}#{{.Message.Chat.Id}}{{end}}, qQoken №{{.QQoken.Qqoken_id}} '{{.QQoken.Qqoken_addr}}' is already created for you!\n\ncheck here: https://tonviewer.com/{{.QQoken.Qqoken_addr}}?section=nft\n\n",
	"reply_markup": "{\"inline_keyboard\": [[{\"text\": \"tonviewer\", \"web_app\": {\"url\": \"https://tonviewer.com/{{.QQoken.Qqoken_addr}}\"}}]]}"
}`

func (s *qTGHooker) tgHookQqokenReadyHandler(rsp http.ResponseWriter, msg tgMessage, qqoken *storage.QQoken) {
	var err error
	var tmpl *template.Template
	if qqoken == nil || qqoken.Qqoken_id == "" {
		log.Printf("qqoken not set: %v\n", qqoken)
		return
	}
	log.Printf("qqoken already set: %v\n", qqoken)
	tmpl, _ = template.New("").Parse(qlaimQqokenReadyReplyTemplate)
	type tmplData struct {
		Message   tgMessage
		QQoken    storage.QQoken
		WebAppUrl string
	}

	rsp.Header().Set("Content-Type", "application/json")
	err = tmpl.Execute(rsp, tmplData{Message: msg, WebAppUrl: s.Opts.webappURL, QQoken: *qqoken})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}

}

func (s *qTGHooker) tgHookDefaultHandler(rsp http.ResponseWriter, msg tgMessage) {
	s.tgHookStartHandler(rsp, msg)
}

func (s *qTGHooker) getUserTapData(msg tgMessage) (*storage.User, *storage.Tap) {
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

var usersListTemplate = `{{range .Users}}- {{.UID}} | {{.Username}} ({{.Name}})\n{{end}}`
var tapsListTemplate = `{{range .Taps}}- {{.UID}} | {{.Score}}/{{.Count}}\n{{end}}`
var qqokensListTemplate = `{{range .Qqokens}}- {{.UID}} | {{.Wallet_addr}}/{{.Qqoken_addr}}\n{{end}}`
var adminReplyTemplate = `{
	"method": "sendMessage",
	"chat_id": "{{.Message.Chat.Id}}",
	"text": "hello, {{.Message.User.Username}}!\n` +
	`\n== recent users ==\n` +
	usersListTemplate +
	`\n== top taps ==\n` +
	tapsListTemplate +
	`\n== qqokens ==\n` +
	qqokensListTemplate + `"
}`

func (s *qTGHooker) tgHookAdminHandler(rsp http.ResponseWriter, msg tgMessage) {
	uid := msg.User.Id
	if uid != s.Opts.botAdminUser {
		log.Printf("access blocked for non-admin user: %v\n", uid)
		s.tgHookDefaultHandler(rsp, msg)
		return
	}
	users, _ := s.storage.GetAllUsers(100)
	taps, _ := s.storage.GetAllTaps(100)
	qqokens, _ := s.storage.GetAllQqokens(100)
	type tmplData struct {
		Message tgMessage
		Users   []storage.User
		Taps    []storage.Tap
		Qqokens []storage.QQoken
	}
	tmpl, _ := template.New("").Parse(adminReplyTemplate)
	rsp.Header().Set("Content-Type", "application/json")
	err := tmpl.Execute(rsp, tmplData{Message: msg, Users: users, Taps: taps, Qqokens: qqokens})
	if err != nil {
		log.Printf("error executing template: %v\n", err)
	}
}
