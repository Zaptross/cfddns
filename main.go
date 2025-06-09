package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

var (
	ipv4Regex = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
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
	slog.Info(getVersion())

	var args DDArgs
	err := envconfig.Process("cloudflare", &args)

	if err != nil {
		slog.Error("Failed to process environment variables", "error", err)
		os.Exit(1)
	}

	api, err := cloudflare.NewWithAPIToken(args.Token)

	if err != nil {
		slog.Error("Failed to create Cloudflare API client", "error", err)
		os.Exit(1)
	}

	zoneID, err := api.ZoneIDByName(args.Domain)

	if err != nil {
		slog.Error("Failed to get zone ID for domain", "domain", args.Domain, "error", err)
		os.Exit(1)
	}

	publicIP, err := getPublicIP()

	if err != nil {
		slog.Error("Failed to get public IP address", "error", err)
		return
	}

	if !isIPv4(publicIP) {
		slog.Error("Invalid public IP address", "ip", publicIP)
		os.Exit(1)
	}

	ctx := context.Background()

	records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		slog.Error("Failed to list DNS records", "error", err)
		os.Exit(1)
		return
	}

	target := strings.Join([]string{args.Subdomain, args.Domain}, ".")
	subdomain, found := lo.Find(records, func(record cloudflare.DNSRecord) bool {
		return record.Name == target
	})

	if !found {
		slog.Error("Subdomain not found", "subdomain", target)
		os.Exit(1)
		return
	}

	if publicIP == subdomain.Content {
		slog.Info("No update needed", "subdomain", subdomain.Name, "currentIP", subdomain.Content, "publicIP", publicIP)
		os.Exit(0)
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
		slog.Error("Failed to update DNS record", "subdomain", subdomain.Name, "error", err)
		os.Exit(1)
		return
	}

	slog.Info("DNS record updated successfully", "message", msg)
	os.Exit(0)
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

func isIPv4(ip string) bool {
	return ipv4Regex.MatchString(ip)
}
