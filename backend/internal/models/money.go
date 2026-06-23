package models

import (
	"encoding/json"
	"time"
)

const (
	MoneyFeatureTrail   = "trail"
	MoneyFeatureTopo    = "topo"
	MoneyFeaturePOI     = "poi"
	MoneyFeatureDrawing = "drawing"

	MoneyStatusDraft    = "draft"
	MoneyStatusActive   = "active"
	MoneyStatusArchived = "archived"

	MoneyNotePrivate = "private"
	MoneyNoteTeam    = "team"
)

type MoneyProject struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	CenterLat   float64   `json:"center_lat"`
	CenterLon   float64   `json:"center_lon"`
	DefaultZoom float64   `json:"default_zoom"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MoneyFeature struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	FeatureType string          `json:"feature_type"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Status      string          `json:"status"`
	GeoJSON     json.RawMessage `json:"geojson"`
	Style       json.RawMessage `json:"style"`
	Properties  json.RawMessage `json:"properties"`
	MinLat      *float64        `json:"min_lat,omitempty"`
	MinLon      *float64        `json:"min_lon,omitempty"`
	MaxLat      *float64        `json:"max_lat,omitempty"`
	MaxLon      *float64        `json:"max_lon,omitempty"`
	CreatedBy   string          `json:"created_by"`
	UpdatedBy   string          `json:"updated_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type MoneyNote struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"project_id"`
	FeatureID  *string   `json:"feature_id,omitempty"`
	Body       string    `json:"body"`
	Visibility string    `json:"visibility"`
	CreatedBy  string    `json:"created_by"`
	UpdatedBy  string    `json:"updated_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MoneyUpload struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	FeatureID        *string   `json:"feature_id,omitempty"`
	NoteID           *string   `json:"note_id,omitempty"`
	OriginalFilename string    `json:"original_filename"`
	StorageKey       string    `json:"-"`
	ContentType      string    `json:"content_type"`
	ByteSize         int64     `json:"byte_size"`
	Width            *int      `json:"width,omitempty"`
	Height           *int      `json:"height,omitempty"`
	ChecksumSHA256   string    `json:"checksum_sha256"`
	UploadedBy       string    `json:"uploaded_by"`
	CreatedAt        time.Time `json:"created_at"`
}

type MoneyPermissions struct {
	CanRead  bool `json:"can_read"`
	CanWrite bool `json:"can_write"`
	IsAdmin  bool `json:"is_admin"`
}

type MoneyProjectResponse struct {
	Project     MoneyProject     `json:"project"`
	User        CurrentUser      `json:"user"`
	Permissions MoneyPermissions `json:"permissions"`
}

type MoneySnapshot struct {
	Project        MoneyProject           `json:"project"`
	Features       []MoneyFeature         `json:"features"`
	NoteCounts     map[string]int         `json:"note_counts"`
	PrimaryUploads map[string]MoneyUpload `json:"primary_uploads"`
}

type MoneyFeatureDetail struct {
	Feature MoneyFeature  `json:"feature"`
	Notes   []MoneyNote   `json:"notes"`
	Uploads []MoneyUpload `json:"uploads"`
}

type MoneyFeatureRequest struct {
	FeatureType string          `json:"feature_type"`
	Title       string          `json:"title"`
	Description *string         `json:"description"`
	Status      string          `json:"status"`
	GeoJSON     json.RawMessage `json:"geojson"`
	Style       json.RawMessage `json:"style"`
	Properties  json.RawMessage `json:"properties"`
}

type MoneyNoteRequest struct {
	Body       string `json:"body"`
	Visibility string `json:"visibility"`
}

type BBox struct {
	MinLon float64
	MinLat float64
	MaxLon float64
	MaxLat float64
}

type MoneyFeatureFilter struct {
	FeatureType  string
	Status       string
	BBox         *BBox
	UpdatedAfter *time.Time
}
