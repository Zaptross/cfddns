package main

import (
	"io"
	"net/http"
	"regexp"
)

var (
	ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
)

func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func isIPv4(ip string) bool {
	return ipv4Regex.MatchString(ip)
}
