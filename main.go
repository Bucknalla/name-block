package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	balenaAPIBase = "https://api.balena-cloud.com"
	version       = "0.1.0"
)

var debug bool

func debugLog(format string, args ...interface{}) {
	if debug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func getMonth() string {
	month := time.Now().Month()
	switch {
	case month >= 11 || month < 2:
		return "winter"
	case month >= 2 && month < 5:
		return "spring"
	case month >= 5 && month < 8:
		return "summer"
	case month >= 8 && month < 11:
		return "autumn"
	default:
		return "winter"
	}
}

func randomWord(words string) string {
	lines := strings.Split(strings.TrimSpace(words), "\n")
	if len(lines) == 0 {
		return ""
	}
	// Filter out empty lines
	nonEmpty := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmpty = append(nonEmpty, trimmed)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	return nonEmpty[rand.Intn(len(nonEmpty))]
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func httpRequest(url string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("statusCode=%d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func getURLs() ([]string, error) {
	urlEnv := os.Getenv("URL")
	if urlEnv == "" {
		return nil, fmt.Errorf("URL environment variable not set")
	}
	return strings.Split(urlEnv, ","), nil
}

func generateName(category string) (string, error) {
	if category != "url" {
		adjPath := fmt.Sprintf("data/%s/adj.txt", category)
		nounPath := fmt.Sprintf("data/%s/noun.txt", category)

		adj, err := readFile(adjPath)
		if err != nil {
			return "", fmt.Errorf("failed to read adjective file: %w", err)
		}

		noun, err := readFile(nounPath)
		if err != nil {
			return "", fmt.Errorf("failed to read noun file: %w", err)
		}

		adjWord := randomWord(adj)
		nounWord := randomWord(noun)
		return fmt.Sprintf("%s-%s", adjWord, nounWord), nil
	}

	// URL-based generation
	urls, err := getURLs()
	if err != nil {
		return "", fmt.Errorf("failed to get URLs: %w", err)
	}

	noun, err := httpRequest(urls[0])
	if err != nil {
		return "", fmt.Errorf("failed to fetch noun from URL: %w", err)
	}
	nounWord := randomWord(noun)

	if len(urls) > 1 && urls[1] != "" {
		adj, err := httpRequest(urls[1])
		if err != nil {
			return "", fmt.Errorf("failed to fetch adjective from URL: %w", err)
		}
		adjWord := randomWord(adj)
		return fmt.Sprintf("%s-%s", adjWord, nounWord), nil
	}

	return nounWord, nil
}

type DeviceEnvVar struct {
	Value string `json:"value"`
}

func getDeviceEnvVar(apiKey, deviceUUID, varName string) (string, error) {
	url := fmt.Sprintf("%s/v6/device_environment_variable?$filter=device/uuid%%20eq%%20'%s'%%20and%%20name%%20eq%%20'%s'", balenaAPIBase, deviceUUID, varName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", nil // Variable doesn't exist
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get env var: statusCode=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		D []DeviceEnvVar `json:"d"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.D) > 0 {
		return result.D[0].Value, nil
	}

	return "", nil
}

func setDeviceEnvVar(apiKey, deviceUUID, varName, value string) error {
	// First, try to get the device ID from UUID
	deviceID, err := getDeviceID(apiKey, deviceUUID)
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	url := fmt.Sprintf("%s/v6/device_environment_variable", balenaAPIBase)

	payload := map[string]interface{}{
		"device": deviceID,
		"name":   varName,
		"value":  value,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set env var: statusCode=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func getDeviceID(apiKey, deviceUUID string) (int, error) {
	url := fmt.Sprintf("%s/v6/device?$filter=uuid%%20eq%%20'%s'", balenaAPIBase, deviceUUID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to get device: statusCode=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		D []struct {
			ID int `json:"id"`
		} `json:"d"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if len(result.D) == 0 {
		return 0, fmt.Errorf("device not found")
	}

	return result.D[0].ID, nil
}

func renameDevice(apiKey, deviceUUID, name string) error {
	deviceID, err := getDeviceID(apiKey, deviceUUID)
	if err != nil {
		return fmt.Errorf("failed to get device ID: %w", err)
	}

	url := fmt.Sprintf("%s/v6/device(%d)", balenaAPIBase, deviceID)

	payload := map[string]interface{}{
		"device_name": name,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to rename device: statusCode=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func rename(apiKey, deviceUUID, name string) error {
	existingValue, err := getDeviceEnvVar(apiKey, deviceUUID, "NAME")
	if err != nil {
		return fmt.Errorf("failed to check existing NAME env var: %w", err)
	}

	if existingValue != "" {
		debugLog("failed, device name already set")
		return nil
	}

	debugLog("setting device envar to %s", name)
	if err := setDeviceEnvVar(apiKey, deviceUUID, "NAME", name); err != nil {
		return fmt.Errorf("failed to set NAME env var: %w", err)
	}

	debugLog("generated name: %s", name)
	if err := renameDevice(apiKey, deviceUUID, name); err != nil {
		return fmt.Errorf("failed to rename device: %w", err)
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	debug = os.Getenv("DEBUG") == "name"
	apiKey := os.Getenv("BALENA_API_KEY")
	deviceUUID := os.Getenv("BALENA_DEVICE_UUID")
	category := os.Getenv("CATEGORY")
	if category == "" {
		category = "season"
	}

	debugLog("name-block: v%s", version)

	if category == "" {
		log.Fatal("CATEGORY environment variable is required")
	}

	switch category {
	case "season":
		category = getMonth()
	case "url":
		// Keep as is
	case "rude", "animals", "pirate", "autumn", "spring", "summer", "winter":
		// Keep as is
	default:
		category = "animals"
	}

	debugLog("category set to %s", category)

	name, err := generateName(category)
	if err != nil {
		log.Fatalf("failed to generate name: %v", err)
	}

	name = strings.ToLower(name)
	if err := rename(apiKey, deviceUUID, name); err != nil {
		log.Fatalf("failed to rename device: %v", err)
	}

	debugLog("Successfully renamed device to: %s", name)
}
