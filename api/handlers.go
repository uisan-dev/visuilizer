package api

import (
	"errors"
	"net/http"
	"strconv"
	"visuilizer/media"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) HandleGetMedia(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	entry, relations, err := s.Store.LoadEntryByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "No media with that ID"}})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"message": "Unable to fetch media"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": MediaResponse{
		Entry:     ToEntryResponse(*entry),
		Relations: ToRelationResponses(relations),
	}})
}

func (s *Server) HandleGetFranchise(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	entries, relations, err := s.Store.LoadFranchise(id)

	if err != nil {
		c.Error(err)
		return
	}

	if len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "No franchise found for that ID"}})
		return
	}

	resp := FranchiseResponse{
		SeedID:  id,
		Entries: make([]EntryResponse, 0, len(entries)),
	}

	for _, e := range entries {
		resp.Entries = append(resp.Entries, ToEntryResponse(e))
	}

	resp.Relations = ToRelationResponses(relations)

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (s *Server) HandleImport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	job, started := s.Importer.Start(id)

	status := http.StatusAccepted
	if !started {
		status = http.StatusOK
	}

	c.JSON(status, gin.H{"data": job})
}

func (s *Server) HandleImportStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	job, ok := s.Importer.Status(id)

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "No import job for that ID"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": job})
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
