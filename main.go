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
	IPAddress      string
	Country        string
	CountryCode    string
	State          string
	StateCode      string
	City           string
	PostalCode     string
	Coordinates    string
	Timezone       string
	TelecomService string
	Organization   string
	ASNumber       string
	ASName         string
	IsMobile       bool
	IsProxy        bool
	IsHosting      bool
}

func main() {
	displayBanner()
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		displayMenu()
		fmt.Print("\n Enter your choice (1-3): ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())
		
		switch choice {
		case "1":
			trackMyIP()
		case "2":
			trackOtherIP(scanner)
		case "3", "exit", "EXIT", "Exit":
			fmt.Println("\nThank you for using IP Tracker! Have A nice day!")
			fmt.Println("=" + strings.Repeat("=", 58))
			os.Exit(0)
		default:
			fmt.Println("\n Invalid choice! Please enter 1, 2, or 3.")
			fmt.Print("\nPress Enter to continue...")
			scanner.Scan()
		}
	}
}

func displayBanner() {
	banner := `

██╗██████╗░░░░░░░████████╗██████╗░░█████╗░░█████╗░██╗░░██╗███████╗██████╗░
██║██╔══██╗░░░░░░╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██║░██╔╝██╔════╝██╔══██╗
██║██████╔╝█████╗░░░██║░░░██████╔╝███████║██║░░╚═╝█████═╝░█████╗░░██████╔╝
██║██╔═══╝░╚════╝░░░██║░░░██╔══██╗██╔══██║██║░░██╗██╔═██╗░██╔══╝░░██╔══██╗
██║██║░░░░░░░░░░░░░░██║░░░██║░░██║██║░░██║╚█████╔╝██║░╚██╗███████╗██║░░██║
╚═╝╚═╝░░░░░░░░░░░░░░╚═╝░░░╚═╝░░╚═╝╚═╝░░╚═╝░╚════╝░╚═╝░░╚═╝╚══════╝╚═╝░░╚═╝
`
	fmt.Println(banner)
}

func displayMenu() {
	fmt.Println("\n" + strings.Repeat("─", 62))
	fmt.Println("=> MAIN MENU <=")
	fmt.Println(strings.Repeat("─", 62))
	fmt.Println("  1. Track MY IP Address (Current User)")
	fmt.Println("  2. Track ANOTHER IP Address (Enter manually)")
	fmt.Println("  3. Exit Program")
	fmt.Println(strings.Repeat("─", 62))
}

func trackMyIP() {
	fmt.Println("\n" + strings.Repeat("─", 62))
	fmt.Println("🖥️  TRACKING YOUR IP ADDRESS...")
	fmt.Println(strings.Repeat("─", 62))
	
	publicIP, err := getPublicIP()
	if err != nil {
		fmt.Printf("\n Error getting your public IP: %v\n", err)
		fmt.Print("\nPress Enter to continue...")
		bufio.NewScanner(os.Stdin).Scan()
		return
	}
	
	fmt.Printf("\nYour Public IP Address: %s\n", publicIP)
	fmt.Println("\n Fetching geolocation data...")
	
	info, err := fetchIPInfo(publicIP)
	if err != nil {
		fmt.Printf("\n Error fetching IP information: %v\n", err)
		fmt.Print("\nPress Enter to continue...")
		bufio.NewScanner(os.Stdin).Scan()
		return
	}
	
	displayInfo(info)
	
	fmt.Print("\nPress Enter to continue...")
	bufio.NewScanner(os.Stdin).Scan()
}

func trackOtherIP(scanner *bufio.Scanner) {
	fmt.Println("\n" + strings.Repeat("─", 62))
	fmt.Println("🌐 TRACK ANOTHER IP ADDRESS")
	fmt.Println(strings.Repeat("─", 62))
	
	fmt.Print("\n Enter IP Address (IPv4 or IPv6): ")
	scanner.Scan()
	ipAddress := strings.TrimSpace(scanner.Text())
	
	if ipAddress == "" {
		fmt.Println("\n No IP address entered!")
		fmt.Print("\nPress Enter to continue...")
		scanner.Scan()
		return
	}
	
	if net.ParseIP(ipAddress) == nil {
		fmt.Printf("\nInvalid IP address format: %s\n", ipAddress)
		fmt.Println("   Please enter a valid IPv4 (e.g., 8.8.8.8) or IPv6 address")
		fmt.Print("\nPress Enter to continue...")
		scanner.Scan()
		return
	}
	
	fmt.Printf("\n Target IP Address: %s\n", ipAddress)
	fmt.Println("\n🔍 Fetching geolocation data...")
	
	info, err := fetchIPInfo(ipAddress)
	if err != nil {
		fmt.Printf("\nError fetching IP information: %v\n", err)
		fmt.Print("\nPress Enter to continue...")
		scanner.Scan()
		return
	}
	
	displayInfo(info)
	
	fmt.Print("\nPress Enter to continue...")
	scanner.Scan()
}

