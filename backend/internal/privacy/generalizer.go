package privacy

import (
	"strings"
	"time"
)

// appCategoryMap maps known app names to their categories.
var appCategoryMap = map[string]string{
	"instagram":  "social_media",
	"facebook":   "social_media",
	"twitter":    "social_media",
	"tiktok":     "social_media",
	"snapchat":   "social_media",
	"reddit":     "social_media",
	"linkedin":   "social_media",
	"threads":    "social_media",
	"youtube":    "entertainment",
	"netflix":    "entertainment",
	"hulu":       "entertainment",
	"disney+":    "entertainment",
	"spotify":    "entertainment",
	"twitch":     "entertainment",
	"slack":      "productivity",
	"teams":      "productivity",
	"zoom":       "productivity",
	"notion":     "productivity",
	"vscode":     "productivity",
	"vs code":    "productivity",
	"gmail":      "productivity",
	"outlook":    "productivity",
	"chrome":     "browser",
	"safari":     "browser",
	"firefox":    "browser",
	"edge":       "browser",
	"maps":       "utilities",
	"weather":    "utilities",
	"calculator": "utilities",
	"settings":   "utilities",
	"whatsapp":   "messaging",
	"imessage":   "messaging",
	"telegram":   "messaging",
	"signal":     "messaging",
	"discord":    "messaging",
	"uber":       "transportation",
	"lyft":       "transportation",
	"doordash":   "food_delivery",
	"ubereats":   "food_delivery",
	"grubhub":    "food_delivery",
	"amazon":     "shopping",
	"ebay":       "shopping",
	"walmart":    "shopping",
	"robinhood":  "finance",
	"venmo":      "finance",
	"cashapp":    "finance",
	"mint":       "finance",
}

// GeneralizeAppName maps a raw app name to its category.
// Falls back to "other" if not found.
func GeneralizeAppName(appName string) string {
	lower := strings.ToLower(strings.TrimSpace(appName))
	if cat, ok := appCategoryMap[lower]; ok {
		return cat
	}
	return "other"
}

// GeneralizeDuration maps exact seconds to a range bucket.
func GeneralizeDuration(secs int) string {
	switch {
	case secs < 60:
		return "0-1m"
	case secs < 300:
		return "1-5m"
	case secs < 900:
		return "5-15m"
	case secs < 1800:
		return "15-30m"
	case secs < 3600:
		return "30-60m"
	default:
		return "60m+"
	}
}

// GeneralizeTimestamp maps a timestamp to a time-of-day bucket.
func GeneralizeTimestamp(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

// GeneralizeAge maps an exact age to a range bucket.
func GeneralizeAge(ageRange string) string {
	// Already generalized from user profile, pass through
	switch ageRange {
	case "13-17", "18-24", "25-34", "35-44", "45-54", "55-64", "65+":
		return ageRange
	default:
		return "unknown"
	}
}
