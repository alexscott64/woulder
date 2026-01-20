package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2/google"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env not found: %v", err)
	}

	// Setup OAuth2 client
	projectID := os.Getenv("GOOGLE_EARTH_ENGINE_PROJECT_ID")
	clientEmail := os.Getenv("GOOGLE_EARTH_ENGINE_CLIENT_EMAIL")
	privateKey := os.Getenv("GOOGLE_EARTH_ENGINE_PRIVATE_KEY")

	if projectID == "" || clientEmail == "" || privateKey == "" {
		log.Fatal("Missing Google Earth Engine credentials in .env")
	}

	serviceAccountJSON := fmt.Sprintf(`{
		"type": "service_account",
		"project_id": "%s",
		"private_key": "%s",
		"client_email": "%s",
		"token_uri": "https://oauth2.googleapis.com/token"
	}`, projectID, privateKey, clientEmail)

	config, err := google.JWTConfigFromJSON([]byte(serviceAccountJSON), "https://www.googleapis.com/auth/earthengine")
	if err != nil {
		log.Fatalf("Failed to parse credentials: %v", err)
	}

	client := config.Client(context.Background())

	// List algorithms
	url := fmt.Sprintf("https://earthengine.googleapis.com/v1/projects/%s/algorithms", projectID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Pretty print the response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(prettyJSON))
}
