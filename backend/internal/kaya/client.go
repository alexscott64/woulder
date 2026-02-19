package kaya

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	graphqlURL     = "https://kaya-beta.kayaclimb.com/graphql"
	rateLimitDelay = 1000 * time.Millisecond // 1 second between requests to be respectful
)

// Client handles communication with the Kaya GraphQL API
type Client struct {
	httpClient      *http.Client
	lastRequestTime time.Time
}

// NewClient creates a new Kaya API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
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

// GraphQLRequest represents a GraphQL request payload
type GraphQLRequest struct {
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
}

// GraphQLResponse represents a GraphQL response with data and errors
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string        `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// executeQuery performs a GraphQL query and returns the raw response
func (c *Client) executeQuery(req GraphQLRequest) (*GraphQLResponse, error) {
	c.rateLimit()

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", graphqlURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "*/*")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	httpReq.Header.Set("Origin", "https://kaya-app.kayaclimb.com")
	httpReq.Header.Set("Referer", "https://kaya-app.kayaclimb.com/")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	return &gqlResp, nil
}

// WebLocation represents a location (destination or area) from Kaya
type WebLocation struct {
	ID                                string          `json:"id"`
	Slug                              string          `json:"slug"`
	Name                              string          `json:"name"`
	Latitude                          *string         `json:"latitude"`
	Longitude                         *string         `json:"longitude"`
	PhotoURL                          *string         `json:"photo_url"`
	Description                       *string         `json:"description"`
	LocationType                      *LocationType   `json:"location_type"`
	ParentLocation                    *ParentLocation `json:"parent_location"`
	IsGBModeratedBouldering           bool            `json:"is_gb_moderated_bouldering"`
	IsGBModeratedRoutes               bool            `json:"is_gb_moderated_routes"`
	HasMapsDisabled                   bool            `json:"has_maps_disabled"`
	DescriptionBouldering             *string         `json:"description_bouldering"`
	DescriptionRoutes                 *string         `json:"description_routes"`
	DescriptionShortBouldering        *string         `json:"description_short_bouldering"`
	DescriptionShortRoutes            *string         `json:"description_short_routes"`
	AccessDescriptionBouldering       *string         `json:"access_description_bouldering"`
	AccessDescriptionRoutes           *string         `json:"access_description_routes"`
	AccessIssuesDescriptionBouldering *string         `json:"access_issues_description_bouldering"`
	AccessIssuesDescriptionRoutes     *string         `json:"access_issues_description_routes"`
	ClimbCount                        int             `json:"climb_count"`
	BoulderCount                      int             `json:"boulder_count"`
	RouteCount                        int             `json:"route_count"`
	AscentCount                       int             `json:"ascent_count"`
	IsAccessSensitive                 bool            `json:"is_access_sensitive"`
	IsClosed                          bool            `json:"is_closed"`
	ClosedDate                        *string         `json:"closed_date"`
	ClimbTypeID                       *string         `json:"climb_type_id"`
}

// LocationType represents the type of location
type LocationType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ParentLocation represents the parent location in the hierarchy
type ParentLocation struct {
	ID           string        `json:"id"`
	Slug         string        `json:"slug"`
	Name         string        `json:"name"`
	Latitude     *string       `json:"latitude"`
	Longitude    *string       `json:"longitude"`
	Description  *string       `json:"description"`
	LocationType *LocationType `json:"location_type"`
}

// WebClimb represents a climb (route or boulder) from Kaya
type WebClimb struct {
	Slug              string       `json:"slug"`
	Name              string       `json:"name"`
	Rating            *float64     `json:"rating"`
	AscentCount       int          `json:"ascent_count"`
	Grade             *Grade       `json:"grade"`
	ClimbType         *ClimbType   `json:"climb_type"`
	Color             *Color       `json:"color"`
	Gym               *Gym         `json:"gym"`
	Board             *Board       `json:"board"`
	Destination       *Destination `json:"destination"`
	Area              *Area        `json:"area"`
	IsGBModerated     bool         `json:"is_gb_moderated"`
	IsAccessSensitive bool         `json:"is_access_sensitive"`
	IsClosed          bool         `json:"is_closed"`
	IsOffensive       bool         `json:"is_offensive"`
}

// Grade represents a climbing grade
type Grade struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ClimbTypeID    *string  `json:"climb_type_id"`
	GradeTypeID    *string  `json:"grade_type_id"`
	Ordering       *int     `json:"ordering"`
	MappedGradeIDs []string `json:"mapped_grade_ids"`
	ClimbTypeGroup *string  `json:"climb_type_group"`
}

// ClimbType represents the type of climb
type ClimbType struct {
	Name string `json:"name"`
}

// Color represents the color of an indoor climb
type Color struct {
	Name string `json:"name"`
}

// Gym represents the gym where an indoor climb is located
type Gym struct {
	Name string `json:"name"`
}

// Board represents the training board (e.g., Kilter, Tension)
type Board struct {
	Name string `json:"name"`
}

// Destination represents the top-level location
type Destination struct {
	Name string `json:"name"`
}

// Area represents the sub-location or area
type Area struct {
	Name string `json:"name"`
}

// WebAscent represents a user's ascent (tick) of a climb
type WebAscent struct {
	ID        string    `json:"id"`
	User      *WebUser  `json:"user"`
	Climb     *WebClimb `json:"climb"`
	Date      string    `json:"date"`
	Comment   *string   `json:"comment"`
	Rating    *int      `json:"rating"`
	Stiffness *int      `json:"stiffness"`
	Grade     *Grade    `json:"grade"`
	Photo     *Media    `json:"photo"`
	Video     *Media    `json:"video"`
}

// WebUser represents a Kaya user
type WebUser struct {
	ID                   string   `json:"id"`
	Username             string   `json:"username"`
	Fname                *string  `json:"fname"`
	Lname                *string  `json:"lname"`
	PhotoURL             *string  `json:"photo_url"`
	IsPrivate            bool     `json:"is_private"`
	Bio                  *string  `json:"bio"`
	Height               *float64 `json:"height"`
	ApeIndex             *float64 `json:"ape_index"`
	LimitGradeBouldering *Grade   `json:"limit_grade_bouldering"`
	LimitGradeRoutes     *Grade   `json:"limit_grade_routes"`
	IsPremium            bool     `json:"is_premium"`
}

// Media represents photo or video media
type Media struct {
	PhotoURL *string `json:"photo_url"`
	ThumbURL *string `json:"thumb_url"`
	VideoURL *string `json:"video_url"`
}

// WebPost represents a user post with media items
type WebPost struct {
	ID          string        `json:"id"`
	Items       []WebPostItem `json:"items"`
	User        *WebUser      `json:"user"`
	DateCreated string        `json:"date_created"`
}

// WebPostItem represents a single media item within a post
type WebPostItem struct {
	ID                string     `json:"id"`
	PostID            string     `json:"post_id"`
	PhotoURL          *string    `json:"photo_url"`
	VideoURL          *string    `json:"video_url"`
	Climb             *WebClimb  `json:"climb"`
	Ascent            *WebAscent `json:"ascent"`
	VideoThumbnailURL *string    `json:"video_thumbnail_url"`
	Caption           *string    `json:"caption"`
}

// GetLocation fetches a location by slug
func (c *Client) GetLocation(slug string) (*WebLocation, error) {
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
      }
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
  }
}`

	req := GraphQLRequest{
		OperationName: "webLocation",
		Variables: map[string]interface{}{
			"slug": slug,
		},
		Query: query,
	}

	resp, err := c.executeQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get location %s: %w", slug, err)
	}

	var result struct {
		WebLocation *WebLocation `json:"webLocation"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse location response: %w", err)
	}

	return result.WebLocation, nil
}

// GetSubLocations fetches sub-locations for a given location
func (c *Client) GetSubLocations(locationID string, climbTypeID *string, offset, count int) ([]*WebLocation, error) {
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
      latitude
      longitude
      description
      location_type {
        id
        name
      }
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
  }
}`

	variables := map[string]interface{}{
		"location_id": locationID,
		"offset":      offset,
		"count":       count,
	}
	if climbTypeID != nil {
		variables["climb_type_id"] = *climbTypeID
	}

	req := GraphQLRequest{
		OperationName: "webLocationsForLocation",
		Variables:     variables,
		Query:         query,
	}

	resp, err := c.executeQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-locations for %s: %w", locationID, err)
	}

	var result struct {
		WebLocationsForLocation []*WebLocation `json:"webLocationsForLocation"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse sub-locations response: %w", err)
	}

	return result.WebLocationsForLocation, nil
}

// GetClimbs fetches climbs for a given location
func (c *Client) GetClimbs(locationID string, climbTypeID *string, offset, count int) ([]*WebClimb, error) {
	query := `query webClimbsForLocation($location_id: ID!, $climb_name: String, $climb_type_id: ID, $min_grade_id: ID, $max_grade_id: ID, $min_rating: Int, $max_rating: Int, $sort_by: sortClimbs, $offset: Int!, $count: Int!) {
  webClimbsForLocation(location_id: $location_id, climb_name: $climb_name, climb_type_id: $climb_type_id, min_grade_id: $min_grade_id, max_grade_id: $max_grade_id, min_rating: $min_rating, max_rating: $max_rating, sort_by: $sort_by, offset: $offset, count: $count, use_reduced_query: true) {
    slug
    name
    rating
    ascent_count
    grade {
      name
      id
      ordering
      climb_type_id
    }
    climb_type {
      name
    }
    color {
      name
    }
    gym {
      name
    }
    board {
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

	variables := map[string]interface{}{
		"location_id": locationID,
		"climb_name":  "",
		"offset":      offset,
		"count":       count,
		"min_rating":  nil,
		"max_rating":  nil,
		"sort_by":     nil,
	}
	if climbTypeID != nil {
		variables["climb_type_id"] = *climbTypeID
	}

	req := GraphQLRequest{
		OperationName: "webClimbsForLocation",
		Variables:     variables,
		Query:         query,
	}

	resp, err := c.executeQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get climbs for %s: %w", locationID, err)
	}

	var result struct {
		WebClimbsForLocation []*WebClimb `json:"webClimbsForLocation"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse climbs response: %w", err)
	}

	return result.WebClimbsForLocation, nil
}

// GetAscents fetches ascents (ticks) for a given location
func (c *Client) GetAscents(locationID string, offset, count int) ([]*WebAscent, error) {
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
      bio
      height
      ape_index
      limit_grade_bouldering {
        name
        id
      }
      limit_grade_routes {
        name
        id
      }
      is_premium
    }
    climb {
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
      color {
        name
      }
      gym {
        name
      }
      board {
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
    date
    comment
    rating
    stiffness
    grade {
      id
      name
      climb_type_id
      grade_type_id
      ordering
      mapped_grade_ids
      climb_type_group
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

	req := GraphQLRequest{
		OperationName: "webAscentsForLocation",
		Variables: map[string]interface{}{
			"location_id": locationID,
			"offset":      offset,
			"count":       count,
		},
		Query: query,
	}

	resp, err := c.executeQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get ascents for %s: %w", locationID, err)
	}

	var result struct {
		WebAscentsForLocation []*WebAscent `json:"webAscentsForLocation"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ascents response: %w", err)
	}

	return result.WebAscentsForLocation, nil
}

// GetPosts fetches posts for a given location
func (c *Client) GetPosts(locationID string, subLocationIDs []string, offset, count int) ([]*WebPost, error) {
	query := `query webPostsForLocation($location_id: ID!, $sub_location_ids: [ID], $count: Int!, $offset: Int!) {
  webPostsForLocation(location_id: $location_id, sub_location_ids: $sub_location_ids, count: $count, offset: $offset) {
    id
    items {
      id
      post_id
      photo_url
      video_url
      climb {
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
      ascent {
        id
        date
        comment
        rating
      }
      video_thumbnail_url
      caption
    }
    user {
      id
      username
      fname
      lname
      photo_url
      is_private
      is_premium
    }
    date_created
  }
}`

	variables := map[string]interface{}{
		"location_id": locationID,
		"offset":      offset,
		"count":       count,
	}
	if len(subLocationIDs) > 0 {
		variables["sub_location_ids"] = subLocationIDs
	}

	req := GraphQLRequest{
		OperationName: "webPostsForLocation",
		Variables:     variables,
		Query:         query,
	}

	resp, err := c.executeQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts for %s: %w", locationID, err)
	}

	var result struct {
		WebPostsForLocation []*WebPost `json:"webPostsForLocation"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse posts response: %w", err)
	}

	return result.WebPostsForLocation, nil
}
