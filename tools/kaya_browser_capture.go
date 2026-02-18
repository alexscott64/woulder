package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// CapturedRequest stores information about a network request
type CapturedRequest struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	RequestBody string            `json:"requestBody,omitempty"`
	Headers     map[string]string `json:"headers"`
	Response    string            `json:"response,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

func main() {
	log.Println("=======================================================================")
	log.Println("Kaya GraphQL API Analyzer - Headless Browser Approach")
	log.Println("=======================================================================")
	log.Println()
	log.Println("Using headless Chrome to bypass Cloudflare and capture GraphQL queries...")
	log.Println()

	// Allocate a new browser context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create context
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set a longer timeout for Cloudflare challenges
	ctx, cancel = context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	// Store captured requests
	capturedRequests := make([]*CapturedRequest, 0)
	requestIDs := make(map[network.RequestID]string)

	// Listen for network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			// Capture GraphQL requests
			if strings.Contains(ev.Request.URL, "/graphql") {
				log.Printf("📡 GraphQL request detected: %s %s", ev.Request.Method, ev.Request.URL)

				headers := make(map[string]string)
				for key, val := range ev.Request.Headers {
					if str, ok := val.(string); ok {
						headers[key] = str
					}
				}

				req := &CapturedRequest{
					URL:       ev.Request.URL,
					Method:    ev.Request.Method,
					Headers:   headers,
					Timestamp: time.Now(),
				}

				// Store request for later matching with response
				requestIDs[ev.RequestID] = ev.Request.URL
				capturedRequests = append(capturedRequests, req)
			}

		case *network.EventRequestWillBeSentExtraInfo:
			// This event contains the actual POST data
			log.Printf("📝 Extra info for request")

		case *network.EventLoadingFinished:
			// Check if this is one of our GraphQL requests
			if url, ok := requestIDs[ev.RequestID]; ok && strings.Contains(url, "/graphql") {
				log.Printf("✓ GraphQL request finished: %s", url)
			}
		}
	})

	var pageHTML string
	var cookies []*network.Cookie

	log.Println("🌐 Navigating to Kaya location page...")

	// Execute the browsing sequence
	err := chromedp.Run(ctx,
		network.Enable(),

		// Navigate to the page
		chromedp.Navigate("https://kaya-app.kayaclimb.com/location/Leavenworth-344933"),

		// Wait for body to be visible (page loaded)
		chromedp.WaitVisible(`body`, chromedp.ByQuery),

		// Wait for content to load
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("✓ Page loaded, waiting for JavaScript to execute...")
			return nil
		}),
		chromedp.Sleep(8*time.Second),

		// Scroll to trigger lazy loading
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight/2)`, nil),
		chromedp.Sleep(2*time.Second),

		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
		chromedp.Sleep(3*time.Second),

		// Try to find and click on things that might trigger more GraphQL requests
		chromedp.Evaluate(`
			// Try to find and click elements
			console.log('Looking for interactive elements...');
			const links = document.querySelectorAll('a[href*="climb"]');
			console.log('Found', links.length, 'climb links');
		`, nil),

		// Get the page HTML for manual analysis
		chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),

		// Get cookies
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),
	)

	if err != nil {
		log.Fatal("❌ Error during browser automation:", err)
	}

	log.Println()
	log.Println("=======================================================================")
	log.Println("Capture Complete")
	log.Println("=======================================================================")
	log.Printf("📊 Captured %d network requests\n", len(capturedRequests))
	log.Printf("🍪 Captured %d cookies\n", len(cookies))
	log.Println()

	// Save captured requests
	if len(capturedRequests) > 0 {
		outputFile := "../docs/kaya-captured-requests.json"
		data, err := json.MarshalIndent(capturedRequests, "", "  ")
		if err != nil {
			log.Fatal("Error marshaling requests:", err)
		}

		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			log.Fatal("Error writing requests file:", err)
		}

		log.Printf("✅ Saved captured requests to: %s\n", outputFile)

		for i, req := range capturedRequests {
			log.Printf("\n  Request %d:", i+1)
			log.Printf("    URL: %s", req.URL)
			log.Printf("    Method: %s", req.Method)
			log.Printf("    Time: %s", req.Timestamp.Format(time.RFC3339))
		}
	}

	// Save cookies
	if len(cookies) > 0 {
		cookiesFile := "../docs/kaya-cookies.json"
		data, err := json.MarshalIndent(cookies, "", "  ")
		if err != nil {
			log.Fatal("Error marshaling cookies:", err)
		}

		if err := os.WriteFile(cookiesFile, data, 0644); err != nil {
			log.Fatal("Error writing cookies file:", err)
		}

		log.Printf("\n✅ Saved %d cookies to: %s\n", len(cookies), cookiesFile)
	}

	// Save page HTML for manual inspection
	htmlFile := "../docs/kaya-page-source.html"
	if err := os.WriteFile(htmlFile, []byte(pageHTML), 0644); err != nil {
		log.Fatal("Error writing HTML file:", err)
	}
	log.Printf("✅ Saved page HTML to: %s (for manual analysis)\n", htmlFile)

	// Extract and analyze script tags for GraphQL queries
	log.Println("\n🔍 Analyzing page for embedded GraphQL queries...")
	if strings.Contains(pageHTML, "graphql") || strings.Contains(pageHTML, "query") {
		log.Println("✓ Page contains GraphQL-related content")
		log.Println("  Check the HTML file for inline queries or JavaScript bundles")
	}

	log.Println()
	log.Println("=======================================================================")
	log.Println("Next Steps:")
	log.Println("=======================================================================")
	log.Println("1. Review ../docs/kaya-page-source.html in a text editor")
	log.Println("2. Look for GraphQL queries in <script> tags or inline JavaScript")
	log.Println("3. Use browser DevTools manually if automated capture missed queries")
	log.Println("4. The captured cookies can be used in the Go GraphQL client")
	log.Println()

	if len(capturedRequests) == 0 {
		log.Println("⚠️  No GraphQL requests were captured automatically.")
		log.Println("   This likely means:")
		log.Println("   - Queries are embedded in JavaScript bundles")
		log.Println("   - Queries are triggered by user interactions we didn't simulate")
		log.Println("   - The app uses a different API pattern")
		log.Println()
		log.Println("   Recommended: Use browser DevTools manually to capture queries")
		log.Println("   See: ../docs/kaya-capture-guide.md")
	}
}
