package models

import (
	"encoding/json"
	"time"
)

const (
	MoneyFeatureArea    = "area"
	MoneyFeatureBoulder = "boulder"
	MoneyFeatureProblem = "problem"
	MoneyFeatureTrail   = "trail"
	MoneyFeatureTopo    = "topo"
	MoneyFeaturePOI     = "poi"
	MoneyFeatureDrawing = "drawing"

	MoneyStatusDraft       = "draft"
	MoneyStatusActive      = "active"
	MoneyStatusArchived    = "archived"
	MoneyStatusScouted     = "scouted"
	MoneyStatusNeedsWork   = "needs-work"
	MoneyStatusCleaning    = "cleaning"
	MoneyStatusEstablished = "established"
	MoneyStatusProject     = "project"
	MoneyStatusSent        = "sent"

	MoneyNotePrivate = "private"
	MoneyNoteTeam    = "team"

	MoneyTrailCategoryConnector          = "connector"
	MoneyTrailCategoryApproach           = "approach"
	MoneyTrailCategoryTrailToArea        = "trail_to_area"
	MoneyTrailCategoryTrailToDestination = "trail_to_destination"
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
	ID              string          `json:"id"`
	ProjectID       string          `json:"project_id"`
	ParentFeatureID *string         `json:"parent_feature_id,omitempty"`
	FeatureType     string          `json:"feature_type"`
	Title           string          `json:"title"`
	Description     *string         `json:"description,omitempty"`
	Status          string          `json:"status"`
	GeoJSON         json.RawMessage `json:"geojson"`
	Style           json.RawMessage `json:"style"`
	Properties      json.RawMessage `json:"properties"`
	MinLat          *float64        `json:"min_lat,omitempty"`
	MinLon          *float64        `json:"min_lon,omitempty"`
	MaxLat          *float64        `json:"max_lat,omitempty"`
	MaxLon          *float64        `json:"max_lon,omitempty"`
	SortOrder       int             `json:"sort_order"`
	ExternalRef     *string         `json:"external_ref,omitempty"`
	ImportSource    *string         `json:"import_source,omitempty"`
	CreatedBy       string          `json:"created_by"`
	UpdatedBy       string          `json:"updated_by"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type MoneyNoteBlock struct {
	Kind     string          `json:"kind"`
	UploadID *string         `json:"upload_id,omitempty"`
	URL      *string         `json:"url,omitempty"`
	Name     *string         `json:"name,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type MoneyNote struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"project_id"`
	FeatureID    *string         `json:"feature_id,omitempty"`
	TargetType   string          `json:"target_type"`
	TargetRef    *string         `json:"target_ref,omitempty"`
	Body         string          `json:"body"`
	Visibility   string          `json:"visibility"`
	Tags         []string        `json:"tags"`
	Blocks       json.RawMessage `json:"blocks"`
	ExternalRef  *string         `json:"external_ref,omitempty"`
	ImportSource *string         `json:"import_source,omitempty"`
	CreatedBy    string          `json:"created_by"`
	UpdatedBy    string          `json:"updated_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type MoneyUpload struct {
	ID                  string          `json:"id"`
	ProjectID           string          `json:"project_id"`
	FeatureID           *string         `json:"feature_id,omitempty"`
	NoteID              *string         `json:"note_id,omitempty"`
	OriginalFilename    string          `json:"original_filename"`
	Title               *string         `json:"title,omitempty"`
	Comments            *string         `json:"comments,omitempty"`
	StorageKey          string          `json:"-"`
	ContentType         string          `json:"content_type"`
	ByteSize            int64           `json:"byte_size"`
	Width               *int            `json:"width,omitempty"`
	Height              *int            `json:"height,omitempty"`
	ChecksumSHA256      string          `json:"checksum_sha256"`
	BlockKind           string          `json:"block_kind"`
	Metadata            json.RawMessage `json:"metadata"`
	AssetKind           string          `json:"asset_kind"`
	StorageBackend      string          `json:"storage_backend"`
	StorageBucket       *string         `json:"storage_bucket,omitempty"`
	StorageRegion       *string         `json:"storage_region,omitempty"`
	StorageETag         *string         `json:"storage_etag,omitempty"`
	StorageVersionID    *string         `json:"storage_version_id,omitempty"`
	Visibility          string          `json:"visibility"`
	SyncStatus          string          `json:"sync_status"`
	DeletedAt           *time.Time      `json:"deleted_at,omitempty"`
	DeletedBy           *string         `json:"deleted_by,omitempty"`
	DeleteRequestedAt   *time.Time      `json:"delete_requested_at,omitempty"`
	PhysicallyDeletedAt *time.Time      `json:"physically_deleted_at,omitempty"`
	UploadedBy          string          `json:"uploaded_by"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type MoneyUploadDownloadURL struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	ProxyURL  string    `json:"proxy_url"`
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

type MoneyCragSnapshot struct {
	Project MoneyProject    `json:"project"`
	Root    *MoneyCragNode  `json:"root"`
	Trails  []MoneyCragNode `json:"trails"`
	Notes   []MoneyNote     `json:"notes"`
	Uploads []MoneyUpload   `json:"uploads"`
}

type MoneyCragNode struct {
	Feature  MoneyFeature    `json:"feature"`
	Children []MoneyCragNode `json:"children"`
	Boulders []MoneyCragNode `json:"boulders"`
	Problems []MoneyCragNode `json:"problems"`
}

type MoneyFeatureDetail struct {
	Feature MoneyFeature  `json:"feature"`
	Notes   []MoneyNote   `json:"notes"`
	Uploads []MoneyUpload `json:"uploads"`
}

type MoneyTrashItem struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	FeatureType     string    `json:"feature_type"`
	ParentFeatureID *string   `json:"parent_feature_id,omitempty"`
	Path            []string  `json:"path"`
	DeletedAt       time.Time `json:"deleted_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DescendantCount int       `json:"descendant_count"`
}

type MoneyTrashResponse struct {
	Items []MoneyTrashItem `json:"items"`
}

type MoneyFeatureRequest struct {
	ParentFeatureID *string         `json:"parent_feature_id,omitempty"`
	FeatureType     string          `json:"feature_type"`
	Title           string          `json:"title"`
	Description     *string         `json:"description"`
	Status          string          `json:"status"`
	GeoJSON         json.RawMessage `json:"geojson"`
	Style           json.RawMessage `json:"style"`
	Properties      json.RawMessage `json:"properties"`
	SortOrder       int             `json:"sort_order"`
	ExternalRef     *string         `json:"external_ref,omitempty"`
	ImportSource    *string         `json:"import_source,omitempty"`
}

type MoneyCragAreaRequest struct {
	ParentFeatureID *string         `json:"parent_feature_id,omitempty"`
	Title           string          `json:"title"`
	Description     *string         `json:"description"`
	GeoJSON         json.RawMessage `json:"geojson"`
	Properties      json.RawMessage `json:"properties"`
}

type MoneyAreaGeometryRequest struct {
	GeoJSON json.RawMessage `json:"geojson"`
}

type MoneyArchiveMode string

const (
	MoneyArchiveModeSubtree         MoneyArchiveMode = "subtree"
	MoneyArchiveModePromoteChildren MoneyArchiveMode = "promote_children"
)

type MoneyArchiveFeatureRequest struct {
	Mode MoneyArchiveMode `json:"mode"`
}

type MoneyMoveFeatureRequest struct {
	ParentFeatureID *string `json:"parent_feature_id"`
	SortOrder       *int    `json:"sort_order,omitempty"`
}

type MoneyCragBoulderRequest struct {
	ParentFeatureID string          `json:"parent_feature_id"`
	Title           string          `json:"title"`
	Description     *string         `json:"description"`
	DevStatus       string          `json:"dev_status"`
	GeoJSON         json.RawMessage `json:"geojson"`
	Properties      json.RawMessage `json:"properties"`
}

type MoneyCragProblemRequest struct {
	BoulderID   string          `json:"boulder_id"`
	Name        string          `json:"name"`
	Grade       string          `json:"grade"`
	Status      string          `json:"status"`
	Stars       int             `json:"stars"`
	FA          *string         `json:"fa,omitempty"`
	Types       []string        `json:"types"`
	Description *string         `json:"description"`
	Properties  json.RawMessage `json:"properties"`
}

type MoneyBoulderStatusRequest struct {
	DevStatus string `json:"dev_status"`
}

type MoneyNoteRequest struct {
	Body       string          `json:"body"`
	Visibility string          `json:"visibility"`
	TargetType string          `json:"target_type"`
	TargetRef  *string         `json:"target_ref,omitempty"`
	Tags       []string        `json:"tags"`
	Blocks     json.RawMessage `json:"blocks"`
}

type MoneyUploadMetadataRequest struct {
	Title    *string `json:"title"`
	Comments *string `json:"comments"`
}

type BBox struct {
	MinLon float64
	MinLat float64
	MaxLon float64
	MaxLat float64
}

type MoneyFeatureFilter struct {
	FeatureType     string
	Status          string
	BBox            *BBox
	UpdatedAfter    *time.Time
	IncludeArchived bool
}
