package models

import "time"

// KayaUser represents a Kaya user profile
type KayaUser struct {
	ID                       int       `json:"id" db:"id"`
	KayaUserID               string    `json:"kaya_user_id" db:"kaya_user_id"`
	Username                 string    `json:"username" db:"username"`
	Fname                    *string   `json:"fname,omitempty" db:"fname"`
	Lname                    *string   `json:"lname,omitempty" db:"lname"`
	PhotoURL                 *string   `json:"photo_url,omitempty" db:"photo_url"`
	Bio                      *string   `json:"bio,omitempty" db:"bio"`
	Height                   *int      `json:"height,omitempty" db:"height"`
	ApeIndex                 *float64  `json:"ape_index,omitempty" db:"ape_index"`
	LimitGradeBoulderingID   *string   `json:"limit_grade_bouldering_id,omitempty" db:"limit_grade_bouldering_id"`
	LimitGradeBoulderingName *string   `json:"limit_grade_bouldering_name,omitempty" db:"limit_grade_bouldering_name"`
	LimitGradeRoutesID       *string   `json:"limit_grade_routes_id,omitempty" db:"limit_grade_routes_id"`
	LimitGradeRoutesName     *string   `json:"limit_grade_routes_name,omitempty" db:"limit_grade_routes_name"`
	IsPrivate                bool      `json:"is_private" db:"is_private"`
	IsPremium                bool      `json:"is_premium" db:"is_premium"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
}

// KayaLocation represents a Kaya location (destination or area)
type KayaLocation struct {
	ID                                int        `json:"id" db:"id"`
	KayaLocationID                    string     `json:"kaya_location_id" db:"kaya_location_id"`
	Slug                              string     `json:"slug" db:"slug"`
	Name                              string     `json:"name" db:"name"`
	Latitude                          *float64   `json:"latitude,omitempty" db:"latitude"`
	Longitude                         *float64   `json:"longitude,omitempty" db:"longitude"`
	PhotoURL                          *string    `json:"photo_url,omitempty" db:"photo_url"`
	Description                       *string    `json:"description,omitempty" db:"description"`
	LocationTypeID                    *string    `json:"location_type_id,omitempty" db:"location_type_id"`
	LocationTypeName                  *string    `json:"location_type_name,omitempty" db:"location_type_name"`
	ParentLocationID                  *string    `json:"parent_location_id,omitempty" db:"parent_location_id"`
	ParentLocationSlug                *string    `json:"parent_location_slug,omitempty" db:"parent_location_slug"`
	ParentLocationName                *string    `json:"parent_location_name,omitempty" db:"parent_location_name"`
	ClimbCount                        int        `json:"climb_count" db:"climb_count"`
	BoulderCount                      int        `json:"boulder_count" db:"boulder_count"`
	RouteCount                        int        `json:"route_count" db:"route_count"`
	AscentCount                       int        `json:"ascent_count" db:"ascent_count"`
	IsGBModeratedBouldering           bool       `json:"is_gb_moderated_bouldering" db:"is_gb_moderated_bouldering"`
	IsGBModeratedRoutes               bool       `json:"is_gb_moderated_routes" db:"is_gb_moderated_routes"`
	IsAccessSensitive                 bool       `json:"is_access_sensitive" db:"is_access_sensitive"`
	IsClosed                          bool       `json:"is_closed" db:"is_closed"`
	HasMapsDisabled                   bool       `json:"has_maps_disabled" db:"has_maps_disabled"`
	ClosedDate                        *time.Time `json:"closed_date,omitempty" db:"closed_date"`
	DescriptionBouldering             *string    `json:"description_bouldering,omitempty" db:"description_bouldering"`
	DescriptionRoutes                 *string    `json:"description_routes,omitempty" db:"description_routes"`
	DescriptionShortBouldering        *string    `json:"description_short_bouldering,omitempty" db:"description_short_bouldering"`
	DescriptionShortRoutes            *string    `json:"description_short_routes,omitempty" db:"description_short_routes"`
	AccessDescriptionBouldering       *string    `json:"access_description_bouldering,omitempty" db:"access_description_bouldering"`
	AccessDescriptionRoutes           *string    `json:"access_description_routes,omitempty" db:"access_description_routes"`
	AccessIssuesDescriptionBouldering *string    `json:"access_issues_description_bouldering,omitempty" db:"access_issues_description_bouldering"`
	AccessIssuesDescriptionRoutes     *string    `json:"access_issues_description_routes,omitempty" db:"access_issues_description_routes"`
	ClimbTypeID                       *string    `json:"climb_type_id,omitempty" db:"climb_type_id"`
	WoulderLocationID                 *int       `json:"woulder_location_id,omitempty" db:"woulder_location_id"`
	LastSyncedAt                      *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	CreatedAt                         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                         time.Time  `json:"updated_at" db:"updated_at"`
}

// KayaClimb represents a Kaya climb (route or boulder)
type KayaClimb struct {
	ID                  int        `json:"id" db:"id"`
	KayaClimbID         *string    `json:"kaya_climb_id,omitempty" db:"kaya_climb_id"`
	Slug                string     `json:"slug" db:"slug"`
	Name                string     `json:"name" db:"name"`
	GradeID             *string    `json:"grade_id,omitempty" db:"grade_id"`
	GradeName           *string    `json:"grade_name,omitempty" db:"grade_name"`
	GradeOrdering       *int       `json:"grade_ordering,omitempty" db:"grade_ordering"`
	GradeClimbTypeID    *string    `json:"grade_climb_type_id,omitempty" db:"grade_climb_type_id"`
	ClimbTypeID         *string    `json:"climb_type_id,omitempty" db:"climb_type_id"`
	ClimbTypeName       *string    `json:"climb_type_name,omitempty" db:"climb_type_name"`
	Rating              *float64   `json:"rating,omitempty" db:"rating"`
	AscentCount         int        `json:"ascent_count" db:"ascent_count"`
	KayaDestinationID   *string    `json:"kaya_destination_id,omitempty" db:"kaya_destination_id"`
	KayaDestinationName *string    `json:"kaya_destination_name,omitempty" db:"kaya_destination_name"`
	KayaAreaID          *string    `json:"kaya_area_id,omitempty" db:"kaya_area_id"`
	KayaAreaName        *string    `json:"kaya_area_name,omitempty" db:"kaya_area_name"`
	ColorName           *string    `json:"color_name,omitempty" db:"color_name"`
	GymName             *string    `json:"gym_name,omitempty" db:"gym_name"`
	BoardName           *string    `json:"board_name,omitempty" db:"board_name"`
	IsGBModerated       bool       `json:"is_gb_moderated" db:"is_gb_moderated"`
	IsAccessSensitive   bool       `json:"is_access_sensitive" db:"is_access_sensitive"`
	IsClosed            bool       `json:"is_closed" db:"is_closed"`
	IsOffensive         bool       `json:"is_offensive" db:"is_offensive"`
	WoulderLocationID   *int       `json:"woulder_location_id,omitempty" db:"woulder_location_id"`
	LastSyncedAt        *time.Time `json:"last_synced_at,omitempty" db:"last_synced_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// KayaAscent represents a Kaya ascent (tick)
type KayaAscent struct {
	ID            int       `json:"id" db:"id"`
	KayaAscentID  string    `json:"kaya_ascent_id" db:"kaya_ascent_id"`
	KayaClimbSlug string    `json:"kaya_climb_slug" db:"kaya_climb_slug"`
	KayaUserID    string    `json:"kaya_user_id" db:"kaya_user_id"`
	Date          time.Time `json:"date" db:"date"`
	Comment       *string   `json:"comment,omitempty" db:"comment"`
	Rating        *int      `json:"rating,omitempty" db:"rating"`
	Stiffness     *int      `json:"stiffness,omitempty" db:"stiffness"`
	GradeID       *string   `json:"grade_id,omitempty" db:"grade_id"`
	GradeName     *string   `json:"grade_name,omitempty" db:"grade_name"`
	PhotoURL      *string   `json:"photo_url,omitempty" db:"photo_url"`
	PhotoThumbURL *string   `json:"photo_thumb_url,omitempty" db:"photo_thumb_url"`
	VideoURL      *string   `json:"video_url,omitempty" db:"video_url"`
	VideoThumbURL *string   `json:"video_thumb_url,omitempty" db:"video_thumb_url"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// KayaPost represents a Kaya post (user content)
type KayaPost struct {
	ID          int       `json:"id" db:"id"`
	KayaPostID  string    `json:"kaya_post_id" db:"kaya_post_id"`
	KayaUserID  string    `json:"kaya_user_id" db:"kaya_user_id"`
	DateCreated time.Time `json:"date_created" db:"date_created"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// KayaPostItem represents a single item within a Kaya post
type KayaPostItem struct {
	ID                int       `json:"id" db:"id"`
	KayaPostItemID    string    `json:"kaya_post_item_id" db:"kaya_post_item_id"`
	KayaPostID        string    `json:"kaya_post_id" db:"kaya_post_id"`
	KayaClimbSlug     *string   `json:"kaya_climb_slug,omitempty" db:"kaya_climb_slug"`
	KayaAscentID      *string   `json:"kaya_ascent_id,omitempty" db:"kaya_ascent_id"`
	PhotoURL          *string   `json:"photo_url,omitempty" db:"photo_url"`
	VideoURL          *string   `json:"video_url,omitempty" db:"video_url"`
	VideoThumbnailURL *string   `json:"video_thumbnail_url,omitempty" db:"video_thumbnail_url"`
	Caption           *string   `json:"caption,omitempty" db:"caption"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// KayaSyncProgress tracks sync status for locations
type KayaSyncProgress struct {
	ID                 int        `json:"id" db:"id"`
	KayaLocationID     string     `json:"kaya_location_id" db:"kaya_location_id"`
	LocationName       string     `json:"location_name" db:"location_name"`
	Status             string     `json:"status" db:"status"` // 'pending', 'in_progress', 'completed', 'failed'
	LastSyncAt         *time.Time `json:"last_sync_at,omitempty" db:"last_sync_at"`
	NextSyncAt         *time.Time `json:"next_sync_at,omitempty" db:"next_sync_at"`
	SyncError          *string    `json:"sync_error,omitempty" db:"sync_error"`
	ClimbsSynced       int        `json:"climbs_synced" db:"climbs_synced"`
	AscentsSynced      int        `json:"ascents_synced" db:"ascents_synced"`
	SubLocationsSynced int        `json:"sub_locations_synced" db:"sub_locations_synced"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}
