package api

import (
	"errors"
	"net/http"
	"strconv"
	"visuilizer/anilist"
	"visuilizer/media"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) HandleGetMedia(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}

	entry, relations, err := s.Client.FetchMedia(id)
	if errors.Is(err, anilist.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "No media with that ID"})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to fetch media"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": MediaResponse{
		Entry:     ToEntryResponse(entry),
		Relations: ToRelationResponses(relations),
	}}
)
}

func (s *Server) HandleGetFranchise(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}

	entries, relations, errs := s.Client.FetchFranchise(id)

	if len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No franchise found for that ID"})
		return
	}

	for _, e := range errs {
		c.Error(e)
	}

	resp := FranchiseResponse{
		SeedID:    id,
		Entries:   make([]EntryResponse, 0, len(entries)),
		Relations: make([]RelationResponse, 0, len(relations)),
	}

	for _, e := range entries {
		resp.Entries = append(resp.Entries, ToEntryResponse(e))
	}

	resp.Relations = ToRelationResponses(relations)

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func ToEntryResponse(e media.Entry) EntryResponse {
	return EntryResponse{
		ID:        e.ID,
		Title:     e.Title,
		Relations: nil,
		Format:    string(e.Format),
		Episodes:  e.Episodes,
		Year:      e.Year,
	}
}

func ToRelationResponses(rels []media.Relation) []RelationResponse {
	out := make([]RelationResponse, 0, len(rels))
	for _, r := range rels {
		out = append(out, RelationResponse{
			From: r.FromID,
			To:   r.ToID,
			Kind: string(r.Kind),
		})
	}
	return out
}
