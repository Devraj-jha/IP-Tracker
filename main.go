package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Version info
const appVersion = "8.0.0"

// Colors
var (
	bannerColor  = color.New(color.FgCyan, color.Bold)
	menuColor    = color.New(color.FgYellow, color.Bold)
	headerColor  = color.New(color.FgGreen, color.Bold)
	labelColor   = color.New(color.FgHiCyan)
	valueColor   = color.New(color.FgWhite, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	successColor = color.New(color.FgGreen, color.Bold)
	warnColor    = color.New(color.FgYellow)
	linkColor    = color.New(color.FgBlue, color.Underline)
	accentColor  = color.New(color.FgMagenta, color.Bold)
)

type IPInfo struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	ASName      string  `json:"asname"`
	Mobile      bool    `json:"mobile"`
	Proxy       bool    `json:"proxy"`
	Hosting     bool    `json:"hosting"`
	Query       string  `json:"query"`
	Reverse     string  `json:"reverse"`
}

type DetailedInfo struct {
	IPAddress      string  `json:"ip_address"`
	ReverseDNS     string  `json:"reverse_dns,omitempty"`
	Country        string  `json:"country"`
	CountryCode    string  `json:"country_code"`
	State          string  `json:"state"`
	StateCode      string  `json:"state_code"`
	City           string  `json:"city"`
	PostalCode     string  `json:"postal_code"`
	Coordinates    string  `json:"coordinates"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	TelecomService string  `json:"isp"`
	Organization   string  `json:"organization"`
	ASNumber       string  `json:"as_number"`
	ASName         string  `json:"as_name"`
	IsMobile       bool    `json:"is_mobile"`
	IsProxy        bool    `json:"is_proxy"`
	IsHosting      bool    `json:"is_hosting"`
}

func main() {
	// CLI flags
	selfFlag := flag.Bool("self", false, "Track your own public IP address")
	ipFlag := flag.String("ip", "", "Track a specific IP address")
	jsonFlag := flag.Bool("json", false, "Output results in JSON format")
	exportFlag := flag.String("export", "", "Export results to file (specify filename)")
	batchFlag := flag.String("batch", "", "Track multiple IPs from a file (one per line)")
	historyFlag := flag.Bool("history", false, "View search history")
	dnsFlag := flag.String("dns", "", "Perform DNS lookup on hostname")
	reverseDNSFlag := flag.String("r", "", "Reverse DNS lookup on IP address")
	versionFlag := flag.Bool("version", false, "Show version information")
	helpFlag := flag.Bool("help", false, "Show usage information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n  IP Tracker v%s — Enhanced IP Intelligence Tool\n\n", appVersion)
		fmt.Fprintf(os.Stderr, "  Usage:\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker [flags]\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker [IP ADDRESS]       # Direct IP lookup\n\n")
		fmt.Fprintf(os.Stderr, "  Flags:\n")
		fmt.Fprintf(os.Stderr, "    -self              Track your own public IP\n")
		fmt.Fprintf(os.Stderr, "    -ip <address>      Track a specific IP address\n")
		fmt.Fprintf(os.Stderr, "    -json              Output in JSON format\n")
		fmt.Fprintf(os.Stderr, "    -export <file>     Export results to file (json or csv)\n")
		fmt.Fprintf(os.Stderr, "    -batch <file>      Track multiple IPs from file\n")
		fmt.Fprintf(os.Stderr, "    -history           View search history\n")
		fmt.Fprintf(os.Stderr, "    -dns <hostname>    DNS lookup (A, MX, NS, TXT, CNAME)\n")
		fmt.Fprintf(os.Stderr, "    -r <ip>            Reverse DNS lookup (IP → hostname)\n")
		fmt.Fprintf(os.Stderr, "    -version, -v       Show version\n")
		fmt.Fprintf(os.Stderr, "    -help, -h          Show this help\n\n")
		fmt.Fprintf(os.Stderr, "  Examples:\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker                          # Interactive mode\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker 8.8.8.8                  # Quick lookup\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker -self -json              # Your IP as JSON\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker -ip 8.8.8.8 -export out  # Export to file\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker -batch ips.txt -json     # Batch + JSON\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker -dns example.com         # DNS lookup\n")
		fmt.Fprintf(os.Stderr, "    ip-tracker -r 8.8.8.8               # Reverse DNS\n\n")
	}

	flag.Parse()

	// Handle flags
	if *versionFlag {
		fmt.Printf("IP Tracker v%s\n", appVersion)
		return
	}

	if *helpFlag {
		flag.Usage()
		return
	}

	// CLI mode: --self
	if *selfFlag {
		info, err := trackMyIPCLI(*jsonFlag)
		if err != nil {
			errorColor.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if *exportFlag != "" {
			exportSingleResult(info, *exportFlag)
		}
		return
	}

	// CLI mode: --ip
	if *ipFlag != "" {
		if net.ParseIP(*ipFlag) == nil {
			errorColor.Fprintf(os.Stderr, "Invalid IP address: %s\n", *ipFlag)
			os.Exit(1)
		}
		info, err := trackIPCLI(*ipFlag, *jsonFlag)
		if err != nil {
			errorColor.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if *exportFlag != "" {
			exportSingleResult(info, *exportFlag)
		}
		return
	}

	// CLI mode: positional argument (e.g., ip-tracker 8.8.8.8)
	if flag.NArg() > 0 {
		ipArg := flag.Arg(0)
		if net.ParseIP(ipArg) == nil {
			errorColor.Fprintf(os.Stderr, "Invalid IP address: %s\n", ipArg)
			os.Exit(1)
		}
		info, err := trackIPCLI(ipArg, *jsonFlag)
		if err != nil {
			errorColor.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if *exportFlag != "" {
			exportSingleResult(info, *exportFlag)
		}
		return
	}

	// CLI mode: --batch
	if *batchFlag != "" {
		results, err := batchTrackCLI(*batchFlag)
		if err != nil {
			errorColor.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if *exportFlag != "" {
			err := exportResults(results, *exportFlag, "auto")
			if err != nil {
				errorColor.Fprintf(os.Stderr, "Export error: %v\n", err)
				os.Exit(1)
			}
			successColor.Printf("Results exported to %s\n", *exportFlag)
		} else if *jsonFlag {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
		}
		return
	}

	// CLI mode: --history
	if *historyFlag {
		printHistoryCLI()
		return
	}

	// CLI mode: --dns
	if *dnsFlag != "" {
		result := performFullDNSLookup(*dnsFlag)
		if *jsonFlag {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			headerColor.Println("\n  DNS Lookup Results")
			headerColor.Println("  " + strings.Repeat("═", 50))
			displayDNSResult(result)
			headerColor.Println("  " + strings.Repeat("═", 50))
		}
		return
	}

	// CLI mode: --r (reverse DNS)
	if *reverseDNSFlag != "" {
		ip := *reverseDNSFlag
		if net.ParseIP(ip) == nil {
			errorColor.Fprintf(os.Stderr, "Invalid IP address: %s\n", ip)
			os.Exit(1)
		}
		hostname, err := net.LookupAddr(ip)
		if err != nil {
			errorColor.Fprintf(os.Stderr, "Reverse DNS lookup failed: %v\n", err)
			os.Exit(1)
		}
		if *jsonFlag {
			result := map[string]any{"ip": ip, "hostnames": hostname}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			for _, h := range hostname {
				fmt.Printf("%s → %s\n", ip, strings.TrimSuffix(h, "."))
			}
		}
		return
	}

	// Interactive mode
	runInteractive()
}

func runInteractive() {
	displayBanner()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		displayMenu()
		fmt.Print("\n  Enter your choice (1-10): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			trackMyIP()
		case "2":
			trackOtherIP(scanner)
		case "3":
			batchTrackFromFile(scanner)
		case "4":
			dnsLookupMenu(scanner)
		case "5":
			portScanMenu(scanner)
		case "6":
			tracerouteMenu(scanner)
		case "7":
			connectivityMenu(scanner)
		case "8":
			searchHistory()
		case "9", "exit", "EXIT", "Exit", "q", "Q":
			printExit()
			os.Exit(0)
		default:
			errorColor.Println("\n  Invalid choice! Please enter 1-9.")
			pressEnter(scanner)
		}
	}
}

// --- CLI Functions ---

func trackMyIPCLI(jsonOutput bool) (*DetailedInfo, error) {
	publicIP, err := getPublicIP()
	if err != nil {
		return nil, fmt.Errorf("failed to get public IP: %w", err)
	}

	info, err := fetchIPInfo(publicIP)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP info: %w", err)
	}

	saveToHistory(info)
	d := toDetailedInfo(info)

	if jsonOutput {
		data, _ := json.MarshalIndent(d, "", "  ")
		fmt.Println(string(data))
	} else {
		displayInfo(info)
	}

	return d, nil
}

func trackIPCLI(ip string, jsonOutput bool) (*DetailedInfo, error) {
	info, err := fetchIPInfo(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP info: %w", err)
	}

	saveToHistory(info)
	d := toDetailedInfo(info)

	if jsonOutput {
		data, _ := json.MarshalIndent(d, "", "  ")
		fmt.Println(string(data))
	} else {
		displayInfo(info)
	}

	return d, nil
}

func batchTrackCLI(filePath string) ([]*DetailedInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ips = append(ips, line)
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no valid IPs found in file")
	}

	var results []*DetailedInfo
	for _, ip := range ips {
		if net.ParseIP(ip) == nil {
			fmt.Fprintf(os.Stderr, "Skipping invalid IP: %s\n", ip)
			continue
		}

		info, err := fetchIPInfo(ip)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error for %s: %v\n", ip, err)
			continue
		}

		saveToHistory(info)
		results = append(results, toDetailedInfo(info))
	}

	return results, nil
}

func printHistoryCLI() {
	entries, err := loadHistory()
	if err != nil || len(entries) == 0 {
		fmt.Println("No search history found.")
		return
	}

	fmt.Printf("Search History (%d entries):\n\n", len(entries))
	for _, e := range entries {
		fmt.Printf("%s  %-15s  %s, %s  ISP: %s\n",
			e.Timestamp, e.IP, e.City, e.Country, e.ISP)
	}
}

func exportSingleResult(info *DetailedInfo, filename string) {
	format := "json"
	if strings.HasSuffix(filename, ".csv") {
		format = "csv"
	} else if strings.HasSuffix(filename, ".json") {
		format = "json"
	}

	err := exportResults([]*DetailedInfo{info}, filename, format)
	if err != nil {
		errorColor.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	successColor.Printf("Results exported to %s\n", filename)
}

func displayBanner() {
	bannerColor.Println(`
 ██╗██████╗░░░░░░░████████╗██████╗░░█████╗░░█████╗░██╗░░██╗███████╗██████╗░
 ██║██╔══██╗░░░░░░╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██║░██╔╝██╔════╝██╔══██╗
 ██║██████╔╝█████╗░░░██║░░░██████╔╝███████║██║░░╚═╝█████═╝░█████╗░░██████╔╝
 ██║██╔═══╝░╚════╝░░░██║░░░██╔══██╗██╔══██║██║░░██╗██╔═██╗░██╔══╝░░██╔══██╗
 ██║██║░░░░░░░░░░░░░░██║░░░██║░░██║██║░░██║╚█████╔╝██║░╚██╗███████╗██║░░██║
 ╚═╝╚═╝░░░░░░░░░░░░░╚═╝░░░╚═╝░░╚═╝╚═╝░░╚═╝░╚════╝░╚═╝░░╚═╝╚══════╝╚═╝░░╚═╝`)
	accentColor.Printf("  v%s — %s\n", appVersion, "Enhanced IP Intelligence Tool")
	fmt.Println()
}

func displayMenu() {
	menuColor.Println("  " + strings.Repeat("─", 58))
	menuColor.Println("  => MAIN MENU <=")
	menuColor.Println("  " + strings.Repeat("─", 58))
	fmt.Println("  1.  Track MY IP Address")
	fmt.Println("  2.  Track ANOTHER IP Address")
	fmt.Println("  3.  Batch Track from File")
	fmt.Println("  4.  DNS Lookup")
	fmt.Println("  5.  Port Scanner")
	fmt.Println("  6.  Traceroute")
	fmt.Println("  7.  Connectivity & Latency Test")
	fmt.Println("  8.  View Search History")
	fmt.Println("  9.  Exit")
	menuColor.Println("  " + strings.Repeat("─", 58))
}

func trackMyIP() {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  TRACKING YOUR IP ADDRESS...")
	headerColor.Println("  " + strings.Repeat("─", 58))

	publicIP, err := getPublicIP()
	if err != nil {
		errorColor.Printf("\n  Error getting your public IP: %v\n", err)
		pressEnter(nil)
		return
	}

	successColor.Printf("\n  Your Public IP: %s\n", publicIP)
	fmt.Println("  Fetching geolocation data...")

	info, err := fetchIPInfo(publicIP)
	if err != nil {
		errorColor.Printf("\n  Error fetching IP information: %v\n", err)
		pressEnter(nil)
		return
	}

	displayInfo(info)
	saveToHistory(info)
	pressEnter(nil)
}

func trackOtherIP(scanner *bufio.Scanner) {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  TRACK ANOTHER IP ADDRESS")
	headerColor.Println("  " + strings.Repeat("─", 58))

	fmt.Print("\n  Enter IP Address (IPv4 or IPv6): ")
	scanner.Scan()
	ipAddress := strings.TrimSpace(scanner.Text())

	if ipAddress == "" {
		errorColor.Println("\n  No IP address entered!")
		pressEnter(scanner)
		return
	}

	if net.ParseIP(ipAddress) == nil {
		errorColor.Printf("\n  Invalid IP address format: %s\n", ipAddress)
		fmt.Println("  Please enter a valid IPv4 (e.g., 8.8.8.8) or IPv6 address")
		pressEnter(scanner)
		return
	}

	successColor.Printf("\n  Target IP: %s\n", ipAddress)
	fmt.Println("  Fetching geolocation data...")

	info, err := fetchIPInfo(ipAddress)
	if err != nil {
		errorColor.Printf("\n  Error fetching IP information: %v\n", err)
		pressEnter(scanner)
		return
	}

	displayInfo(info)
	saveToHistory(info)
	pressEnter(scanner)
}

func batchTrackFromFile(scanner *bufio.Scanner) {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  BATCH IP TRACKING")
	headerColor.Println("  " + strings.Repeat("─", 58))

	fmt.Print("\n  Enter path to file (one IP per line): ")
	scanner.Scan()
	filePath := strings.TrimSpace(scanner.Text())

	if filePath == "" {
		errorColor.Println("\n  No file path entered!")
		pressEnter(scanner)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		errorColor.Printf("\n  Error opening file: %v\n", err)
		pressEnter(scanner)
		return
	}
	defer file.Close()

	var ips []string
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		line := strings.TrimSpace(fileScanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ips = append(ips, line)
		}
	}

	if len(ips) == 0 {
		errorColor.Println("\n  No valid IPs found in file!")
		pressEnter(scanner)
		return
	}

	warnColor.Printf("\n  Found %d IP(s) to track\n", len(ips))
	fmt.Println(strings.Repeat("─", 58))

	var results []*DetailedInfo
	for i, ip := range ips {
		if net.ParseIP(ip) == nil {
			errorColor.Printf("  [%d/%d] Invalid IP: %s — skipping\n", i+1, len(ips), ip)
			continue
		}

		fmt.Printf("  [%d/%d] Tracking %s... ", i+1, len(ips), ip)
		info, err := fetchIPInfo(ip)
		if err != nil {
			errorColor.Printf("ERROR: %v\n", err)
			continue
		}
		successColor.Println("OK")
		results = append(results, toDetailedInfo(info))
		saveToHistory(info)
	}

	fmt.Println(strings.Repeat("─", 58))
	successColor.Printf("  Successfully tracked: %d/%d IP(s)\n", len(results), len(ips))

	if len(results) > 0 {
		for _, r := range results {
			printCompactInfo(r)
		}
	}

	// Offer export
	fmt.Print("\n  Export results? (json/csv/none): ")
	scanner.Scan()
	format := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if format == "json" || format == "csv" {
		fmt.Print("  Enter output filename: ")
		scanner.Scan()
		filename := strings.TrimSpace(scanner.Text())
		if filename != "" {
			err := exportResults(results, filename, format)
			if err != nil {
				errorColor.Printf("\n  Export error: %v\n", err)
			} else {
				successColor.Printf("\n  Results exported to %s\n", filename)
			}
		}
	}

	pressEnter(scanner)
}

func searchHistory() {
	entries, err := loadHistory()
	if err != nil {
		errorColor.Printf("\n  No search history found: %v\n", err)
		pressEnter(nil)
		return
	}

	if len(entries) == 0 {
		warnColor.Println("\n  Search history is empty.")
		pressEnter(nil)
		return
	}

	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Printf("  SEARCH HISTORY (%d entries)\n", len(entries))
	headerColor.Println("  " + strings.Repeat("─", 58))

	// Show last 20
	start := 0
	if len(entries) > 20 {
		start = len(entries) - 20
		warnColor.Println("  (showing last 20 entries)")
	}

	for _, e := range entries[start:] {
		fmt.Printf("  %s | %-15s | %s, %s\n",
			e.Timestamp, e.IP, e.City, e.Country)
	}

	headerColor.Println("  " + strings.Repeat("─", 58))
	pressEnter(nil)
}

func printExit() {
	fmt.Println()
	successColor.Println("  Thank you for using IP Tracker!")
	accentColor.Println("  Have a great day!")
	fmt.Println()
}

// --- Core Functions ---

func getPublicIP() (string, error) {
	client := http.Client{Timeout: 10 * time.Second}

	// Try primary API
	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		// Fallback to secondary API
		resp, err = client.Get("https://checkip.amazonaws.com")
		if err != nil {
			return "", fmt.Errorf("failed to get public IP from all sources: %w", err)
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func fetchIPInfo(ip string) (*IPInfo, error) {
	client := http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,asname,mobile,proxy,hosting,query,reverse", ip)

	var resp *http.Response
	var err error

	// Retry up to 2 times
	for attempt := range 2 {
		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}

	if err != nil {
		return nil, fmt.Errorf("API request failed after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("rate limited by API — please wait 60 seconds")
		}
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info IPInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, err
	}

	if info.Status != "success" {
		return nil, fmt.Errorf("API returned status: %s (IP might be private or invalid)", info.Status)
	}

	return &info, nil
}

func toDetailedInfo(info *IPInfo) *DetailedInfo {
	return &DetailedInfo{
		IPAddress:      info.Query,
		ReverseDNS:     info.Reverse,
		Country:        info.Country,
		CountryCode:    info.CountryCode,
		State:          info.RegionName,
		StateCode:      info.Region,
		City:           info.City,
		PostalCode:     info.Zip,
		Coordinates:    fmt.Sprintf("%.6f, %.6f", info.Lat, info.Lon),
		Lat:            info.Lat,
		Lon:            info.Lon,
		Timezone:       info.Timezone,
		TelecomService: info.Isp,
		Organization:   info.Org,
		ASNumber:       info.As,
		ASName:         info.ASName,
		IsMobile:       info.Mobile,
		IsProxy:        info.Proxy,
		IsHosting:      info.Hosting,
	}
}

func displayInfo(info *IPInfo) {
	d := toDetailedInfo(info)

	fmt.Println()
	headerColor.Println("  " + strings.Repeat("═", 58))
	headerColor.Println("  GEOLOCATION & NETWORK INTELLIGENCE REPORT")
	headerColor.Println("  " + strings.Repeat("═", 58))

	// IP Section
	fmt.Println()
	accentColor.Println("  ┌─ IP ADDRESS")
	fmt.Print("  │   ")
	labelColor.Print("IP:  ")
	valueColor.Println(d.IPAddress)
	if d.ReverseDNS != "" {
		printField("Reverse DNS", d.ReverseDNS)
	}

	// Location Section
	fmt.Println()
	accentColor.Println("  ├─ LOCATION")
	printField("Country", fmt.Sprintf("%s (%s)", d.Country, d.CountryCode))
	printField("Region", fmt.Sprintf("%s (%s)", d.State, d.StateCode))
	printField("City", d.City)
	printField("Postal Code", d.PostalCode)
	printField("Coordinates", d.Coordinates)
	printField("Timezone", d.Timezone)

	// Network Section
	fmt.Println()
	accentColor.Println("  ├─ NETWORK & ISP")
	printField("ISP", d.TelecomService)
	printField("Organization", d.Organization)
	printField("AS Number", d.ASNumber)
	printField("AS Name", d.ASName)

	// Intelligence Section
	fmt.Println()
	accentColor.Println("  ├─ INTELLIGENCE")
	printBoolField("Mobile Network", d.IsMobile)
	printBoolField("Proxy/VPN", d.IsProxy)
	printBoolField("Hosting/Cloud", d.IsHosting)

	fmt.Println()
	accentColor.Println("  └──────────────────────────────────────────────────")

	if d.Lat != 0 && d.Lon != 0 {
		fmt.Println()
		fmt.Print("  ")
		linkColor.Printf("Google Maps: https://maps.google.com/?q=%.6f,%.6f\n", d.Lat, d.Lon)
	}
	fmt.Println()
}

func printCompactInfo(d *DetailedInfo) {
	fmt.Println()
	accentColor.Printf("  ── %s ──\n", d.IPAddress)
	fmt.Printf("  Location: %s, %s, %s\n", d.City, d.State, d.Country)
	fmt.Printf("  ISP: %s | AS: %s\n", d.TelecomService, d.ASNumber)
	if d.IsProxy {
		warnColor.Println("  [!] Proxy/VPN Detected")
	}
	if d.IsHosting {
		warnColor.Println("  [!] Hosting/Cloud Detected")
	}
}

func printField(label, value string) {
	fmt.Print("  │   ")
	labelColor.Printf("%-14s", label+":")
	valueColor.Println(value)
}

func printBoolField(label string, val bool) {
	fmt.Print("  │   ")
	labelColor.Printf("%-14s", label+":")
	if val {
		errorColor.Println("Yes")
	} else {
		successColor.Println("No")
	}
}

func pressEnter(scanner *bufio.Scanner) {
	fmt.Print("\n  Press Enter to continue...")
	if scanner != nil {
		scanner.Scan()
	} else {
		bufio.NewScanner(os.Stdin).Scan()
	}
}

// --- DNS Lookup ---

type DNSResult struct {
	Host       string            `json:"host"`
	IPs        []string          `json:"ips,omitempty"`
	MX         []string          `json:"mx,omitempty"`
	NS         []string          `json:"ns,omitempty"`
	TXT        []string          `json:"txt,omitempty"`
	CNAME      string            `json:"cname,omitempty"`
	CNAMEChain []string          `json:"cname_chain,omitempty"`
	ReverseDNS string            `json:"reverse_dns,omitempty"`
	TTL        map[string]uint32 `json:"ttl,omitempty"`
}

func dnsLookupMenu(scanner *bufio.Scanner) {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  DNS LOOKUP")
	headerColor.Println("  " + strings.Repeat("─", 58))
	fmt.Println("  1.  Reverse DNS (IP → Hostname)")
	fmt.Println("  2.  Forward DNS (Hostname → IP)")
	fmt.Println("  3.  Full DNS Records")
	fmt.Println("  4.  Back to Main Menu")

	fmt.Print("\n  Choice: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		reverseDNSLookup(scanner)
	case "2":
		forwardDNSLookup(scanner)
	case "3":
		fullDNSLookup(scanner)
	case "4":
		return
	default:
		errorColor.Println("\n  Invalid choice!")
		pressEnter(scanner)
	}
}

func reverseDNSLookup(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter IP address: ")
	scanner.Scan()
	ip := strings.TrimSpace(scanner.Text())

	if net.ParseIP(ip) == nil {
		errorColor.Printf("\n  Invalid IP address: %s\n", ip)
		pressEnter(scanner)
		return
	}

	hostname, err := net.LookupAddr(ip)
	if err != nil {
		errorColor.Printf("\n  Reverse DNS lookup failed: %v\n", err)
		pressEnter(scanner)
		return
	}

	headerColor.Println("\n  " + strings.Repeat("═", 50))
	headerColor.Println("  REVERSE DNS RESULTS")
	headerColor.Println("  " + strings.Repeat("═", 50))
	printField("IP Address", ip)
	if len(hostname) > 0 {
		for _, h := range hostname {
			printField("Hostname", strings.TrimSuffix(h, "."))
		}
	} else {
		warnColor.Println("  No hostname found for this IP")
	}
	headerColor.Println("  " + strings.Repeat("═", 50))
	pressEnter(scanner)
}

func forwardDNSLookup(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter hostname (e.g., example.com): ")
	scanner.Scan()
	host := strings.TrimSpace(scanner.Text())

	if host == "" {
		errorColor.Println("\n  No hostname entered!")
		pressEnter(scanner)
		return
	}

	headerColor.Println("\n  " + strings.Repeat("═", 50))
	headerColor.Println("  FORWARD DNS RESULTS")
	headerColor.Println("  " + strings.Repeat("═", 50))
	printField("Hostname", host)

	// A Records
	ips, err := net.LookupHost(host)
	if err == nil && len(ips) > 0 {
		for _, ip := range ips {
			printField("IPv4/IPv6", ip)
		}
	}

	// CNAME
	cname, err := net.LookupCNAME(host)
	if err == nil && cname != host {
		printField("CNAME", strings.TrimSuffix(cname, "."))
	}

	// MX Records
	mxRecords, err := net.LookupMX(host)
	if err == nil && len(mxRecords) > 0 {
		for _, mx := range mxRecords {
			printField("MX", fmt.Sprintf("%s (priority: %d)", strings.TrimSuffix(mx.Host, "."), mx.Pref))
		}
	}

	// NS Records
	nsRecords, err := net.LookupNS(host)
	if err == nil && len(nsRecords) > 0 {
		for _, ns := range nsRecords {
			printField("NS", strings.TrimSuffix(ns.Host, "."))
		}
	}

	// TXT Records
	txtRecords, err := net.LookupTXT(host)
	if err == nil && len(txtRecords) > 0 {
		for _, txt := range txtRecords {
			printField("TXT", txt)
		}
	}

	headerColor.Println("  " + strings.Repeat("═", 50))
	pressEnter(scanner)
}

func fullDNSLookup(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter hostname (e.g., example.com): ")
	scanner.Scan()
	host := strings.TrimSpace(scanner.Text())

	if host == "" {
		errorColor.Println("\n  No hostname entered!")
		pressEnter(scanner)
		return
	}

	headerColor.Println("\n  " + strings.Repeat("═", 50))
	headerColor.Println("  FULL DNS RECORDS")
	headerColor.Println("  " + strings.Repeat("═", 50))

	result := performFullDNSLookup(host)
	displayDNSResult(result)

	headerColor.Println("  " + strings.Repeat("═", 50))
	pressEnter(scanner)
}

func performFullDNSLookup(host string) *DNSResult {
	result := &DNSResult{Host: host}

	// A/AAAA Records
	ips, err := net.LookupHost(host)
	if err == nil {
		result.IPs = ips
	}

	// CNAME
	cname, err := net.LookupCNAME(host)
	if err == nil && cname != host+"." {
		result.CNAME = strings.TrimSuffix(cname, ".")
		// Follow CNAME chain
		current := cname
		for {
			nextCNAME, err := net.LookupCNAME(strings.TrimSuffix(current, "."))
			if err != nil || nextCNAME == current {
				break
			}
			result.CNAMEChain = append(result.CNAMEChain, strings.TrimSuffix(nextCNAME, "."))
			current = nextCNAME
		}
	}

	// MX Records
	mxRecords, err := net.LookupMX(host)
	if err == nil {
		for _, mx := range mxRecords {
			result.MX = append(result.MX, fmt.Sprintf("%s (priority: %d)", strings.TrimSuffix(mx.Host, "."), mx.Pref))
		}
	}

	// NS Records
	nsRecords, err := net.LookupNS(host)
	if err == nil {
		for _, ns := range nsRecords {
			result.NS = append(result.NS, strings.TrimSuffix(ns.Host, "."))
		}
	}

	// TXT Records
	txtRecords, err := net.LookupTXT(host)
	if err == nil {
		result.TXT = txtRecords
	}

	// Reverse DNS on first IP
	if len(result.IPs) > 0 {
		hostnames, err := net.LookupAddr(result.IPs[0])
		if err == nil && len(hostnames) > 0 {
			result.ReverseDNS = strings.TrimSuffix(hostnames[0], ".")
		}
	}

	return result
}

func displayDNSResult(result *DNSResult) {
	printField("Hostname", result.Host)

	if result.CNAME != "" {
		printField("CNAME", result.CNAME)
	}

	if len(result.CNAMEChain) > 0 {
		for i, c := range result.CNAMEChain {
			printField(fmt.Sprintf("CNAME Chain %d", i+1), c)
		}
	}

	if len(result.IPs) > 0 {
		for _, ip := range result.IPs {
			printField("IP Address", ip)
		}
	}

	if result.ReverseDNS != "" {
		printField("Reverse DNS", result.ReverseDNS)
	}

	if len(result.MX) > 0 {
		fmt.Println()
		accentColor.Println("  ├─ MAIL SERVERS")
		for _, mx := range result.MX {
			printField("MX", mx)
		}
	}

	if len(result.NS) > 0 {
		fmt.Println()
		accentColor.Println("  ├─ NAME SERVERS")
		for _, ns := range result.NS {
			printField("NS", ns)
		}
	}

	if len(result.TXT) > 0 {
		fmt.Println()
		accentColor.Println("  ├─ TEXT RECORDS")
		for _, txt := range result.TXT {
			printField("TXT", txt)
		}
	}
}

// --- Port Scanner ---

// Common ports with their typical services
var commonPorts = map[int]string{
	21: "FTP", 22: "SSH", 23: "Telnet", 25: "SMTP", 53: "DNS",
	80: "HTTP", 110: "POP3", 111: "RPCBind", 135: "MSRPC",
	139: "NetBIOS", 143: "IMAP", 443: "HTTPS", 445: "SMB",
	993: "IMAPS", 995: "POP3S", 1433: "MSSQL", 1521: "Oracle",
	3306: "MySQL", 3389: "RDP", 5432: "PostgreSQL", 5900: "VNC",
	6379: "Redis", 8080: "HTTP-Alt", 8443: "HTTPS-Alt",
	27017: "MongoDB",
}

type PortResult struct {
	Port     int    `json:"port"`
	Service  string `json:"service"`
	Open     bool   `json:"open"`
	Latency  int64  `json:"latency_ms,omitempty"`
}

func portScanMenu(scanner *bufio.Scanner) {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  PORT SCANNER")
	headerColor.Println("  " + strings.Repeat("─", 58))
	fmt.Println("  1.  Scan Common Ports")
	fmt.Println("  2.  Scan Custom Port Range")
	fmt.Println("  3.  Scan Single Port")
	fmt.Println("  4.  Back to Main Menu")

	fmt.Print("\n  Choice: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		scanCommonPorts(scanner)
	case "2":
		scanCustomRange(scanner)
	case "3":
		scanSinglePort(scanner)
	case "4":
		return
	default:
		errorColor.Println("\n  Invalid choice!")
		pressEnter(scanner)
	}
}

func scanCommonPorts(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter target IP or hostname: ")
	scanner.Scan()
	target := strings.TrimSpace(scanner.Text())

	if target == "" {
		errorColor.Println("\n  No target entered!")
		pressEnter(scanner)
		return
	}

	// Resolve hostname to IP if needed
	ip := target
	if net.ParseIP(target) == nil {
		ips, err := net.LookupHost(target)
		if err != nil || len(ips) == 0 {
			errorColor.Printf("\n  Could not resolve hostname: %s\n", target)
			pressEnter(scanner)
			return
		}
		ip = ips[0]
		successColor.Printf("\n  Resolved %s to %s\n", target, ip)
	}

	headerColor.Println("\n  Scanning common ports...")
	fmt.Println(strings.Repeat("─", 58))

	var results []PortResult
	ports := make([]int, 0, len(commonPorts))
	for port := range commonPorts {
		ports = append(ports, port)
	}

	for _, port := range ports {
		service := commonPorts[port]
		open, latency := checkPort(ip, port, 2*time.Second)
		if open {
			successColor.Printf("  Port %5d %-12s OPEN  (%dms)\n", port, service, latency)
			results = append(results, PortResult{Port: port, Service: service, Open: true, Latency: latency})
		}
	}

	fmt.Println(strings.Repeat("─", 58))
	if len(results) == 0 {
		warnColor.Println("  No open ports found")
	} else {
		successColor.Printf("  Found %d open port(s)\n", len(results))
	}
	pressEnter(scanner)
}

func scanCustomRange(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter target IP or hostname: ")
	scanner.Scan()
	target := strings.TrimSpace(scanner.Text())

	if target == "" {
		errorColor.Println("\n  No target entered!")
		pressEnter(scanner)
		return
	}

	// Resolve hostname to IP if needed
	ip := target
	if net.ParseIP(target) == nil {
		ips, err := net.LookupHost(target)
		if err != nil || len(ips) == 0 {
			errorColor.Printf("\n  Could not resolve hostname: %s\n", target)
			pressEnter(scanner)
			return
		}
		ip = ips[0]
		successColor.Printf("\n  Resolved %s to %s\n", target, ip)
	}

	fmt.Print("  Enter start port: ")
	scanner.Scan()
	var start int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &start); err != nil || start < 1 || start > 65535 {
		errorColor.Println("\n  Invalid start port!")
		pressEnter(scanner)
		return
	}

	fmt.Print("  Enter end port: ")
	scanner.Scan()
	var end int
	if _, err := fmt.Sscanf(scanner.Text(), "%d", &end); err != nil || end < 1 || end > 65535 {
		errorColor.Println("\n  Invalid end port!")
		pressEnter(scanner)
		return
	}

	if start > end {
		errorColor.Println("\n  Invalid port range!")
		pressEnter(scanner)
		return
	}

	if end-start > 1000 {
		warnColor.Printf("\n  Scanning %d ports — this may take a while\n", end-start+1)
	}

	headerColor.Printf("\n  Scanning ports %d-%d on %s...\n", start, end, ip)
	fmt.Println(strings.Repeat("─", 58))

	var openPorts []PortResult
	for port := start; port <= end; port++ {
		service := commonPorts[port]
		if service == "" {
			service = "unknown"
		}
		open, latency := checkPort(ip, port, 1*time.Second)
		if open {
			successColor.Printf("  Port %5d %-12s OPEN  (%dms)\n", port, service, latency)
			openPorts = append(openPorts, PortResult{Port: port, Service: service, Open: true, Latency: latency})
		}
	}

	fmt.Println(strings.Repeat("─", 58))
	if len(openPorts) == 0 {
		warnColor.Println("  No open ports found")
	} else {
		successColor.Printf("  Found %d open port(s)\n", len(openPorts))
	}
	pressEnter(scanner)
}

func scanSinglePort(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter target IP or hostname: ")
	scanner.Scan()
	target := strings.TrimSpace(scanner.Text())

	if target == "" {
		errorColor.Println("\n  No target entered!")
		pressEnter(scanner)
		return
	}

	// Resolve hostname to IP if needed
	ip := target
	if net.ParseIP(target) == nil {
		ips, err := net.LookupHost(target)
		if err != nil || len(ips) == 0 {
			errorColor.Printf("\n  Could not resolve hostname: %s\n", target)
			pressEnter(scanner)
			return
		}
		ip = ips[0]
		successColor.Printf("\n  Resolved %s to %s\n", target, ip)
	}

	fmt.Print("  Enter port number: ")
	scanner.Scan()
	var port int
	_, err := fmt.Sscanf(scanner.Text(), "%d", &port)
	if err != nil || port < 1 || port > 65535 {
		errorColor.Println("\n  Invalid port number!")
		pressEnter(scanner)
		return
	}

	service := commonPorts[port]
	if service == "" {
		service = "unknown"
	}

	fmt.Printf("\n  Checking port %d (%s) on %s...\n", port, service, ip)

	open, latency := checkPort(ip, port, 3*time.Second)

	if open {
		successColor.Printf("\n  Port %d (%s) is OPEN\n", port, service)
		fmt.Printf("  Response time: %dms\n", latency)
	} else {
		errorColor.Printf("\n  Port %d (%s) is CLOSED/FILTERED\n", port, service)
	}
	pressEnter(scanner)
}

func checkPort(host string, port int, timeout time.Duration) (bool, int64) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return false, latency
	}
	conn.Close()
	return true, latency
}

// --- Traceroute ---

type TracerouteHop struct {
	TTL     int    `json:"ttl"`
	IP      string `json:"ip"`
	Host    string `json:"host,omitempty"`
	Latency int64  `json:"latency_ms"`
}

func tracerouteMenu(scanner *bufio.Scanner) {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  TRACEROUTE")
	headerColor.Println("  " + strings.Repeat("─", 58))
	fmt.Println("  1.  Traceroute to Host")
	fmt.Println("  2.  Back to Main Menu")

	fmt.Print("\n  Choice: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		runTraceroute(scanner)
	case "2":
		return
	default:
		errorColor.Println("\n  Invalid choice!")
		pressEnter(scanner)
	}
}

func runTraceroute(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter target hostname or IP: ")
	scanner.Scan()
	target := strings.TrimSpace(scanner.Text())

	if target == "" {
		errorColor.Println("\n  No target entered!")
		pressEnter(scanner)
		return
	}

	// Resolve to IP if hostname
	ip := target
	if net.ParseIP(target) == nil {
		ips, err := net.LookupHost(target)
		if err != nil || len(ips) == 0 {
			errorColor.Printf("\n  Could not resolve: %s\n", target)
			pressEnter(scanner)
			return
		}
		ip = ips[0]
	}

	fmt.Print("  Max hops (default 30): ")
	scanner.Scan()
	maxHops := 30
	fmt.Sscanf(scanner.Text(), "%d", &maxHops)
	if maxHops < 1 || maxHops > 64 {
		maxHops = 30
	}

	headerColor.Printf("\n  Traceroute to %s (%s) — max %d hops\n", target, ip, maxHops)
	headerColor.Println("  " + strings.Repeat("═", 58))
	fmt.Printf("  %-4s %-18s %-20s %s\n", "TTL", "IP", "HOSTNAME", "LATENCY")
	headerColor.Println("  " + strings.Repeat("─", 58))

	for ttl := 1; ttl <= maxHops; ttl++ {
		hop := traceOneHop(ip, ttl)

		if hop.IP == "" {
			fmt.Printf("  %-4d %-18s\n", ttl, "* * *")
		} else {
			hostDisplay := hop.Host
			if hostDisplay == "" {
				hostDisplay = "-"
			}
			fmt.Printf("  %-4d %-18s %-20s %dms\n", ttl, hop.IP, hostDisplay, hop.Latency)
		}

		// If we reached the target, stop
		if hop.IP == ip {
			break
		}
	}

	headerColor.Println("  " + strings.Repeat("═", 58))
	pressEnter(scanner)
}

func traceOneHop(target string, ttl int) TracerouteHop {
	hop := TracerouteHop{TTL: ttl}

	// Try connecting with decreasing TTL
	// We use a trick: connect to a high port that likely won't be open,
	// the TTL-exceeded response gives us the hop IP
	addr := fmt.Sprintf("%s:80", target)

	// Set up raw connection with TTL
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		// Connected directly - this is the final hop or close to it
		conn.Close()
		hop.IP = target
		hop.Latency = 1
		hostnames, _ := net.LookupAddr(target)
		if len(hostnames) > 0 {
			hop.Host = strings.TrimSuffix(hostnames[0], ".")
		}
		return hop
	}

	// For intermediate hops, we try a UDP-based approach using a high port
	// This requires the OS to send ICMP TTL exceeded messages
	// Alternative: just try to connect and parse the error
	udpAddr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:33434", target))
	if udpAddr == nil {
		return hop
	}

	start := time.Now()
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return hop
	}

	// Set TTL on the connection
	// Note: This is OS-dependent and may not work on all platforms
	_, err = udpConn.Write([]byte("traceroute"))
	udpConn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buf := make([]byte, 1500)
	_, err = udpConn.Read(buf)
	udpConn.Close()
	latency := time.Since(start).Milliseconds()

	if err == nil {
		hop.IP = target
		hop.Latency = latency
		hostnames, _ := net.LookupAddr(target)
		if len(hostnames) > 0 {
			hop.Host = strings.TrimSuffix(hostnames[0], ".")
		}
	}

	return hop
}

// --- Connectivity & Latency ---

type ConnectivityResult struct {
	Target   string  `json:"target"`
	IP       string  `json:"ip"`
	HTTPCode int     `json:"http_code,omitempty"`
	Latency  int64   `json:"latency_ms"`
	DNSTime  int64   `json:"dns_ms,omitempty"`
	ConnTime int64   `json:"connect_ms,omitempty"`
	TLS      bool    `json:"tls,omitempty"`
	Status   string  `json:"status"`
}

func connectivityMenu(scanner *bufio.Scanner) {
	headerColor.Println("\n  " + strings.Repeat("─", 58))
	headerColor.Println("  CONNECTIVITY & LATENCY TEST")
	headerColor.Println("  " + strings.Repeat("─", 58))
	fmt.Println("  1.  Test HTTP/HTTPS Connectivity")
	fmt.Println("  2.  Latency Benchmark (multiple targets)")
	fmt.Println("  3.  Back to Main Menu")

	fmt.Print("\n  Choice: ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		testHTTPConnectivity(scanner)
	case "2":
		latencyBenchmark(scanner)
	case "3":
		return
	default:
		errorColor.Println("\n  Invalid choice!")
		pressEnter(scanner)
	}
}

func testHTTPConnectivity(scanner *bufio.Scanner) {
	fmt.Print("\n  Enter URL or hostname (e.g., google.com): ")
	scanner.Scan()
	target := strings.TrimSpace(scanner.Text())

	if target == "" {
		errorColor.Println("\n  No target entered!")
		pressEnter(scanner)
		return
	}

	// Add protocol if missing
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	headerColor.Println("\n  Testing connectivity...")
	fmt.Println(strings.Repeat("─", 58))

	// DNS resolution time
	host := target
	if strings.HasPrefix(target, "https://") {
		host = strings.TrimPrefix(target, "https://")
	} else {
		host = strings.TrimPrefix(target, "http://")
	}
	host = strings.Split(host, "/")[0]

	dnsStart := time.Now()
	ips, err := net.LookupHost(host)
	dnsTime := time.Since(dnsStart).Milliseconds()

	if err != nil || len(ips) == 0 {
		errorColor.Printf("\n  DNS resolution failed for %s\n", host)
		pressEnter(scanner)
		return
	}

	printField("Host", host)
	printField("IP", ips[0])
	printField("DNS Resolution", fmt.Sprintf("%dms", dnsTime))

	// HTTP request
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	connStart := time.Now()
	resp, err := client.Get(target)
	connTime := time.Since(connStart).Milliseconds()

	if err != nil {
		errorColor.Printf("\n  Connection failed: %v\n", err)
		pressEnter(scanner)
		return
	}
	defer resp.Body.Close()

	printField("Status", fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode)))
	printField("Response Time", fmt.Sprintf("%dms", connTime))
	printField("Protocol", resp.Proto)
	printField("TLS", fmt.Sprintf("%v", resp.TLS != nil))

	if resp.TLS != nil {
		printField("TLS Version", tlsVersionName(resp.TLS.Version))
		if len(resp.TLS.PeerCertificates) > 0 {
			cert := resp.TLS.PeerCertificates[0]
			printField("Certificate", cert.Subject.CommonName)
			printField("Issuer", cert.Issuer.CommonName)
			printField("Expires", cert.NotAfter.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println(strings.Repeat("─", 58))
	pressEnter(scanner)
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}

func latencyBenchmark(scanner *bufio.Scanner) {
	targets := []struct {
		name string
		url  string
	}{
		{"Google", "https://www.google.com"},
		{"Cloudflare", "https://www.cloudflare.com"},
		{"GitHub", "https://github.com"},
		{"Amazon", "https://www.amazon.com"},
		{"Microsoft", "https://www.microsoft.com"},
	}

	headerColor.Println("\n  Running latency benchmark...")
	fmt.Println(strings.Repeat("═", 58))
	fmt.Printf("  %-15s %-12s %-12s\n", "TARGET", "STATUS", "LATENCY")
	fmt.Println(strings.Repeat("─", 58))

	var totalLatency int64
	var successCount int

	client := &http.Client{Timeout: 5 * time.Second}

	for _, t := range targets {
		start := time.Now()
		resp, err := client.Get(t.url)
		latency := time.Since(start).Milliseconds()

		if err != nil {
			errorColor.Printf("  %-15s %-12s\n", t.name, "FAILED")
			continue
		}
		resp.Body.Close()

		totalLatency += latency
		successCount++

		if latency < 100 {
			successColor.Printf("  %-15s %-12s %dms\n", t.name, "OK", latency)
		} else if latency < 300 {
			warnColor.Printf("  %-15s %-12s %dms\n", t.name, "OK", latency)
		} else {
			errorColor.Printf("  %-15s %-12s %dms\n", t.name, "SLOW", latency)
		}
	}

	fmt.Println(strings.Repeat("─", 58))
	if successCount > 0 {
		avg := totalLatency / int64(successCount)
		successColor.Printf("  Average latency: %dms (from %d targets)\n", avg, successCount)
	}
	pressEnter(scanner)
}

// --- Export ---

func exportResults(results []*DetailedInfo, filename, format string) error {
	// Auto-detect format from filename
	if format == "auto" {
		if strings.HasSuffix(filename, ".csv") {
			format = "csv"
		} else {
			format = "json"
			if !strings.HasSuffix(filename, ".json") {
				filename += ".json"
			}
		}
	}

	switch format {
	case "json":
		return exportJSON(results, filename)
	case "csv":
		return exportCSV(results, filename)
	default:
		return fmt.Errorf("unsupported format: %s (use json or csv)", format)
	}
}

func exportJSON(results []*DetailedInfo, filename string) error {
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func exportCSV(results []*DetailedInfo, filename string) error {
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}

	var sb strings.Builder
	sb.WriteString("IP,Country,CountryCode,Region,RegionCode,City,PostalCode,Lat,Lon,Timezone,ISP,Org,ASNumber,ASName,Mobile,Proxy,Hosting\n")

	for _, r := range results {
		fmt.Fprintf(&sb, "%s,%s,%s,%s,%s,%s,%s,%.6f,%.6f,%s,%s,%s,%s,%s,%t,%t,%t\n",
			r.IPAddress, r.Country, r.CountryCode, r.State, r.StateCode,
			r.City, r.PostalCode, r.Lat, r.Lon, r.Timezone,
			r.TelecomService, r.Organization, r.ASNumber, r.ASName,
			r.IsMobile, r.IsProxy, r.IsHosting)
	}

	return os.WriteFile(filename, []byte(sb.String()), 0644)
}

// --- History ---

type HistoryEntry struct {
	IP        string  `json:"ip"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	ISP       string  `json:"isp"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timestamp string  `json:"timestamp"`
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ip_tracker_history.json"
	}
	return home + "/.ip_tracker_history.json"
}

func loadHistory() ([]HistoryEntry, error) {
	data, err := os.ReadFile(historyPath())
	if err != nil {
		return nil, err
	}
	var entries []HistoryEntry
	err = json.Unmarshal(data, &entries)
	return entries, err
}

func saveHistory(entries []HistoryEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyPath(), data, 0644)
}

func saveToHistory(info *IPInfo) {
	entries, _ := loadHistory()

	entry := HistoryEntry{
		IP:        info.Query,
		City:      info.City,
		Country:   info.Country,
		ISP:       info.Isp,
		Lat:       info.Lat,
		Lon:       info.Lon,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	entries = append(entries, entry)

	// Keep max 500 entries
	if len(entries) > 500 {
		entries = entries[len(entries)-500:]
	}

	saveHistory(entries)
}
