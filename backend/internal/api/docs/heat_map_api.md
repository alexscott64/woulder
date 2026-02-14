# Heat Map API Documentation

## Overview
The Heat Map API provides endpoints to visualize climbing activity across North America using Mountain Project data. The API aggregates tick data to show where climbers are actively climbing.

## Base URL
```
http://localhost:8080/api
```

## Endpoints

### 1. Get Heat Map Activity

Returns aggregated climbing activity data for the heat map visualization.

**Endpoint:** `GET /heat-map/activity`

**Query Parameters:**
- `start_date` (required): Start date in YYYY-MM-DD format
- `end_date` (required): End date in YYYY-MM-DD format
- `min_lat` (optional): Minimum latitude for bounds filtering
- `max_lat` (optional): Maximum latitude for bounds filtering
- `min_lon` (optional): Minimum longitude for bounds filtering
- `max_lon` (optional): Maximum longitude for bounds filtering
- `min_activity` (optional): Minimum tick count threshold (default: 1)
- `limit` (optional): Maximum number of points to return (default: 500, max: 1000)

**Example Request:**
```bash
curl "http://localhost:8080/api/heat-map/activity?start_date=2024-01-01&end_date=2024-12-31&min_activity=5&limit=100"
```

**Example Response:**
```json
{
  "points": [
    {
      "mp_area_id": 108123672,
      "name": "Lower Wall",
      "latitude": 47.8,
      "longitude": -121.6,
      "activity_score": 200,
      "total_ticks": 100,
      "active_routes": 50,
      "last_activity": "2024-12-15T08:00:00Z",
      "unique_climbers": 30,
      "has_subareas": true
    }
  ],
  "count": 1,
  "filters": {
    "start_date": "2024-01-01",
    "end_date": "2024-12-31",
    "min_activity": 5,
    "limit": 100
  }
}
```

**Response Fields:**
- `mp_area_id`: Mountain Project area ID
- `name`: Area name
- `latitude`, `longitude`: Geographic coordinates
- `activity_score`: Weighted score (recent activity scores higher)
- `total_ticks`: Total number of climbs in date range
- `active_routes`: Number of unique routes climbed
- `last_activity`: Timestamp of most recent climb
- `unique_climbers`: Number of distinct climbers
- `has_subareas`: Whether this area contains sub-areas

---

### 2. Get Area Activity Detail

Returns detailed activity information for a specific area.

**Endpoint:** `GET /heat-map/area/:area_id/detail`

**Path Parameters:**
- `area_id` (required): Mountain Project area ID

**Query Parameters:**
- `start_date` (required): Start date in YYYY-MM-DD format
- `end_date` (required): End date in YYYY-MM-DD format

**Example Request:**
```bash
curl "http://localhost:8080/api/heat-map/area/108123672/detail?start_date=2024-01-01&end_date=2024-12-31"
```

**Example Response:**
```json
{
  "mp_area_id": 108123672,
  "name": "Lower Wall",
  "parent_mp_area_id": 108123669,
  "latitude": 47.8,
  "longitude": -121.6,
  "total_ticks": 100,
  "active_routes": 50,
  "unique_climbers": 30,
  "last_activity": "2024-12-15T08:00:00Z",
  "recent_ticks": [
    {
      "mp_route_id": 114417549,
      "route_name": "Test Boulder",
      "rating": "V4",
      "user_name": "climber123",
      "climbed_at": "2024-12-15T08:00:00Z",
      "style": "Send",
      "comment": "Great climb!"
    }
  ],
  "recent_comments": [
    {
      "id": 123,
      "user_name": "climber456",
      "comment_text": "Conditions were perfect!",
      "commented_at": "2024-12-14T10:00:00Z",
      "mp_route_id": 114417549,
      "route_name": "Test Boulder"
    }
  ],
  "activity_timeline": [
    {
      "date": "2024-12-01T00:00:00Z",
      "tick_count": 5,
      "route_count": 3
    }
  ],
  "top_routes": [
    {
      "mp_route_id": 114417549,
      "name": "Test Boulder",
      "rating": "V4",
      "tick_count": 10,
      "last_activity": "2024-12-15T08:00:00Z"
    }
  ]
}
```

**Response Fields:**
- `recent_ticks`: Last 20 climbs in the area
- `recent_comments`: Last 10 comments on routes/areas
- `activity_timeline`: Daily aggregation of activity
- `top_routes`: Top 10 routes by activity

---

### 3. Get Routes by Bounds

Returns routes within geographic bounds with recent activity.

**Endpoint:** `GET /heat-map/routes`

