package rivers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/alexscott64/woulder/backend/internal/models"
)

const usgsWaterServicesURL = "https://waterservices.usgs.gov/nwis/iv/"

type USGSClient struct {
	httpClient *http.Client
}

// USGS API response structures
type usgsResponse struct {
	Value struct {
		TimeSeries []struct {
			SourceInfo struct {
				SiteName string `json:"siteName"`
			} `json:"sourceInfo"`
			Variable struct {
				VariableCode []struct {
					Value string `json:"value"`
				} `json:"variableCode"`
				VariableName        string `json:"variableName"`
				VariableDescription string `json:"variableDescription"`
				Unit                struct {
					UnitCode string `json:"unitCode"`
				} `json:"unit"`
			} `json:"variable"`
			Values []struct {
				Value []struct {
					Value    string `json:"value"`
					DateTime string `json:"dateTime"`
				} `json:"value"`
			} `json:"values"`
		} `json:"timeSeries"`
	} `json:"value"`
}

func NewUSGSClient() *USGSClient {
	return &USGSClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// EstimateFlowFromDrainageRatio estimates flow using drainage area ratio method
// Formula: Q_small = Q_large × (A_small / A_large)^0.7
// The 0.7 exponent accounts for non-linear scaling in mountainous watersheds
func EstimateFlowFromDrainageRatio(gaugeFlowCFS, riverDrainageAreaSqMi, gaugeDrainageAreaSqMi float64) float64 {
	if gaugeDrainageAreaSqMi == 0 {
		return gaugeFlowCFS // Avoid division by zero
	}

	ratio := riverDrainageAreaSqMi / gaugeDrainageAreaSqMi
	// Use power of 0.7 for mountainous terrain (0.6-0.8 is typical range)
	scaleFactor := math.Pow(ratio, 0.7)

	return gaugeFlowCFS * scaleFactor
}

// GetRiverData fetches current river gauge data from USGS
func (c *USGSClient) GetRiverData(gaugeID string) (float64, float64, string, error) {
	// Fetch discharge (flow in CFS) and gauge height
	url := fmt.Sprintf("%s?format=json&sites=%s&parameterCd=00060,00065&siteStatus=active",
		usgsWaterServicesURL, gaugeID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to fetch USGS data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, 0, "", fmt.Errorf("USGS API error (status %d): %s", resp.StatusCode, string(body))
	}

	var data usgsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, "", fmt.Errorf("failed to decode USGS response: %w", err)
	}

	var flowCFS float64
	var gaugeHeightFt float64
	var timestamp string

	// Parse the time series data
	for _, ts := range data.Value.TimeSeries {
		if len(ts.Values) == 0 || len(ts.Values[0].Value) == 0 {
			continue
		}

		latestValue := ts.Values[0].Value[0]
		timestamp = latestValue.DateTime

		// Check variable code to determine if it's discharge or gauge height
		if len(ts.Variable.VariableCode) > 0 {
			varCode := ts.Variable.VariableCode[0].Value

			switch varCode {
			case "00060": // Discharge (CFS)
				flowCFS, err = strconv.ParseFloat(latestValue.Value, 64)
				if err != nil {
					log.Printf("Failed to parse flow value: %v", err)
				}
			case "00065": // Gauge height (ft)
				gaugeHeightFt, err = strconv.ParseFloat(latestValue.Value, 64)
				if err != nil {
					log.Printf("Failed to parse gauge height value: %v", err)
				}
			}
		}
	}

	if flowCFS == 0 {
		return 0, 0, "", fmt.Errorf("no flow data available for gauge %s", gaugeID)
	}

	return flowCFS, gaugeHeightFt, timestamp, nil
}

// CalculateCrossingStatus determines if a river crossing is safe based on flow
func CalculateCrossingStatus(river models.River, flowCFS float64) (string, string, bool, float64) {
	percentOfSafe := (flowCFS / float64(river.SafeCrossingCFS)) * 100

	var status string
	var message string
	var isSafe bool

	// Determine base status
	if flowCFS <= float64(river.SafeCrossingCFS) {
		status = "safe"
		message = fmt.Sprintf("Safe to cross. Flow is %.0f%% of safe threshold.", percentOfSafe)
		isSafe = true
	} else if flowCFS <= float64(river.CautionCrossingCFS) {
		status = "caution"
		message = fmt.Sprintf("Use caution. Flow is %.0f%% of safe threshold.", percentOfSafe)
		isSafe = false
	} else {
		status = "unsafe"
		message = fmt.Sprintf("Unsafe to cross! Flow is %.0f%% of safe threshold.", percentOfSafe)
		isSafe = false
	}

	// Prefix status with "estimated" if flow is estimated
	if river.IsEstimated {
		status = "estimated " + status
	}

	return status, message, isSafe, percentOfSafe
}
