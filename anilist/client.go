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
	"visuilizer/debug"
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
		Query:     MediaQuery,
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

func (c *Client) FetchFranchise(seedID int) ([]media.Entry, []media.Relation, []error) {
	visited := map[int]bool{}
	queue := []int{seedID}

	errs := []error{}

	var entries []media.Entry
	var relations []media.Relation

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if visited[id] {
			continue
		}
		visited[id] = true

		debug.Debugf("Visiting %d\n", id)

		entry, rels, err := c.FetchMedia(id)
		if err != nil {
			debug.Debugf("FetchMedia error: %s\n", err.Error())
			errs = append(errs, err)
		}

		entries = append(entries, entry)

		for _, r := range rels {
			relations = append(relations, r)
			if !visited[r.ToID] {
				queue = append(queue, r.ToID)
				debug.Debugf("Added to queue: %d\n", r.ToID)
			}
		}

		debug.Debugf("Queue length: %d\n", len(queue))
	}

	return entries, relations, nil
}
