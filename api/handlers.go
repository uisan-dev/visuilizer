package api

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"visuilizer/layout"
	"visuilizer/media"
	"visuilizer/store"

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
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"error": gin.H{"message": "Franchise not imported"}}})
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

func (s *Server) HandleGetGraph(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	entries, relations, err := s.Store.LoadFranchise(id)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load franchise"}})
		return
	}
	if len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "No franchise found for that ID"}})
		return
	}

	nodes, err := layout.AssignDepths(entries, relations)
	hasCycle := errors.Is(err, layout.ErrCycle)
	if err != nil && !hasCycle {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to create franchise layout"}})
		return
	}

	g := layout.Build(nodes, relations)

	resp := layout.GraphResponse{
		SeedID:   id,
		Width:    g.Width,
		Height:   g.Height,
		HasCycle: hasCycle,
		Nodes:    make([]layout.PositionResponse, 0, len(g.Nodes)),
		Edges:    make([]layout.EdgeResponse, 0, len(g.Edges)),
	}

	for _, n := range g.Nodes {
		resp.Nodes = append(resp.Nodes, layout.PositionResponse{
			ID:      n.Entry.ID,
			Title:   n.Entry.Title,
			Format:  string(n.Entry.Format),
			Year:    n.Entry.Year,
			Depth:   n.Depth,
			InCycle: n.InCycle,
			X:       n.X,
			Y:       n.Y,
		})
	}

	for _, e := range g.Edges {
		resp.Edges = append(resp.Edges, layout.EdgeResponse{
			From: e.FromID,
			To:   e.ToID,
			Kind: string(e.Kind),
			X1:   e.X1,
			Y1:   e.Y1,
			X2:   e.X2,
			Y2:   e.Y2,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (s *Server) HandleSaveLayout(c *gin.Context) {
	seedID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	var req store.SaveLayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "Invalid layout payload"}})
		return
	}

	entries, _, err := s.Store.LoadFranchise(seedID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load franchise"}})
		return
	}

	if len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"error": gin.H{"message": "Franchise not imported"}}})
	}

	valid := make(map[int]bool, len(entries))
	for _, e := range entries {
		valid[e.ID] = true
	}

	positions := make([]store.SavedNode, 0, len(req.Nodes))
	seen := make(map[int]bool, len(req.Nodes))

	for _, n := range req.Nodes {
		if !valid[n.ID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": fmt.Sprintf("Entry %d is not part of this franchise", n.ID)}})
			return
		}

		if seen[n.ID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": fmt.Sprintf("Entry %d appears multiple times", n.ID)}})
			return
		}

		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": fmt.Sprintf("Entry %d has invalid coordinates", n.ID)}})
			return
		}

		seen[n.ID] = true
		positions = append(positions, store.SavedNode{EntryID: n.ID, X: n.X, Y: n.Y})
	}

	layout, err := s.Store.SaveLayout(seedID, req.Direction, positions)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to save layout"}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": layout.ID}})
}

func (s *Server) HandleGetLayout(c *gin.Context) {
	layoutID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "Layout ID must be a number"}})
		return
	}

	saved, positions, err := s.Store.LoadLayout(uint(layoutID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "No layout with that ID"}})
		return
	}
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load layout"}})
		return
	}

	entries, relations, err := s.Store.LoadFranchise(saved.SeedID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load franchise"}})
		return
	}

	g := layout.GraphFromSaved(entries, relations, positions, saved.Direction)

	resp := layout.LayoutResponse{
		ID:        saved.ID,
		SeedID:    saved.SeedID,
		Direction: saved.Direction,
		CreatedAt: saved.CreatedAt,
		Nodes:     make([]layout.LayoutNodeResponse, 0, len(g.Nodes)),
		Edges:     make([]layout.EdgeResponse, 0, len(g.Edges)),
	}

	for _, n := range g.Nodes {
		resp.Nodes = append(resp.Nodes, layout.LayoutNodeResponse{
			ID:     n.Entry.ID,
			Title:  n.Entry.Title,
			Format: string(n.Entry.Format),
			Year:   n.Entry.Year,
			X:      n.X,
			Y:      n.Y,
		})
	}
	for _, e := range g.Edges {
		resp.Edges = append(resp.Edges, layout.EdgeResponse{
			From: e.FromID, To: e.ToID, Kind: string(e.Kind),
			X1: e.X1, Y1: e.Y1, X2: e.X2, Y2: e.Y2,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}
func (s *Server) HandleGetLayoutSVG(c *gin.Context) {
	layoutID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "Layout ID must be a number"}})
		return
	}

	saved, positions, err := s.Store.LoadLayout(uint(layoutID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "No layout with that ID"}})
		return
	}
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load layout"}})
		return
	}

	entries, relations, err := s.Store.LoadFranchise(saved.SeedID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load franchise"}})
		return
	}

	g := layout.GraphFromSaved(entries, relations, positions, saved.Direction)
	c.Data(http.StatusOK, "image/svg+xml", layout.Render(g, saved.Direction))
}

func (s *Server) HandleListLayouts(c *gin.Context) {
	seedID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "ID must be a number"}})
		return
	}

	layouts, err := s.Store.LoadLayoutsForFranchise(seedID)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "Unable to load layouts"}})
		return
	}

	resp := make([]layout.LayoutSummaryResponse, 0, len(layouts))
	for _, l := range layouts {
		resp = append(resp, layout.LayoutSummaryResponse{
			ID:        l.ID,
			SeedID:    l.SeedID,
			Direction: l.Direction,
			CreatedAt: l.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
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
