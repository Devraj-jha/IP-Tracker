package main

import (
	"bufio"
	"encoding/json"
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
const appVersion = "2.0.0"

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
}

type DetailedInfo struct {
	IPAddress      string  `json:"ip_address"`
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
	displayBanner()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		displayMenu()
		fmt.Print("\n  Enter your choice (1-5): ")
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
			searchHistory()
		case "5", "exit", "EXIT", "Exit", "q", "Q":
			printExit()
			os.Exit(0)
		default:
			errorColor.Println("\n  Invalid choice! Please enter 1-5.")
			pressEnter(scanner)
		}
	}
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
	fmt.Println("  4.  View Search History")
	fmt.Println("  5.  Exit")
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

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,asname,mobile,proxy,hosting,query", ip)

	var resp *http.Response
	var err error

	// Retry up to 2 times
	for attempt := 0; attempt < 2; attempt++ {
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

// --- Export ---

func exportResults(results []*DetailedInfo, filename, format string) error {
	switch format {
	case "json":
		return exportJSON(results, filename)
	case "csv":
		return exportCSV(results, filename)
	default:
		return fmt.Errorf("unsupported format: %s", format)
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
