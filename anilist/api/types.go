package api

import (
	"errors"
	"net/http"
	"strconv"
	"visuilizer/anilist"
	"visuilizer/media"

	"github.com/gin-gonic/gin"
)

type EntryResponse struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Format   string `json:"format"`
	Episodes int    `json:"episodes"`
	Year     int    `json:"year"`
}

type RelationResponse struct {
	From int    `json:"from"`
	To   int    `json:"to"`
	Kind string `json:"kind"`
}

type FranchiseResponse struct {
	SeedID    int                `json:"seed_id"`
	Entries   []EntryResponse    `json:"entries"`
	Relations []RelationResponse `json:"relations"`
}

func (s *Server) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) HandleGetMedia(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID must be a number"})
		return
	}

	entry, _, err := s.Client.FetchMedia(id)
	if errors.Is(err, anilist.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "No media with that ID"})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch media"})
	}

	c.JSON(http.StatusOK, ToEntryResponse(entry))
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

	for _, r := range relations {
		resp.Relations = append(resp.Relations, RelationResponse{
			From: r.FromID,
			To:   r.ToID,
			Kind: string(r.Kind),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func ToEntryResponse(e media.Entry) EntryResponse {
	return EntryResponse{
		ID:       e.ID,
		Title:    e.Title,
		Format:   string(e.Format),
		Episodes: e.Episodes,
		Year:     e.Year,
	}
}
