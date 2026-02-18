package kaya

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// BrowserClient uses headless Chrome to bypass Cloudflare protection
type BrowserClient struct {
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	graphqlURL    string
	authToken     string
	rateLimitWait time.Duration
	lastRequest   time.Time
}

// NewBrowserClient creates a new browser-based GraphQL client that can bypass Cloudflare
// authToken: JWT token for Kaya API authentication (optional, can be empty for public data)
func NewBrowserClient(authToken string) (*BrowserClient, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true), // Disable CORS for cross-origin requests
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	return &BrowserClient{
		allocCtx:      allocCtx,
		allocCancel:   allocCancel,
		graphqlURL:    "https://kaya-beta.kayaclimb.com/graphql",
		authToken:     authToken,
		rateLimitWait: 2 * time.Second, // 2 seconds between requests
	}, nil
}

// Close cleans up browser resources
func (c *BrowserClient) Close() {
	if c.allocCancel != nil {
		c.allocCancel()
	}
}

// rateLimit ensures we respect rate limits
func (c *BrowserClient) rateLimit() {
	if !c.lastRequest.IsZero() {
		elapsed := time.Since(c.lastRequest)
		if elapsed < c.rateLimitWait {
			time.Sleep(c.rateLimitWait - elapsed)
		}
	}
	c.lastRequest = time.Now()
}

