# cfddns

A simple cloudflare ddns script written in Go, for use as a cron job in Kubernetes.

Checkout the repository here: [GitHub](https://github.com/Zaptross/cfddns)

## Setup

1. Clone the repository

Create your cloudflare API token from the [API Tokens](https://dash.cloudflare.com/profile/api-tokens) page.

2. Ensure the token has the following permissions:

   - Zone - Zone Settings - Read
   - Zone - Zone - Read
   - Zone - DNS - Edit

3. Edit the `values.yaml` file and add your token, domain, subdomain and whether you want to use the proxied option.

   - `token`: Your cloudflare API token
   - `domain`: Your domain name
   - `subdomain`: Your subdomain name
   - `proxied`: Whether you want to use the proxied option
   - `comment`: Whether to add a comment to the DNS record indicating it was updated by this script (default: true)

## Usage

To install the application using Helm, follow these steps:

```bash
helm install -f values.yaml cfddns ./helm
```
