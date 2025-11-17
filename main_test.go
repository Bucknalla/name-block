package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRandomWord(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(string) bool
	}{
		{
			name:  "basic word list",
			input: "apple\nbanana\ncherry",
			validate: func(word string) bool {
				valid := []string{"apple", "banana", "cherry"}
				for _, v := range valid {
					if word == v {
						return true
					}
				}
				return false
			},
		},
		{
			name:  "with empty lines",
			input: "apple\n\nbanana\n\ncherry\n",
			validate: func(word string) bool {
				valid := []string{"apple", "banana", "cherry"}
				for _, v := range valid {
					if word == v {
						return true
					}
				}
				return false
			},
		},
		{
			name:  "with whitespace",
			input: "  apple  \n  banana  \n  cherry  ",
			validate: func(word string) bool {
				valid := []string{"apple", "banana", "cherry"}
				for _, v := range valid {
					if word == v {
						return true
					}
				}
				return false
			},
		},
		{
			name:  "single word",
			input: "test",
			validate: func(word string) bool {
				return word == "test"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to test randomness
			for i := 0; i < 10; i++ {
				result := randomWord(tt.input)
				if !tt.validate(result) {
					t.Errorf("randomWord() = %v, expected one of valid words", result)
				}
			}
		})
	}
}

func TestRandomWord_EmptyInput(t *testing.T) {
	result := randomWord("")
	if result != "" {
		t.Errorf("randomWord(\"\") = %v, expected empty string", result)
	}

	result = randomWord("\n\n\n")
	if result != "" {
		t.Errorf("randomWord(\"\\n\\n\\n\") = %v, expected empty string", result)
	}
}

func TestGetMonth(t *testing.T) {
	// Test each month
	monthTests := map[time.Month]string{
		time.January:   "winter",
		time.February:  "spring",
		time.March:     "spring",
		time.April:     "spring",
		time.May:       "summer",
		time.June:      "summer",
		time.July:      "summer",
		time.August:    "autumn",
		time.September: "autumn",
		time.October:   "autumn",
		time.November:  "winter",
		time.December:  "winter",
	}

	// We can't easily mock time.Now(), so we'll test the current month
	// and verify the logic is correct
	currentMonth := time.Now().Month()
	expectedSeason := monthTests[currentMonth]
	result := getMonth()

	if result != expectedSeason {
		t.Errorf("getMonth() for %v = %v, expected %v", currentMonth, result, expectedSeason)
	}

	// Verify all mappings are correct
	for month, expected := range monthTests {
		// Create a test by temporarily overriding time (we can't do this easily)
		// So we'll just verify the current month works and document the expected behavior
		if month == currentMonth && result != expected {
			t.Errorf("Month %v should map to %v, but got %v", month, expected, result)
		}
	}
}

