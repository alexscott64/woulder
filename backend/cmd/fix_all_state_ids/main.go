package main

import (
	"context"
	"fmt"
	"log"

	"github.com/alexscott64/woulder/backend/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")

	db, err := database.New()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// All correct IDs from mountainproject.com/route-guide
	correctIDs := map[string]string{
		"Alabama":        "105905173",
		"Alaska":         "105909311",
		"Arizona":        "105708962",
		"Arkansas":       "105901027",
		"California":     "105708959",
		"Colorado":       "105708956",
		"Connecticut":    "105806977", // was 105908076
		"Delaware":       "106861605", // was 108249388
		"Florida":        "111721391",
		"Georgia":        "105897947",
		"Hawaii":         "106316122",
		"Idaho":          "105708958",
		"Illinois":       "105911816",
		"Indiana":        "112389571", // was 105908701
		"Iowa":           "106092653",
		"Kansas":         "107235316", // was 105910743
		"Kentucky":       "105868674",
		"Louisiana":      "116720343", // was 105910197
		"Maine":          "105948977",
		"Maryland":       "106029417", // was 105908492
		"Massachusetts":  "105908062", // was 105907214
		"Michigan":       "106113246",
		"Minnesota":      "105812481", // was 105910238
		"Mississippi":    "108307056", // was 105907550
		"Missouri":       "105899020", // was 105901239
		"Montana":        "105907492",
		"Nebraska":       "116096758", // was 105911954
		"Nevada":         "105708961",
		"New Hampshire":  "105872225",
		"New Jersey":     "106374428", // was 105909408
		"New Mexico":     "105708964",
		"New York":       "105800424",
		"North Carolina": "105873282", // was 105905638
		"North Dakota":   "106598130", // was 105912378
		"Ohio":           "105994953",
		"Oklahoma":       "105854466", // was 105911215
		"Oregon":         "105708965", // was 105708960 - KEY FIX!
		"Pennsylvania":   "105913279", // was 105908143
		"Rhode Island":   "106842810", // was 105912449
		"South Carolina": "107638915", // was 105906948
		"South Dakota":   "105708963",
		"Tennessee":      "105887760", // was 105905667
		"Texas":          "105835804", // was 105709046
		"Utah":           "105708957",
		"Vermont":        "105891603",
		"Virginia":       "105852400", // was 105906523
		"Washington":     "105708966",
		"West Virginia":  "105855459", // was 105907015
		"Wisconsin":      "105708968",
		"Wyoming":        "105708960",
	}

	updated := 0
	failed := 0

	for state, id := range correctIDs {
		err := db.UpdateStateConfigMPAreaID(ctx, state, id)
		if err != nil {
			log.Printf("❌ Error updating %s: %v", state, err)
			failed++
		} else {
			log.Printf("✓ Updated %s to %s", state, id)
			updated++
		}
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("✓ Successfully updated: %d states\n", updated)
	if failed > 0 {
		fmt.Printf("❌ Failed: %d states\n", failed)
	}
	fmt.Printf("========================================\n")
}
