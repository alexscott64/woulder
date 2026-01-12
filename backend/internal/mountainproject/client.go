package mountainproject

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL        = "https://www.mountainproject.com/api/v2"
	rateLimitDelay = 500 * time.Millisecond // 500ms between requests to be respectful
)

// Client handles communication with the Mountain Project API
type Client struct {
	httpClient      *http.Client
	lastRequestTime time.Time
}

// NewClient creates a new Mountain Project API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// rateLimit ensures we don't exceed rate limits by waiting if needed
func (c *Client) rateLimit() {
	elapsed := time.Since(c.lastRequestTime)
	if elapsed < rateLimitDelay {
		time.Sleep(rateLimitDelay - elapsed)
	}
	c.lastRequestTime = time.Now()
}

// AreaResponse represents the response from the Mountain Project area endpoint
type AreaResponse struct {
	ID       int            `json:"id"`
	Title    string         `json:"title"`
	Type     string         `json:"type"`     // "Area" typically
	Children []ChildElement `json:"children"` // Can be subareas or routes
}

// ChildElement represents either a subarea or a route within an area
type ChildElement struct {
	ID         int      `json:"id"`
	Title      string   `json:"title"`
	Type       string   `json:"type"`        // "Area" for subareas, "Route" for routes
	RouteTypes []string `json:"route_types"` // e.g., ["Boulder", "Trad"], only present for routes
}

// TickResponse represents the response from the Mountain Project ticks endpoint
type TickResponse struct {
	Data []Tick `json:"data"`
}

// Tick represents a single climb log entry
type Tick struct {
	Date    string          `json:"date"`    // "Jan 2, 2006, 3:04 pm"
	Style   string          `json:"style"`   // "Lead", "Flash", "Send", etc.
	Comment *string         `json:"comment"` // User's comment (can be null)
	Text    json.RawMessage `json:"text"`    // Full tick text (can be string, false, or empty)
	User    json.RawMessage `json:"user"`    // User who logged the tick (can be object or false)
}

// TickUser represents the user who logged a tick
type TickUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"` // Mountain Project username
}

// GetTextString extracts the text field as a string, handling cases where it's false or empty
func (t *Tick) GetTextString() string {
	// Try to unmarshal as string
	var text string
	if err := json.Unmarshal(t.Text, &text); err == nil {
		return text
	}
	// If it's not a string (e.g., false), return empty string
	return ""
}

// GetUserName extracts the username from the user field, handling cases where it's false
func (t *Tick) GetUserName() string {
	// Try to unmarshal as TickUser object
	var user TickUser
	if err := json.Unmarshal(t.User, &user); err == nil {
		return user.Name
	}
	// If it's not an object (e.g., false for anonymous/deleted users), return empty string
	return ""
}

// GetArea fetches area data including children (subareas and routes)
func (c *Client) GetArea(areaID string) (*AreaResponse, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/areas/%s", baseURL, areaID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Woulder/1.0 (https://woulder.com)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch area %s: %w", areaID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d for area %s: %s", resp.StatusCode, areaID, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var areaResp AreaResponse
	if err := json.Unmarshal(body, &areaResp); err != nil {
		return nil, fmt.Errorf("failed to parse area response for %s: %w", areaID, err)
	}

	return &areaResp, nil
}

// GetRouteTicks fetches tick data (climb logs) for a specific route
func (c *Client) GetRouteTicks(routeID string) ([]Tick, error) {
	c.rateLimit()

	url := fmt.Sprintf("%s/routes/%s/ticks", baseURL, routeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Woulder/1.0 (https://woulder.com)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ticks for route %s: %w", routeID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d for route %s: %s", resp.StatusCode, routeID, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tickResp TickResponse
	if err := json.Unmarshal(body, &tickResp); err != nil {
		return nil, fmt.Errorf("failed to parse tick response for route %s: %w", routeID, err)
	}

	return tickResp.Data, nil
}