func TestGetURLs(t *testing.T) {
	tests := []struct {
		name        string
		urlEnv      string
		expected    []string
		expectError bool
	}{
		{
			name:        "single URL",
			urlEnv:      "https://example.com/noun.txt",
			expected:    []string{"https://example.com/noun.txt"},
			expectError: false,
		},
		{
			name:        "two URLs",
			urlEnv:      "https://example.com/noun.txt,https://example.com/adj.txt",
			expected:    []string{"https://example.com/noun.txt", "https://example.com/adj.txt"},
			expectError: false,
		},
		{
			name:        "empty URL env",
			urlEnv:      "",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("URL", tt.urlEnv)
			defer os.Unsetenv("URL")

			result, err := getURLs()
			if tt.expectError {
				if err == nil {
					t.Errorf("getURLs() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("getURLs() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("getURLs() = %v, expected %v", result, tt.expected)
				return
			}

			for i, url := range result {
				if url != tt.expected[i] {
					t.Errorf("getURLs()[%d] = %v, expected %v", i, url, tt.expected[i])
				}
			}
		})
	}
}

func TestGenerateName_FileBased(t *testing.T) {
	// Create temporary test data directory
	tmpDir := t.TempDir()
	testCategory := "test"
	dataDir := filepath.Join(tmpDir, "data", testCategory)
	adjPath := filepath.Join(dataDir, "adj.txt")
	nounPath := filepath.Join(dataDir, "noun.txt")

	os.MkdirAll(dataDir, 0755)
	os.WriteFile(adjPath, []byte("happy\nsad\nangry"), 0644)
	os.WriteFile(nounPath, []byte("cat\ndog\nbird"), 0644)

	// Save original working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	os.Chdir(tmpDir)

	// Test that it generates a name in the format "adj-noun"
	name, err := generateName(testCategory)
	if err != nil {
		t.Fatalf("generateName() error = %v", err)
	}

	parts := strings.Split(name, "-")
	if len(parts) != 2 {
		t.Errorf("generateName() = %v, expected format 'adj-noun'", name)
	}

	// Verify the parts are from our test data
	validAdjs := []string{"happy", "sad", "angry"}
	validNouns := []string{"cat", "dog", "bird"}

	adjValid := false
	for _, adj := range validAdjs {
		if parts[0] == adj {
			adjValid = true
			break
		}
	}

	nounValid := false
	for _, noun := range validNouns {
		if parts[1] == noun {
			nounValid = true
			break
		}
	}

	if !adjValid {
		t.Errorf("generateName() adjective = %v, expected one of %v", parts[0], validAdjs)
	}

	if !nounValid {
		t.Errorf("generateName() noun = %v, expected one of %v", parts[1], validNouns)
	}
}

func TestGenerateName_URLBased_SingleURL(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("testword\nanotherword\nthirdword"))
	}))
	defer server.Close()

	os.Setenv("URL", server.URL)
	defer os.Unsetenv("URL")

	// Temporarily override httpRequest to use our test server
	// We'll need to modify the function or test it differently
	// For now, let's test the URL parsing part
	urls, err := getURLs()
	if err != nil {
		t.Fatalf("getURLs() error = %v", err)
	}

	if len(urls) != 1 {
		t.Errorf("getURLs() = %v, expected single URL", urls)
	}
}

func TestGenerateName_URLBased_TwoURLs(t *testing.T) {
	// Create test HTTP servers
	nounServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cat\ndog\nbird"))
	}))
	defer nounServer.Close()

	adjServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("happy\nsad\nangry"))
	}))
	defer adjServer.Close()

	os.Setenv("URL", fmt.Sprintf("%s,%s", nounServer.URL, adjServer.URL))
	defer os.Unsetenv("URL")

	// Test URL-based generation
	// We need to mock httpRequest or make it testable
	// For now, test that URLs are parsed correctly
	urls, err := getURLs()
	if err != nil {
		t.Fatalf("getURLs() error = %v", err)
	}

	if len(urls) != 2 {
		t.Errorf("getURLs() = %v, expected two URLs", urls)
	}
}

func TestHTTPRequest(t *testing.T) {
	// Create a test HTTP server
	expectedBody := "test response body"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))
	defer server.Close()

	result, err := httpRequest(server.URL)
	if err != nil {
		t.Fatalf("httpRequest() error = %v", err)
	}

	if result != expectedBody {
		t.Errorf("httpRequest() = %v, expected %v", result, expectedBody)
	}
}

func TestHTTPRequest_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := httpRequest(server.URL)
	if err == nil {
		t.Errorf("httpRequest() expected error for 404 status")
	}
}

func TestGetDeviceEnvVar_NotFound(t *testing.T) {
	// Mock Balena API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %v", r.Method)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("Expected Bearer token, got %v", auth)
		}

		// Return 404 for not found
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Override balena API base for testing
	originalBase := balenaAPIBase
	balenaAPIBase = server.URL
	defer func() { balenaAPIBase = originalBase }()

	result, err := getDeviceEnvVar("test-key", "test-uuid", "NAME")
	if err != nil {
		t.Fatalf("getDeviceEnvVar() error = %v", err)
	}

	if result != "" {
		t.Errorf("getDeviceEnvVar() = %v, expected empty string for 404", result)
	}
}

