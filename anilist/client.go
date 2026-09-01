package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"visuilizer/config"
	"visuilizer/media"
)

type Client struct {
	HTTP    *http.Client
	Limiter *time.Ticker
}

func NewClient() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: config.ClientTimeout},
		Limiter: time.NewTicker(config.ClientLimiterDuration),
	}
}

func (c *Client) FetchMedia(id int) (media.Entry, []media.Relation, error) {
	<-c.Limiter.C
	body, err := json.Marshal(GraphQLRequest{
		Query:     mediaQuery,
		Variables: map[string]any{"id": id},
	})

	if err != nil {
		return media.Entry{}, nil, err
	}

	req, err := http.NewRequest("POST", config.AniListGraphQLURL, bytes.NewReader(body))
	if err != nil {
		return media.Entry{}, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; visuilizer/0.1)")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return media.Entry{}, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return media.Entry{}, nil, err
	}

	var gqlResp GraphQLResponse

	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return media.Entry{}, nil, err
	}

	if len(gqlResp.Errors) > 0 {
		var messages []string
		for _, m := range gqlResp.Errors {
			messages = append(messages, m.Message)
		}
		return media.Entry{}, nil, fmt.Errorf("AniList GraphQL: %s", strings.Join(messages, ", "))
	}

	if gqlResp.Data == nil || gqlResp.Data.Media == nil {
		return media.Entry{}, nil, ErrNotFound
	}

	entry := gqlResp.Data.Media.ToEntry()
	relations := gqlResp.Data.Media.GetRelations()

	return entry, relations, nil
}
