package main

import (
	"fmt"
	"net/http"
	"strings"
)

func CheckHasCookieHeader(header http.Header) bool {
	fmt.Printf("%+v\n", header)
	if strings.Contains(strings.Join(header["Cookie"], " "), "GCLB=") {
		return true
	}

	return false
}