**Query Parameters:**
- `min_lat` (required): Minimum latitude
- `max_lat` (required): Maximum latitude
- `min_lon` (required): Minimum longitude
- `max_lon` (required): Maximum longitude
- `start_date` (optional): Start date (default: 30 days ago)
- `end_date` (optional): End date (default: today)
- `limit` (optional): Maximum routes to return (default: 100, max: 500)

**Example Request:**
```bash
curl "http://localhost:8080/api/heat-map/routes?min_lat=47.0&max_lat=48.0&min_lon=-122.0&max_lon=-121.0&limit=50"
```

**Example Response:**
```json
{
  "routes": [
    {
      "mp_route_id": 114417549,
      "name": "Test Boulder",
      "rating": "V4",
      "latitude": 47.8,
      "longitude": -121.6,
      "tick_count": 10,
      "last_activity": "2024-12-15T08:00:00Z",
      "mp_area_id": 108123672,
      "area_name": "Lower Wall"
    }
  ],
  "count": 1
}
```

---

## Error Responses

All endpoints follow standard error response format:

```json
{
  "error": "Error message description"
}
```

**HTTP Status Codes:**
- `200 OK`: Successful request
- `400 Bad Request`: Invalid parameters (missing required params, invalid date format, etc.)
- `404 Not Found`: Resource not found (area detail endpoint only)
- `500 Internal Server Error`: Server error

**Common Error Examples:**

Missing required parameter:
```json
{
  "error": "start_date is required (format: YYYY-MM-DD)"
}
```

Invalid date format:
```json
{
  "error": "Invalid start_date format (use YYYY-MM-DD)"
}
```

Invalid bounds:
```json
{
  "error": "Invalid bounds parameters. All 4 bounds required: min_lat, max_lat, min_lon, max_lon"
}
```

Area not found:
```json
{
  "error": "Area not found"
}
```

---

## Performance Notes

- **Heat map activity endpoint**: Optimized for queries up to 1 year
- **Recommended**: Use bounds filtering for large date ranges to improve performance
- **Result limits**: Endpoints cap results (activity: 1000, routes: 500) to maintain performance
- **Query performance**: All queries use indexed columns (latitude, longitude, climbed_at)
- **Expected response times**:
  - Heat map activity: < 500ms for 500 points
  - Area detail: < 300ms
  - Routes by bounds: < 400ms

---

## Usage Examples

### Get Recent Activity in Washington State

```bash
curl "http://localhost:8080/api/heat-map/activity?\
start_date=2024-01-01&\
end_date=2024-12-31&\
min_lat=45.5&\
max_lat=49.0&\
min_lon=-124.0&\
max_lon=-117.0&\
min_activity=10&\
limit=200"
```

### Get Detailed Activity for Specific Area

```bash
curl "http://localhost:8080/api/heat-map/area/105708962/detail?\
start_date=2024-01-01&\
end_date=2024-12-31"
```

### Find Active Routes Near Leavenworth, WA

```bash
curl "http://localhost:8080/api/heat-map/routes?\
min_lat=47.5&\
max_lat=47.7&\
min_lon=-120.8&\
max_lon=-120.6&\
start_date=2024-11-01&\
end_date=2024-12-31&\
limit=100"
```

---

## Activity Score Calculation

The `activity_score` field uses recency weighting to highlight areas with recent activity:

- **Last 7 days**: Ticks count 2x (multiplier = 2.0)
- **Last 30 days**: Ticks count 1.5x (multiplier = 1.5)
- **Older**: Ticks count 1x (multiplier = 1.0)

**Formula:** `activity_score = total_ticks * recency_multiplier`

This ensures that areas with recent activity appear "hotter" on the map, even if their absolute tick count is lower than areas with older activity.

---

## Integration Notes

### Frontend Integration

When building a map interface:

1. **Initial Load**: Fetch activity data with a reasonable date range (last 30-90 days)
2. **Map Interactions**: Update bounds parameters when user pans/zooms
3. **Filtering**: Allow users to adjust date range and min_activity threshold
4. **Detail Views**: Fetch area detail when user clicks a heat map point
5. **Caching**: Use React Query or similar with 5-10 minute stale time

### Rate Limiting

Currently no rate limiting is implemented. Consider implementing if needed for production use.

---

## Future Enhancements

Potential future additions (not yet implemented):

- Heatmap intensity visualization (continuous gradient vs discrete points)
- Time-lapse animation of activity over time
- Grade filtering (e.g., only show V4-V8 routes)
- Weather overlay integration
- Export functionality (CSV, GeoJSON)
