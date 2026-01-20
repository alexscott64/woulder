package main

import (
	"bytes"
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

	log.Println("=== Google Earth Engine API Test ===")
	log.Println()

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
	log.Printf("✓ OAuth2 client initialized (project: %s)\n", projectID)
	log.Println()

	// Test coordinates
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Leavenworth, WA", 47.6, -120.9},
		{"Bishop, CA", 37.35, -118.7},
	}

	for _, tc := range testCases {
		log.Printf("Testing: %s (%.6f, %.6f)", tc.name, tc.lat, tc.lon)
		coverage, err := getTreeCoverage(client, projectID, tc.lat, tc.lon)
		if err != nil {
			log.Printf("  ✗ ERROR: %v\n", err)
		} else {
			log.Printf("  ✓ Tree coverage: %.1f%%\n", coverage)
		}
		log.Println()
	}

	log.Println("=== Test Complete ===")
}

func getTreeCoverage(client *http.Client, projectID string, lat, lon float64) (float64, error) {
	ctx := context.Background()

	// Use USGS NLCD (National Land Cover Database) for tree canopy
	// Dataset: USGS/NLCD_RELEASES/2021_REL/NLCD
	// Band: tree_canopy

	// Create expression using Cloud API format with values and result
	// Algorithm names must be fully qualified: algorithms/Image.load, algorithms/GeometryConstructors.Point, etc.
	expression := map[string]interface{}{
		"expression": map[string]interface{}{
			"result": "0",
			"values": map[string]interface{}{
				// Node 0: Final reduceRegion call
				"0": map[string]interface{}{
					"functionInvocationValue": map[string]interface{}{
						"functionName": "algorithms/Image.reduceRegion",
						"arguments": map[string]interface{}{
							"image": map[string]interface{}{
								"valueReference": "1", // Reference to node 1 (Image.select)
							},
							"geometry": map[string]interface{}{
								"valueReference": "2", // Reference to node 2 (Geometry.Point)
							},
							"reducer": map[string]interface{}{
								"valueReference": "3", // Reference to node 3 (Reducer.first)
							},
							"scale": map[string]interface{}{
								"constantValue": 30,
							},
						},
					},
				},
				// Node 1: Image.select
				"1": map[string]interface{}{
					"functionInvocationValue": map[string]interface{}{
						"functionName": "algorithms/Image.select",
						"arguments": map[string]interface{}{
							"input": map[string]interface{}{
								"valueReference": "4", // Reference to node 4 (Image load)
							},
							"bandSelectors": map[string]interface{}{
								"constantValue": []string{"tree_canopy"},
							},
						},
					},
				},
				// Node 2: Geometry.Point
				"2": map[string]interface{}{
					"functionInvocationValue": map[string]interface{}{
						"functionName": "algorithms/GeometryConstructors.Point",
						"arguments": map[string]interface{}{
							"coordinates": map[string]interface{}{
								"constantValue": []float64{lon, lat},
							},
						},
					},
				},
				// Node 3: Reducer.first
				"3": map[string]interface{}{
					"functionInvocationValue": map[string]interface{}{
						"functionName": "algorithms/Reducer.first",
						"arguments": map[string]interface{}{},
					},
				},
				// Node 4: Image load
				"4": map[string]interface{}{
					"functionInvocationValue": map[string]interface{}{
						"functionName": "algorithms/Image.load",
						"arguments": map[string]interface{}{
							"id": map[string]interface{}{
								"constantValue": "USGS/NLCD_RELEASES/2021_REL/NLCD",
							},
						},
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(expression)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use computeValue endpoint
	url := fmt.Sprintf("https://earthengine.googleapis.com/v1/projects/%s/value:compute", projectID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract tree_canopy value
	if resultValue, ok := result["result"].(map[string]interface{}); ok {
		if treeCanopy, ok := resultValue["tree_canopy"].(float64); ok {
			return treeCanopy, nil
		}
	}

	return 0, fmt.Errorf("unexpected response format: %v", result)
}
