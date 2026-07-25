# IP Tracker

A powerful command-line IP intelligence tool written in Go. Track IP addresses, scan ports, run DNS lookups, traceroutes, and more — all from your terminal.

![Version](https://img.shields.io/badge/version-8.0.0-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

| Feature | Description |
|---------|-------------|
| **IP Geolocation** | Country, city, coordinates, timezone, ISP, ASN |
| **Proxy/VPN Detection** | Detect if an IP uses a proxy, VPN, or is hosted |
| **Batch Tracking** | Track multiple IPs from a file at once |
| **DNS Lookup** | Forward (A, MX, NS, TXT, CNAME) and reverse DNS |
| **Port Scanner** | Scan common ports or custom ranges |
| **Traceroute** | Hop-by-hop path tracing with latency |
| **Connectivity Test** | HTTP/HTTPS test with TLS info and latency benchmark |
| **Export** | Save results to JSON or CSV |
| **Search History** | Keeps last 500 lookups in `~/.ip_tracker_history.json` |
| **CLI Flags** | Full non-interactive mode with JSON output |

## Screenshots

### Main Menu
![Main Menu](assets/main.png)

### Self Tracking
![Self Tracking](assets/self_track.png)

### Manual IP Lookup
![Manual Lookup](assets/manual_track.png)

## Installation

### Using Go Install

```bash
go install github.com/Devraj-jha/IP-Tracker@latest
```

### Building from Source

```bash
git clone https://github.com/Devraj-jha/IP-Tracker.git
cd IP-Tracker
go build -o ip-tracker .
```

### Using Make

```bash
git clone https://github.com/Devraj-jha/IP-Tracker.git
cd IP-Tracker
make build        # Build for current platform
make darwin       # Build for macOS (arm64 + amd64)
make linux        # Build for Linux (amd64)
make windows      # Build for Windows (amd64)
make all-platforms # Build for all platforms
```

## How to Run

### Interactive Mode

Just run the binary with no arguments:

```bash
./ip-tracker
```

You'll see the banner and main menu:

```
  => MAIN MENU <=
  1.  Track MY IP Address
  2.  Track ANOTHER IP Address
  3.  Batch Track from File
  4.  DNS Lookup
  5.  Port Scanner
  6.  Traceroute
  7.  Connectivity & Latency Test
  8.  View Search History
  9.  Exit
```

Type a number and press Enter to use any feature.

### CLI Mode (Non-Interactive)

Use flags for scripting and automation:

```bash
# Track your own IP
./ip-tracker -self

# Track a specific IP
./ip-tracker 8.8.8.8
./ip-tracker -ip 8.8.8.8

# Output as JSON (great for scripting)
./ip-tracker -self -json
./ip-tracker 8.8.8.8 -json

# Export to file
./ip-tracker -self -export results.json
./ip-tracker -ip 8.8.8.8 -export results.csv

# Batch track from file
./ip-tracker -batch ips.txt
./ip-tracker -batch ips.txt -json

# DNS lookup
./ip-tracker -dns example.com
./ip-tracker -dns example.com -json

# Reverse DNS
./ip-tracker -r 8.8.8.8

# View history
./ip-tracker -history

# Show help
./ip-tracker -help

# Show version
./ip-tracker -version
```

### Batch File Format

Create a text file with one IP per line:

```
# ips.txt — lines starting with # are ignored
8.8.8.8
1.1.1.1
208.67.222.222
```

Then run:

```bash
./ip-tracker -batch ips.txt -json -export results.json
```

## CLI Flags Reference

| Flag | Description | Example |
|------|-------------|---------|
| `-self` | Track your own public IP | `./ip-tracker -self` |
| `-ip <addr>` | Track a specific IP | `./ip-tracker -ip 8.8.8.8` |
| `-json` | Output as JSON | `./ip-tracker -self -json` |
| `-export <file>` | Export to JSON/CSV | `./ip-tracker -self -export out.json` |
| `-batch <file>` | Batch track from file | `./ip-tracker -batch ips.txt` |
| `-dns <host>` | DNS lookup | `./ip-tracker -dns example.com` |
| `-r <ip>` | Reverse DNS lookup | `./ip-tracker -r 8.8.8.8` |
| `-history` | View search history | `./ip-tracker -history` |
| `-version` | Show version | `./ip-tracker -version` |
| `-help` | Show help | `./ip-tracker -help` |

## Interactive Features

### 1. Track My IP
Automatically detects your public IP and displays full geolocation data.

### 2. Track Another IP
Enter any IPv4 or IPv6 address to get geolocation and network info.

### 3. Batch Track
Load a file of IPs and track them all at once with export options.

### 4. DNS Lookup
- **Reverse DNS**: IP → hostname
- **Forward DNS**: hostname → IP, MX, NS, TXT, CNAME records
- **Full DNS**: Complete record dump with CNAME chain following

### 5. Port Scanner
- **Common Ports**: Scans 25+ well-known services (HTTP, SSH, MySQL, etc.)
- **Custom Range**: Scan any port range (1-65535)
- **Single Port**: Check one port with latency

### 6. Traceroute
Hop-by-hop network path tracing with latency measurement per hop.

### 7. Connectivity Test
- **HTTP/HTTPS Test**: DNS time, connect time, TLS version, certificate info
- **Latency Benchmark**: Tests multiple CDN targets and averages response time

### 8. Search History
View your last 20 lookups (full history stored in `~/.ip_tracker_history.json`).

## API Reference

| Service | Endpoint | Purpose | Limits |
|---------|----------|---------|--------|
| ipify.org | https://api.ipify.org | Get public IP | None |
| ip-api.com | http://ip-api.com/json/{ip} | Geolocation data | 45 req/min |
| checkip.amazonaws.com | https://checkip.amazonaws.com | Fallback public IP | None |

## Example Output

```
  ════════════════════════════════════════════════════════════
  GEOLOCATION & NETWORK INTELLIGENCE REPORT
  ════════════════════════════════════════════════════════════

  ┌─ IP ADDRESS
  │   IP:             8.8.8.8

  ├─ LOCATION
  │   Country:        United States (US)
  │   Region:         California (CA)
  │   City:           Mountain View
  │   Postal Code:    94035
  │   Coordinates:    37.4056, -122.0775
  │   Timezone:       America/Los_Angeles

  ├─ NETWORK & ISP
  │   ISP:            Google LLC
  │   Organization:   Google LLC
  │   AS Number:      AS15169
  │   AS Name:        GOOGLE

  ├─ INTELLIGENCE
  │   Mobile:         No
  │   Proxy/VPN:      No
  │   Hosting/Cloud:  Yes

  └──────────────────────────────────────────────────────────

  Google Maps: https://maps.google.com/?q=37.405600,-122.077500
```

## Building for Different Platforms

```bash
# Using Make (recommended)
make all-platforms

# Manual builds
GOOS=linux GOARCH=amd64 go build -o ip-tracker-linux main.go
GOOS=darwin GOARCH=arm64 go build -o ip-tracker-mac-arm main.go
GOOS=darwin GOARCH=amd64 go build -o ip-tracker-mac-intel main.go
GOOS=windows GOARCH=amd64 go build -o ip-tracker.exe main.go
```

## Requirements

| Requirement | Version |
|-------------|---------|
| Go | 1.22 or higher |
| Internet | Required for API calls |
| OS | Linux, macOS, Windows |

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Invalid IP | Use proper format: `192.168.1.1` (IPv4) or `2001:db8::1` (IPv6) |
| API rate limit | Wait 60 seconds between requests |
| No internet | Check network connection and firewall |
| Private IP | Private IPs (192.168.x.x) can't be geolocated publicly |
| Port scan blocked | Some ports may be filtered by firewall |
| Traceroute fails | May need elevated permissions on some systems |

## License

MIT

---

Made by Devraj-jha
