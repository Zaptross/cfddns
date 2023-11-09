package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/samber/lo"
)

type DDArgs struct {
	// Your Cloudflare API key. You can find your API key on your Cloudflare
	// account's "My Profile" page, under the "API Keys" section.
	Token string

	// The domain to update. For example, if you want to update the record
	// "home.example.com", the domain would be "example.com".
	Domain string

	// The subdomain to update. For example, if you want to update the record
	// "home.example.com", the subdomain would be "home".
	Subdomain string

	// When true, the DNS record will be proxied through Cloudflare.
	Proxy bool

	// When true, a comment will be added to the DNS record when updated, indicating
	// that the record was updated by this tool.
	Comment bool `default:"false"`
}

func main() {
	var args DDArgs
	err := envconfig.Process("cloudflare", &args)

	if err != nil {
		errorExit(err)
	}

	api, err := cloudflare.NewWithAPIToken(args.Token)

	if err != nil {
		errorExit(err)
	}

	zoneID, err := api.ZoneIDByName(args.Domain)

	if err != nil {
		errorExit(err)
	}

	publicIP, err := getPublicIP()

	if err != nil || !isIPv4(publicIP) {
		errorExit(err)
		return
	}

	ctx := context.Background()

	records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		errorExit(err)
		return
	}

	target := strings.Join([]string{args.Subdomain, args.Domain}, ".")
	subdomain, found := lo.Find(records, func(record cloudflare.DNSRecord) bool {
		return record.Name == target
	})

	if !found {
		errorExit(errors.New("subdomain not found"))
		return
	}

	if publicIP == subdomain.Content {
		successExit("No changes detected")
		return
	}

	// If the comment flag is set false, persist the existing comment.
	msg := subdomain.Comment
	if args.Comment {
		msg = fmt.Sprintf("cfddns:%s->%s@%s", subdomain.Content, publicIP, time.Now().Format("2006-01-02T15:04"))
	}

	update := cloudflare.UpdateDNSRecordParams{
		Proxied:  &args.Proxy,
		Comment:  &msg,
		Type:     subdomain.Type,
		Name:     subdomain.Name,
		Content:  publicIP,
		TTL:      subdomain.TTL,
		Data:     subdomain.Data,
		ID:       subdomain.ID,
		Priority: subdomain.Priority,
		Tags:     subdomain.Tags,
	}

	_, err = api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(subdomain.ZoneID), update)

	if err != nil {
		errorExit(err)
		return
	}

	successExit(msg)
}

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

func successExit(msg string) {
	fmt.Println(msg)
	os.Exit(0)
}

func errorExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func isIPv4(ip string) bool {
	chunks := strings.Split(ip, ".")

	if len(chunks) != 4 {
		return false
	}

	for _, chunk := range chunks {
		if cl := len(chunk); cl == 0 || cl > 3 {
			return false
		}

		for _, char := range chunk {
			if char < '0' || char > '9' {
				return false
			}
		}
	}

	return true
}
