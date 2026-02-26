package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	QBittorrentURL  string
	Username        string
	Password        string
	PortFile        string
	CheckInterval   time.Duration
}

type QBittorrentClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	sid        string
}

func loadConfig() (*Config, error) {
	qbURL := getEnv("QBITTORRENT_URL", "http://localhost:30024")
	username := getEnv("QBITTORRENT_USERNAME", "admin")
	password := os.Getenv("QBITTORRENT_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("QBITTORRENT_PASSWORD environment variable is required")
	}
	
	portFile := getEnv("PORT_FILE", "/tmp/gluetun/forwarded_port")
	checkInterval := getEnvInt("CHECK_INTERVAL", 30)

	return &Config{
		QBittorrentURL:  qbURL,
		Username:        username,
		Password:        password,
		PortFile:        portFile,
		CheckInterval:   time.Duration(checkInterval) * time.Second,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func NewQBittorrentClient(baseURL, username, password string) (*QBittorrentClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &QBittorrentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
		username: username,
		password: password,
	}, nil
}

func (c *QBittorrentClient) Login() error {
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", c.baseURL)
	
	data := url.Values{}
	data.Set("username", c.username)
	data.Set("password", c.password)

	resp, err := c.httpClient.PostForm(loginURL, data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(body))

	if resp.StatusCode != http.StatusOK || bodyStr != "Ok." {
		return fmt.Errorf("login failed: status=%d, body=%s", resp.StatusCode, bodyStr)
	}

	log.Println("Successfully authenticated with qBittorrent")
	return nil
}

func (c *QBittorrentClient) GetListeningPort() (int, error) {
	prefsURL := fmt.Sprintf("%s/api/v2/app/preferences", c.baseURL)
	
	resp, err := c.httpClient.Get(prefsURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get preferences: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return 0, fmt.Errorf("authentication expired")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var prefs map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return 0, fmt.Errorf("failed to decode preferences: %w", err)
	}

	port, ok := prefs["listen_port"].(float64)
	if !ok {
		return 0, fmt.Errorf("listen_port not found in preferences")
	}

	return int(port), nil
}

func (c *QBittorrentClient) SetListeningPort(port int) error {
	setPrefsURL := fmt.Sprintf("%s/api/v2/app/setPreferences", c.baseURL)
	
	prefs := map[string]interface{}{
		"listen_port": port,
	}
	
	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	data := url.Values{}
	data.Set("json", string(prefsJSON))

	resp, err := c.httpClient.PostForm(setPrefsURL, data)
	if err != nil {
		return fmt.Errorf("failed to set preferences: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication expired")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func readPortFile(filename string) (int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to read port file: %w", err)
	}

	portStr := strings.TrimSpace(string(data))
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port number out of range: %d", port)
	}

	return port, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("qBittorrent Port Sync starting...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded:")
	log.Printf("  qBittorrent URL: %s", config.QBittorrentURL)
	log.Printf("  Username: %s", config.Username)
	log.Printf("  Port file: %s", config.PortFile)
	log.Printf("  Check interval: %v", config.CheckInterval)

	client, err := NewQBittorrentClient(config.QBittorrentURL, config.Username, config.Password)
	if err != nil {
		log.Fatalf("Failed to create qBittorrent client: %v", err)
	}

	// Initial login
	if err := client.Login(); err != nil {
		log.Fatalf("Initial login failed: %v", err)
	}

	// Wait for port file to exist
	log.Printf("Waiting for port file: %s", config.PortFile)
	for {
		if _, err := os.Stat(config.PortFile); err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	log.Println("Port file found, starting sync loop...")

	var lastPort int

	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	// Do initial sync immediately
	syncPort(client, config.PortFile, &lastPort)

	for range ticker.C {
		syncPort(client, config.PortFile, &lastPort)
	}
}

func syncPort(client *QBittorrentClient, portFile string, lastPort *int) {
	// Read port from file
	filePort, err := readPortFile(portFile)
	if err != nil {
		log.Printf("Error reading port file: %v", err)
		return
	}

	// Check if port has changed
	if filePort == *lastPort {
		log.Printf("Port unchanged: %d", filePort)
		return
	}

	log.Printf("Port changed from %d to %d, updating qBittorrent...", *lastPort, filePort)

	// Get current port from qBittorrent
	currentPort, err := client.GetListeningPort()
	if err != nil {
		if strings.Contains(err.Error(), "authentication expired") {
			log.Println("Session expired, re-authenticating...")
			if err := client.Login(); err != nil {
				log.Printf("Re-authentication failed: %v", err)
				return
			}
			// Retry getting current port
			currentPort, err = client.GetListeningPort()
			if err != nil {
				log.Printf("Failed to get current port after re-auth: %v", err)
				return
			}
		} else {
			log.Printf("Failed to get current port: %v", err)
			return
		}
	}

	log.Printf("qBittorrent current port: %d", currentPort)

	// Update if different
	if currentPort != filePort {
		if err := client.SetListeningPort(filePort); err != nil {
			if strings.Contains(err.Error(), "authentication expired") {
				log.Println("Session expired during set, re-authenticating...")
				if err := client.Login(); err != nil {
					log.Printf("Re-authentication failed: %v", err)
					return
				}
				// Retry setting port
				if err := client.SetListeningPort(filePort); err != nil {
					log.Printf("Failed to set port after re-auth: %v", err)
					return
				}
			} else {
				log.Printf("Failed to set listening port: %v", err)
				return
			}
		}
		log.Printf("✓ Successfully updated qBittorrent listening port to %d", filePort)
	} else {
		log.Printf("qBittorrent already configured with correct port: %d", filePort)
	}

	*lastPort = filePort
}
