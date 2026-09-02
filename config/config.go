package config

import "time"

const (
	AniListGraphQLURL     = "https://graphql.anilist.co"
	ClientTimeout         = 15 * time.Second
	ClientLimiterDuration = 1 * time.Second
	Debug                 = true
)