// executeQuery executes a GraphQL query using headless browser to bypass Cloudflare
func (c *BrowserClient) executeQuery(query string, variables map[string]interface{}, result interface{}) error {
	c.rateLimit()

	// Create a new context for this request
	ctx, cancel := chromedp.NewContext(c.allocCtx)
	defer cancel()

	// Set timeout for this request
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Extract operation name from query (first word after "query" or "mutation")
	operationName := ""
	if strings.Contains(query, "query ") {
		parts := strings.Split(strings.TrimSpace(strings.Split(query, "query ")[1]), "(")
		if len(parts) > 0 {
			operationName = strings.TrimSpace(parts[0])
		}
	}

	// Prepare the GraphQL request body with operation name
	requestBody := map[string]interface{}{
		"operationName": operationName,
		"query":         query,
		"variables":     variables,
	}
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: Log the request being sent
	log.Printf("[DEBUG] Sending GraphQL request: %s", string(requestJSON))

	var responseBody string

	// Simple approach: Navigate, wait for Cloudflare, make synchronous XHR request
	// We need to properly escape the JSON for JavaScript
	escapedJSON := strings.ReplaceAll(string(requestJSON), `\`, `\\`)
	escapedJSON = strings.ReplaceAll(escapedJSON, `"`, `\"`)
	escapedJSON = strings.ReplaceAll(escapedJSON, "\n", `\n`)
	escapedJSON = strings.ReplaceAll(escapedJSON, "\r", `\r`)
	escapedJSON = strings.ReplaceAll(escapedJSON, "\t", `\t`)

	// Navigate to the site and make the GraphQL request
	err = chromedp.Run(ctx,
		chromedp.Navigate("https://kaya-app.kayaclimb.com"),
		chromedp.Sleep(3*time.Second), // Wait for Cloudflare and page load
	)

	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}

	// Log auth token status
	if c.authToken == "" {
		log.Printf("[WARNING] No auth token provided - API calls may fail for protected data")
	} else {
		log.Printf("[DEBUG] Using provided auth token: %s...", c.authToken[:min(len(c.authToken), 50)])
	}

	// Now make the GraphQL request with the authorization header
	err = chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				var xhr = new XMLHttpRequest();
				xhr.open('POST', '%s', false); // Synchronous request
				xhr.setRequestHeader('Content-Type', 'application/json');
				xhr.setRequestHeader('Accept', '*/*');
				xhr.setRequestHeader('Origin', 'https://kaya-app.kayaclimb.com');
				xhr.setRequestHeader('Referer', 'https://kaya-app.kayaclimb.com/location/Leavenworth-344933');
				%s
				xhr.withCredentials = true; // Include cookies
				xhr.send("%s");
				return xhr.responseText;
			})();
		`, c.graphqlURL,
			func() string {
				if c.authToken != "" {
					return fmt.Sprintf(`xhr.setRequestHeader('authorization', 'Bearer %s');`, c.authToken)
				}
				return ""
			}(),
			escapedJSON), &responseBody),
	)

	if err != nil {
		return fmt.Errorf("browser request failed: %w", err)
	}

	// Debug: Log the response
	if len(responseBody) < 500 {
		log.Printf("[DEBUG] Received response: %s", responseBody)
	} else {
		log.Printf("[DEBUG] Received response (truncated): %s...", responseBody[:500])
	}

	// Parse GraphQL response
	var graphqlResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message    string                 `json:"message"`
			Extensions map[string]interface{} `json:"extensions"`
		} `json:"errors"`
	}

	if err := json.Unmarshal([]byte(responseBody), &graphqlResp); err != nil {
		log.Printf("Failed to parse GraphQL response: %s", responseBody)
		return fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		var errorMessages []string
		for _, err := range graphqlResp.Errors {
			msg := err.Message
			if err.Extensions != nil {
				extJSON, _ := json.Marshal(err.Extensions)
				msg += fmt.Sprintf(" (extensions: %s)", string(extJSON))
			}
			errorMessages = append(errorMessages, msg)
		}
		log.Printf("[DEBUG] GraphQL errors: %s", strings.Join(errorMessages, ", "))
		return fmt.Errorf("GraphQL errors: %s", strings.Join(errorMessages, ", "))
	}

	// Unmarshal the data into the result interface
	if err := json.Unmarshal(graphqlResp.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// GetLocation fetches location details by slug using headless browser
func (c *BrowserClient) GetLocation(slug string) (*WebLocation, error) {
	query := `query webLocation($slug: String!) {
  webLocation(slug: $slug) {
    id
    slug
    name
    latitude
    longitude
    photo_url
    description
    location_type {
      id
      name
      __typename
    }
    parent_location {
      id
      slug
      name
      latitude
      longitude
      description
      location_type {
        id
        name
        __typename
      }
      __typename
    }
    is_gb_moderated_bouldering
    is_gb_moderated_routes
    has_maps_disabled
    description_bouldering
    description_routes
    description_short_bouldering
    description_short_routes
    access_description_bouldering
    access_description_routes
    access_issues_description_bouldering
    access_issues_description_routes
    climb_count
    boulder_count
    route_count
    ascent_count
    is_access_sensitive
    is_closed
    closed_date
    climb_type_id
    __typename
  }
}`

	variables := map[string]interface{}{
		"slug": slug,
	}

	var response struct {
		WebLocation *WebLocation `json:"webLocation"`
	}

	if err := c.executeQuery(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to get location %s: %w", slug, err)
	}

	if response.WebLocation == nil {
		return nil, fmt.Errorf("location %s not found", slug)
	}

	return response.WebLocation, nil
}

// GetSubLocations fetches child locations for a parent location using headless browser
func (c *BrowserClient) GetSubLocations(locationID string, climbTypeID *string, offset, count int) ([]*WebLocation, error) {
	query := `query webLocationsForLocation($location_id: ID!, $offset: Int!, $count: Int!, $climb_type_id: ID) {
  webLocationsForLocation(location_id: $location_id, offset: $offset, count: $count, climb_type_id: $climb_type_id) {
    id
    slug
    name
    latitude
    longitude
    photo_url
    description
    location_type {
      id
      name
    }
    parent_location {
      id
      slug
      name
    }
    climb_count
    boulder_count
    route_count
    ascent_count
    is_access_sensitive
    is_closed
  }
}`

	var climbType interface{}
	if climbTypeID != nil {
		climbType = *climbTypeID
	}

	variables := map[string]interface{}{
		"location_id":   locationID,
		"offset":        offset,
		"count":         count,
		"climb_type_id": climbType,
	}

	var response struct {
		WebLocationsForLocation []*WebLocation `json:"webLocationsForLocation"`
	}

	if err := c.executeQuery(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to get sub-locations for %s: %w", locationID, err)
	}

	return response.WebLocationsForLocation, nil
}

// GetClimbs fetches climbs for a location using headless browser
func (c *BrowserClient) GetClimbs(locationID string, climbTypeID *string, offset, count int) ([]*WebClimb, error) {
	query := `query webClimbsForLocation($location_id: ID!, $climb_name: String, $climb_type_id: ID, $min_grade_id: ID, $max_grade_id: ID, $min_rating: Int, $max_rating: Int, $sort_by: sortClimbs, $offset: Int!, $count: Int!) {
  webClimbsForLocation(location_id: $location_id, climb_name: $climb_name, climb_type_id: $climb_type_id, min_grade_id: $min_grade_id, max_grade_id: $max_grade_id, min_rating: $min_rating, max_rating: $max_rating, sort_by: $sort_by, offset: $offset, count: $count, use_reduced_query: true) {
    slug
    name
    rating
    ascent_count
    grade {
      name
      id
    }
    climb_type {
      name
    }
    destination {
      name
    }
    area {
      name
    }
    is_gb_moderated
    is_access_sensitive
    is_closed
    is_offensive
  }
}`

	var climbType interface{}
	if climbTypeID != nil {
		climbType = *climbTypeID
	}

	variables := map[string]interface{}{
		"location_id":   locationID,
		"climb_name":    "",
		"climb_type_id": climbType,
		"min_rating":    nil,
		"max_rating":    nil,
		"sort_by":       nil,
		"offset":        offset,
		"count":         count,
	}

	var response struct {
		WebClimbsForLocation []*WebClimb `json:"webClimbsForLocation"`
	}

	if err := c.executeQuery(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to get climbs for location %s: %w", locationID, err)
	}

	return response.WebClimbsForLocation, nil
}

// GetAscents fetches ascents for a location using headless browser
func (c *BrowserClient) GetAscents(locationID string, offset, count int) ([]*WebAscent, error) {
	query := `query webAscentsForLocation($location_id: ID!, $count: Int!, $offset: Int!) {
  webAscentsForLocation(location_id: $location_id, offset: $offset, count: $count) {
    id
    user {
      id
      username
      fname
      lname
      photo_url
      is_private
      limit_grade_bouldering {
        name
        id
      }
     limit_grade_routes {
        name
        id
      }
    }
    climb {
      slug
      name
      rating
      grade {
        name
        id
      }
      climb_type {
        name
      }
    }
    date
    comment
    rating
    stiffness
    grade {
      name
      id
    }
    photo {
      photo_url
      thumb_url
    }
    video {
      video_url
      thumb_url
    }
  }
}`

	variables := map[string]interface{}{
		"location_id": locationID,
		"offset":      offset,
		"count":       count,
	}

	var response struct {
		WebAscentsForLocation []*WebAscent `json:"webAscentsForLocation"`
	}

	if err := c.executeQuery(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to get ascents for location %s: %w", locationID, err)
	}

	return response.WebAscentsForLocation, nil
}

// GetPosts fetches posts for a location using headless browser
func (c *BrowserClient) GetPosts(locationID string, subLocationIDs []string, offset, count int) ([]*WebPost, error) {
	query := `query webPostsForLocation($location_id: ID!, $sub_location_ids: [ID], $count: Int!, $offset: Int!) {
  webPostsForLocation(location_id: $location_id, sub_location_ids: $sub_location_ids, count: $count, offset: $offset) {
    id
    items {
      id
      post_id
      photo_url
      video_url
      video_thumbnail_url
      caption
      climb {
        slug
        name
        grade {
          name
        }
      }
    }
    user {
      id
      username
      fname
      lname
      photo_url
    }
    date_created
  }
}`

	variables := map[string]interface{}{
		"location_id":      locationID,
		"sub_location_ids": subLocationIDs,
		"offset":           offset,
		"count":            count,
	}

	var response struct {
		WebPostsForLocation []*WebPost `json:"webPostsForLocation"`
	}

	if err := c.executeQuery(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to get posts for location %s: %w", locationID, err)
	}

	return response.WebPostsForLocation, nil
}
