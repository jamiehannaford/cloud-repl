package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
)

func ensureMethod(r *http.Request, expected string) {
	if r.Method != expected {
		panic(fmt.Sprintf("%s is not an expected method", r.Method))
	}
}

func checkErr(msg string, err error) {
	if err != nil {
		panic(fmt.Sprintf("An error occurred while %s: %s", msg, err.Error()))
	}
}

func hyphens(count int) string {
	return strings.Repeat("-", count+2)
}

func randomStr(prefix string, n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return prefix + string(bytes)
}