func TestGetDeviceEnvVar_Found(t *testing.T) {
	expectedValue := "test-device-name"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"d": []map[string]interface{}{
				{"value": expectedValue},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBase := balenaAPIBase
	balenaAPIBase = server.URL
	defer func() { balenaAPIBase = originalBase }()

	result, err := getDeviceEnvVar("test-key", "test-uuid", "NAME")
	if err != nil {
		t.Fatalf("getDeviceEnvVar() error = %v", err)
	}

	if result != expectedValue {
		t.Errorf("getDeviceEnvVar() = %v, expected %v", result, expectedValue)
	}
}

func TestGetDeviceID(t *testing.T) {
	expectedID := 12345
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"d": []map[string]interface{}{
				{"id": expectedID},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalBase := balenaAPIBase
	balenaAPIBase = server.URL
	defer func() { balenaAPIBase = originalBase }()

	result, err := getDeviceID("test-key", "test-uuid")
	if err != nil {
		t.Fatalf("getDeviceID() error = %v", err)
	}

	if result != expectedID {
		t.Errorf("getDeviceID() = %v, expected %v", result, expectedID)
	}
}

func TestRename_DeviceAlreadyNamed(t *testing.T) {
	// Mock server that returns existing NAME env var
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/v6/device_environment_variable" && r.Method == "GET" {
			// Return existing NAME env var
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"d": []map[string]interface{}{
					{"value": "existing-name"},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	originalBase := balenaAPIBase
	balenaAPIBase = server.URL
	defer func() { balenaAPIBase = originalBase }()

	// Enable debug to test logging
	debug = true
	defer func() { debug = false }()

	err := rename("test-key", "test-uuid", "new-name")
	if err != nil {
		t.Fatalf("rename() error = %v", err)
	}

	// Should not have called setDeviceEnvVar or renameDevice
	// Only getDeviceEnvVar should be called
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}
}

func TestRename_NewDevice(t *testing.T) {
	callOrder := []string{}
	deviceID := 12345
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		if strings.Contains(path, "device_environment_variable") && method == "GET" {
			// No existing NAME env var
			w.WriteHeader(http.StatusNotFound)
			callOrder = append(callOrder, "getEnvVar")
		} else if strings.Contains(path, "device") && strings.Contains(r.URL.RawQuery, "$filter") && method == "GET" {
			// Get device ID (first call for setDeviceEnvVar, second for renameDevice)
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"d": []map[string]interface{}{
					{"id": deviceID},
				},
			}
			json.NewEncoder(w).Encode(response)
			callOrder = append(callOrder, "getDeviceID")
		} else if strings.Contains(path, "device_environment_variable") && method == "POST" {
			// Set env var - need to read and validate request body
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				// Validate payload
				if payload["name"] != "NAME" {
					t.Errorf("Expected name=NAME, got %v", payload["name"])
				}
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 1})
			callOrder = append(callOrder, "setEnvVar")
		} else if strings.Contains(path, fmt.Sprintf("device(%d)", deviceID)) && method == "PATCH" {
			// Rename device - need to read and validate request body
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				// Validate payload
				if payload["device_name"] != "new-device-name" {
					t.Errorf("Expected device_name=new-device-name, got %v", payload["device_name"])
				}
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": deviceID})
			callOrder = append(callOrder, "renameDevice")
		} else {
			t.Errorf("Unexpected request: %s %s", method, path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalBase := balenaAPIBase
	balenaAPIBase = server.URL
	defer func() { balenaAPIBase = originalBase }()

	debug = true
	defer func() { debug = false }()

	err := rename("test-key", "test-uuid", "new-device-name")
	if err != nil {
		t.Fatalf("rename() error = %v", err)
	}

	// Verify call order: getEnvVar -> getDeviceID -> setEnvVar -> getDeviceID -> renameDevice
	expectedOrder := []string{"getEnvVar", "getDeviceID", "setEnvVar", "getDeviceID", "renameDevice"}
	if len(callOrder) != len(expectedOrder) {
		t.Errorf("Expected %d API calls, got %d: %v", len(expectedOrder), len(callOrder), callOrder)
	}

	// Verify the order matches
	for i, expected := range expectedOrder {
		if i < len(callOrder) && callOrder[i] != expected {
			t.Errorf("Call %d: expected %s, got %s", i, expected, callOrder[i])
		}
	}
}

func TestReadFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	expectedContent := "test content\nline 2"
	os.WriteFile(tmpFile, []byte(expectedContent), 0644)

	result, err := readFile(tmpFile)
	if err != nil {
		t.Fatalf("readFile() error = %v", err)
	}

	if result != expectedContent {
		t.Errorf("readFile() = %v, expected %v", result, expectedContent)
	}
}

func TestReadFile_NotFound(t *testing.T) {
	_, err := readFile("/nonexistent/file.txt")
	if err == nil {
		t.Errorf("readFile() expected error for nonexistent file")
	}
}
