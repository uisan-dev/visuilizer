package anilist

import "visuilizer/media"

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
	Format    string `json:"format"`
	Episodes  *int   `json:"episodes"`
	StartDate struct {
		Year *int `json:"year"`
	} `json:"startDate"`
	Relations struct {
		Edges []RelationEdge `json:"edges"`
	} `json:"relations"`
}

func (m *Media) ToEntry() media.Entry {
	title := m.Title.Romaji
	if m.Title.English != nil {
		title = *m.Title.English
	}

	episodes := 0
	if m.Episodes != nil {
		episodes = *m.Episodes
	}

	year := 0
	if m.StartDate.Year != nil {
		year = *m.StartDate.Year
	}

	return media.Entry{
		ID:       m.ID,
		Title:    title,
		Format:   m.Format,
		Episodes: episodes,
		Year:     year,
	}
}

func (m *Media) GetRelations() []media.Relation {
	var relations []media.Relation
	for _, re := range m.Relations.Edges {
		nodeFormat := media.Format(re.Node.Format)
		nodeRelation := media.RelationKind(re.RelationType)
		if nodeFormat.IsAnime() || nodeRelation.FollowForFranchise() {
			relations = append(relations, media.Relation{
				FromID: m.ID,
				ToID:   re.Node.ID,
				Kind:   media.RelationKind(re.RelationType),
			})
		}
	}

	return relations
}

type RelationEdge struct {
	RelationType string `json:"relationType"`
	Node         struct {
		ID     int    `json:"id"`
		Format string `json:"format"`
	} `json:"node"`
}
