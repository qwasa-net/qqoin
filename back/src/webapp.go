package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"qqoin.backend/storage"
)

type qWebAppBack struct {
	Opts    *QQOptions
	storage *storage.QStorage
}

type qWebAppTapInput struct {
	Init   string  `json:"init"`
	Energy int64   `json:"e"`
	Score  int64   `json:"s"`
	XYZ    []int64 `json:"xyz"`
	UID    int64   `json:"uid"`
}

type initParamUser struct {
	Id        int64  `json:"id"`
	Username  string `json:"username"`
	IsPremium bool   `json:"is_premium"`
}

func (s *qWebAppBack) tapsHandler(rsp http.ResponseWriter, req *http.Request) {

	// decode incoming message
	var payload qWebAppTapInput
	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		log.Printf("Error decoding incoming message: %v\n", err)
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}
	if s.Opts.debug {
		js, _ := json.Marshal(payload)
		log.Printf("WebApp message: %s\n", js)
	}

	// validate data received via the Mini App
	valid, msq := validateWebAppInitData(payload, s.Opts.botToken)
	if !valid {
		log.Printf("Failed message validation: %s\n", msq)
		// reject invalid hash
		if !s.Opts.webappIgnoreHash {
			http.Error(rsp, "", http.StatusForbidden)
			return
		}
	}

	// get or update tap
	if payload.Energy == 0 && payload.Score == 0 {
		s.getTaps(rsp, payload)
	} else {
		s.updateTaps(rsp, payload)
	}

}

func (s *qWebAppBack) getTaps(rsp http.ResponseWriter, payload qWebAppTapInput) {
	// get tap
	tap, err := s.storage.GetTap(int64(payload.UID))
	if err != nil || tap == nil {
		log.Printf("Error getting tap: %v\n", err)
		tap = &storage.Tap{Score: 0, Energy: 0}
	}
	js, _ := json.Marshal(tap)
	log.Printf("tap: %s\n", js)
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(http.StatusOK)
	rsp.Write(js)
}

func (s *qWebAppBack) updateTaps(rsp http.ResponseWriter, payload qWebAppTapInput) {
	tap := &storage.Tap{
		UID:    int64(payload.UID),
		Score:  0,
		Energy: int64(payload.Energy),
		Count:  1,
	}
	err := s.storage.CreateUpdateTap(tap)
	if err != nil {
		log.Printf("Error updating tap: %v\n", err)
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
	}
	if s.Opts.debug {
		dbtap, _ := s.storage.GetTap(int64(payload.UID))
		js, _ := json.Marshal(dbtap)
		log.Printf("tap updated: %s\n", js)
	}
	rsp.Header().Set("Content-Type", "application/json")
	rsp.WriteHeader(http.StatusOK)
	rsp.Write([]byte(`{"status":"ok"}`))
}

func validateWebAppInitData(payload qWebAppTapInput, botToken string) (bool, string) {
	var err error
	initParams, err := url.ParseQuery(payload.Init)
	if err != nil {
		return false, "failed to read query string"
	}
	valid := validateWebAppInitDataHash(initParams, botToken)
	if !valid {
		return false, "hash mismatch"
	}
	var initUser initParamUser
	err = json.NewDecoder(strings.NewReader(initParams.Get("user"))).Decode(&initUser)
	if err != nil {
		return false, "failed to decode"
	}
	if initUser.Id != payload.UID {
		return false, "uid missmatch"
	}
	return true, ""
}

func validateWebAppInitDataHash(values url.Values, botToken string) bool {
	// To validate data received via the Mini App,
	// one should send the data from the Telegram.WebApp.initData field to the bot's backend.
	// The data is a query string, which is composed of a series of field-value pairs.
	// https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
	receivedHash := values.Get("hash")

	// Data-check-string is a chain of all received fields, sorted alphabetically
	// in the format key=<value> with a line feed character
	var keys []string
	for k := range values {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	dataCheckString := ""
	for _, k := range keys {
		dataCheckString += k + "=" + values.Get(k) + "\n"
	}
	dataCheckString = strings.TrimSuffix(dataCheckString, "\n")

	// verify … by comparing the received hash parameter with the … signature
	// of the data-check-string with the secret key, which is the signature
	// of the bot's token with the constant string "WebAppData" used as a key.
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	h := hmac.New(sha256.New, secretKey.Sum(nil))
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	// ∃:o
	// return hmac.Equal([]byte(receivedHash), []byte(expectedHash))
	return receivedHash == expectedHash
}