func getPublicIP() (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	
	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}

func fetchIPInfo(ip string) (*IPInfo, error) {
	client := http.Client{Timeout: 10 * time.Second}
	
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,asname,mobile,proxy,hosting,query", ip)
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
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

func displayInfo(info *IPInfo) {
	detailed := DetailedInfo{
		IPAddress:      info.Query,
		Country:        info.Country,
		CountryCode:    info.CountryCode,
		State:          info.RegionName,
		StateCode:      info.Region,
		City:           info.City,
		PostalCode:     info.Zip,
		Coordinates:    fmt.Sprintf("%.6f, %.6f", info.Lat, info.Lon),
		Timezone:       info.Timezone,
		TelecomService: info.Isp,
		Organization:   info.Org,
		ASNumber:       info.As,
		ASName:         info.ASName,
		IsMobile:       info.Mobile,
		IsProxy:        info.Proxy,
		IsHosting:      info.Hosting,
	}
	
	fmt.Println("\n" + strings.Repeat("═", 66))
	fmt.Println(" GEOLOCATION & NETWORK INTELLIGENCE REPORT")
	fmt.Println(strings.Repeat("═", 66))
	fmt.Println()
	
	fmt.Println(" ┌─ IP ADDRESS INFORMATION")
	fmt.Println(" │")
	fmt.Printf(" │   🌐 IP Address:      %s\n", detailed.IPAddress)
	fmt.Println(" │")
	
	fmt.Println(" ┌─ LOCATION DETAILS")
	fmt.Println(" │")
	fmt.Printf(" │    => Country:         %s (%s)\n", detailed.Country, detailed.CountryCode)
	fmt.Printf(" │    => State/Region:    %s (%s)\n", detailed.State, detailed.StateCode)
	fmt.Printf(" │    => City:            %s\n", detailed.City)
	fmt.Printf(" │    => Postal Code:     %s\n", detailed.PostalCode)
	fmt.Printf(" │    => Coordinates:     %s\n", detailed.Coordinates)
	fmt.Printf(" │    => Timezone:        %s\n", detailed.Timezone)
	fmt.Println(" │")
	
	fmt.Println(" ┌─ NETWORK & TELECOM SERVICE")
	fmt.Println(" │")
	fmt.Printf(" │       ISP:             %s\n", detailed.TelecomService)
	fmt.Printf(" │       Organization:    %s\n", detailed.Organization)
	fmt.Printf(" │       AS Number:       %s\n", detailed.ASNumber)
	fmt.Printf(" │       AS Name:         %s\n", detailed.ASName)
	fmt.Println(" │")
	
	fmt.Println(" ┌─ ADDITIONAL INTELLIGENCE")
	fmt.Println(" │")
	fmt.Printf(" │       Mobile Network:  %s\n", boolToSymbol(detailed.IsMobile))
	fmt.Printf(" │       Proxy/VPN:       %s\n", boolToSymbol(detailed.IsProxy))
	fmt.Printf(" │       Hosting/Cloud:   %s\n", boolToSymbol(detailed.IsHosting))
	fmt.Println(" │")
	
	fmt.Println(" └─────────────────────────────────────────────────────────")
	fmt.Println(strings.Repeat("═", 66))
	
	if info.Lat != 0 && info.Lon != 0 {
		fmt.Printf("\nGoogle Maps: https://maps.google.com/?q=%.6f,%.6f\n", info.Lat, info.Lon)
		fmt.Println(strings.Repeat("─", 66))
	}
}

func boolToSymbol(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}