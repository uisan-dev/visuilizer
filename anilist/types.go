package anilist

import (
	"visuilizer/debug"
	"visuilizer/media"
)

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type GraphQLResponse struct {
	Data   *MediaData     `json:"data"`
	Errors []GraphQLError `json:"errors"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type MediaData struct {
	Media *Media `json:"Media"`
}

type Media struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string  `json:"romaji"`
		English *string `json:"english"`
	} `json:"title"`
	Format    media.Format `json:"format"`
	Episodes  *int         `json:"episodes"`
	StartDate struct {
		Year *int `json:"year"`
	} `json:"startDate"`
	Relations struct {
		Edges []RelationEdge `json:"edges"`
	} `json:"relations"`
}

func (m *Media) ToEntry() media.Entry {

	episodes := 0
	if m.Episodes != nil {
		episodes = *m.Episodes
	}

	year := 0
	if m.StartDate.Year != nil {
		year = *m.StartDate.Year
	}

	e := media.Entry{
		ID:       m.ID,
		Title:    m.Title.Romaji,
		Format:   m.Format,
		Episodes: episodes,
		Year:     year,
	}

	if m.Title.English != nil && *m.Title.English != "" {
		e.TitleEnglish = *m.Title.English
	}
	return e
}

func (m *Media) GetRelations() []media.Relation {
	var relations []media.Relation
	for _, re := range m.Relations.Edges {
		if !re.Node.Format.IsAnime() || !re.RelationType.FollowForFranchise() {
			continue
		}
		relations = append(relations, media.Relation{
			FromID: m.ID,
			ToID:   re.Node.ID,
			Kind:   re.RelationType,
		})
		debug.Debugf("GetRelations: FromID: %d - ToID: %d - Kind: %s\n", m.ID, re.Node.ID, re.RelationType)

	}

	return relations
}

type RelationEdge struct {
	RelationType media.RelationKind `json:"relationType"`
	Node         struct {
		ID     int          `json:"id"`
		Format media.Format `json:"format"`
	} `json:"node"`
}
