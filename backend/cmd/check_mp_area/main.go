package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/alexscott64/woulder/backend/internal/mountainproject"
)

func main() {
	client := mountainproject.NewClient()

	// Check Stawamus Chief
	areaID := "105805895"
	log.Printf("Fetching area %s (Stawamus Chief)...", areaID)

	area, err := client.GetArea(areaID)
	if err != nil {
		log.Fatalf("Failed to fetch area: %v", err)
	}

	fmt.Println("\n========================================")
	fmt.Printf("Area: %s (ID: %d)\n", area.Title, area.ID)
	fmt.Printf("Type: %s\n", area.Type)
	fmt.Printf("GPS: [%.6f, %.6f]\n", area.Coordinates[1], area.Coordinates[0])
	fmt.Println("========================================\n")

	fmt.Printf("Children (%d total):\n\n", len(area.Children))

	subareas := []mountainproject.ChildElement{}
	routes := []mountainproject.ChildElement{}

	for _, child := range area.Children {
		if child.Type == "Area" {
			subareas = append(subareas, child)
		} else {
			routes = append(routes, child)
		}
	}

	if len(subareas) > 0 {
		fmt.Printf("SUB-AREAS (%d):\n", len(subareas))
		for i, subarea := range subareas {
			fmt.Printf("  %d. %s (ID: %d)\n", i+1, subarea.Title, subarea.ID)
		}
		fmt.Println()
	}

	if len(routes) > 0 {
		fmt.Printf("ROUTES (%d):\n", len(routes))
		for i, route := range routes {
			routeTypes := route.RouteTypes
			fmt.Printf("  %d. %s (ID: %d) - Types: %v\n", i+1, route.Title, route.ID, routeTypes)
		}
	}

	// Pretty print full JSON for inspection
	fmt.Println("\n========================================")
	fmt.Println("Full JSON Response:")
	fmt.Println("========================================")
	jsonBytes, _ := json.MarshalIndent(area, "", "  ")
	fmt.Println(string(jsonBytes))
}
