package config

import (
	"os"
	"strconv"
)

// Config represents runtime application configuration.
type Config struct {
	Port           string
	DBPath         string
	StripeKey      string
	HubSpotKey     string
	NotionKey      string
	NotionPageID   string
	MinClusterSize int
	MinPoints      int
}

// LoadConfig loads configuration from environment variables with sensible defaults.
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./feedbackloom.db"
	}

	minClusterSize := 2
	if val := os.Getenv("MIN_CLUSTER_SIZE"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 2 {
			minClusterSize = i
		}
	}

	minPoints := 2
	if val := os.Getenv("MIN_POINTS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil && i >= 2 {
			minPoints = i
		}
	}

	return &Config{
		Port:           port,
		DBPath:         dbPath,
		StripeKey:      os.Getenv("STRIPE_KEY"),
		HubSpotKey:     os.Getenv("HUBSPOT_KEY"),
		NotionKey:      os.Getenv("NOTION_KEY"),
		NotionPageID:   os.Getenv("NOTION_PAGE_ID"),
		MinClusterSize: minClusterSize,
		MinPoints:      minPoints,
	}
}
