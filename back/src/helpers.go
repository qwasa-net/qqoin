package main

import (
	"os"
	"strconv"
	"strings"
)

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
