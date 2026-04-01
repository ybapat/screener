package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8080"

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	AgeRange    string `json:"age_range,omitempty"`
	Country     string `json:"country,omitempty"`
}

type loginResp struct {
	Data struct {
		Tokens struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
	} `json:"data"`
}

type screenTimeRecord struct {
	AppName     string `json:"app_name"`
	AppCategory string `json:"app_category"`
	DurationSec int    `json:"duration_secs"`
	StartedAt   string `json:"started_at"`
	EndedAt     string `json:"ended_at"`
	DeviceType  string `json:"device_type"`
	OS          string `json:"os"`
}

var (
	apps = []struct {
		name     string
		category string
	}{
		{"Instagram", "social_media"}, {"TikTok", "social_media"}, {"Twitter", "social_media"},
		{"Facebook", "social_media"}, {"Reddit", "social_media"}, {"LinkedIn", "social_media"},
		{"YouTube", "entertainment"}, {"Netflix", "entertainment"}, {"Spotify", "entertainment"},
		{"Twitch", "entertainment"}, {"Disney+", "entertainment"},
		{"Slack", "productivity"}, {"VS Code", "productivity"}, {"Notion", "productivity"},
		{"Gmail", "productivity"}, {"Zoom", "productivity"}, {"Teams", "productivity"},
		{"Chrome", "browser"}, {"Safari", "browser"}, {"Firefox", "browser"},
		{"WhatsApp", "messaging"}, {"Discord", "messaging"}, {"Telegram", "messaging"},
		{"Uber", "transportation"}, {"Maps", "utilities"},
		{"DoorDash", "food_delivery"}, {"Amazon", "shopping"},
		{"Robinhood", "finance"}, {"Venmo", "finance"},
	}

	ageRanges = []string{"18-24", "25-34", "35-44", "45-54"}
	countries = []string{"US", "UK", "CA", "DE", "FR", "JP", "AU", "IN", "BR"}
	devices   = []string{"phone", "tablet", "desktop"}
	oses      = []string{"ios", "android", "macos", "windows"}

	firstNames = []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry", "Iris", "Jack"}
	lastNames  = []string{"Smith", "Johnson", "Lee", "Chen", "Patel", "Brown", "Kim", "Garcia", "Taylor", "Wilson"}
)

func main() {
	fmt.Println("=== Screener Seed Data Generator ===")
	fmt.Println()

	// Create 10 sellers
	var sellerTokens []string
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("%s %s", firstNames[i], lastNames[i])
		email := fmt.Sprintf("seller%d@test.com", i+1)
		token := registerUser(registerReq{
			Email:       email,
			Password:    "password123",
			DisplayName: name,
			Role:        "seller",
			AgeRange:    ageRanges[rand.Intn(len(ageRanges))],
			Country:     countries[rand.Intn(len(countries))],
		})
		if token != "" {
			sellerTokens = append(sellerTokens, token)
			fmt.Printf("  Created seller: %s (%s)\n", name, email)
		}
	}

	// Create 3 buyers
	var buyerTokens []string
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("Buyer %d", i+1)
		email := fmt.Sprintf("buyer%d@test.com", i+1)
		token := registerUser(registerReq{
			Email:       email,
			Password:    "password123",
			DisplayName: name,
			Role:        "buyer",
		})
		if token != "" {
			buyerTokens = append(buyerTokens, token)
			fmt.Printf("  Created buyer: %s (%s)\n", name, email)
		}
	}

	fmt.Println()
	fmt.Println("Uploading screen time data for sellers...")

	// Upload 30 days of data per seller
	for i, token := range sellerTokens {
		recordCount := 0
		for day := 0; day < 30; day++ {
			date := time.Now().AddDate(0, 0, -day)
			records := generateDayRecords(date)
			uploadData(token, records)
			recordCount += len(records)
		}
		fmt.Printf("  Seller %d: uploaded %d records across 30 days\n", i+1, recordCount)
	}

	// Top up buyer credits
	fmt.Println()
	fmt.Println("Topping up buyer credits...")
	for i, token := range buyerTokens {
		topupCredits(token, 50000)
		fmt.Printf("  Buyer %d: topped up 50000 credits\n", i+1)
	}

	fmt.Println()
	fmt.Println("=== Seed complete! ===")
	fmt.Println()
	fmt.Println("Test accounts:")
	fmt.Println("  Sellers: seller1@test.com through seller10@test.com (password: password123)")
	fmt.Println("  Buyers:  buyer1@test.com through buyer3@test.com (password: password123)")
}

func registerUser(req registerReq) string {
	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("  ERROR registering %s: %v\n", req.Email, err)
		return ""
	}
	defer resp.Body.Close()

	var result loginResp
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.Tokens.AccessToken
}

func uploadData(token string, records []screenTimeRecord) {
	body, _ := json.Marshal(map[string]any{"records": records})
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/data/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func topupCredits(token string, amount int64) {
	body, _ := json.Marshal(map[string]int64{"amount": amount})
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/credits/topup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func generateDayRecords(date time.Time) []screenTimeRecord {
	numRecords := 5 + rand.Intn(15) // 5-20 app sessions per day
	records := make([]screenTimeRecord, 0, numRecords)

	currentTime := time.Date(date.Year(), date.Month(), date.Day(), 7, 0, 0, 0, time.UTC)

	for i := 0; i < numRecords; i++ {
		app := apps[rand.Intn(len(apps))]
		duration := 60 + rand.Intn(3540) // 1 min to ~60 min
		device := devices[rand.Intn(len(devices))]

		os := "ios"
		if device == "desktop" {
			os = oses[2+rand.Intn(2)] // macos or windows
		} else {
			os = oses[rand.Intn(2)] // ios or android
		}

		startedAt := currentTime.Add(time.Duration(rand.Intn(60)) * time.Minute)
		endedAt := startedAt.Add(time.Duration(duration) * time.Second)

		records = append(records, screenTimeRecord{
			AppName:     app.name,
			AppCategory: app.category,
			DurationSec: duration,
			StartedAt:   startedAt.Format(time.RFC3339),
			EndedAt:     endedAt.Format(time.RFC3339),
			DeviceType:  device,
			OS:          os,
		})

		currentTime = endedAt.Add(time.Duration(5+rand.Intn(30)) * time.Minute)
	}

	return records
}
