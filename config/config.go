package config

import "time"

const (
	Debug = true

	// Server

	AniListGraphQLURL     = "https://graphql.anilist.co"
	ClientTimeout         = 15 * time.Second
	ClientLimiterDuration = 1 * time.Second

	// Importer

	MaxJobInMemoryPeriod = 24 * time.Hour

	// Layout

	LayoutColumnWidth float64 = 220
	LayoutRowHeight   float64 = 110
	LayoutNodeWidth   float64 = 200
	LayoutNodeHeight  float64 = 48
	LayoutPadX        float64 = 40
	LayoutPadY        float64 = 40
)
