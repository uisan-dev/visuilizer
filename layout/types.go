package layout

import (
	"time"
	"visuilizer/config"
	"visuilizer/media"
)

type Node struct {
	Entry   media.Entry
	Depth   int
	InCycle bool
}

const (
	ColumnWidth = config.LayoutColumnWidth
	RowHeight   = config.LayoutRowHeight
	NodeWidth   = config.LayoutNodeWidth
	NodeHeight  = config.LayoutNodeHeight
	PadX        = config.LayoutPadX
	PadY        = config.LayoutPadY
)

type Position struct {
	Entry   media.Entry
	Depth   int
	InCycle bool
	X, Y    float64
}

type Edge struct {
	FromID int
	ToID   int
	Kind   media.RelationKind
	X1, Y1 float64
	X2, Y2 float64
}

type Graph struct {
	Nodes  []Position
	Edges  []Edge
	Width  float64
	Height float64
}

type PositionResponse struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Format  string  `json:"format"`
	Year    int     `json:"year"`
	Depth   int     `json:"depth"`
	InCycle bool    `json:"in_cycle"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
}

type EdgeResponse struct {
	From int     `json:"from"`
	To   int     `json:"to"`
	Kind string  `json:"kind"`
	X1   float64 `json:"x1"`
	Y1   float64 `json:"y1"`
	X2   float64 `json:"x2"`
	Y2   float64 `json:"y2"`
}

type GraphResponse struct {
	SeedID   int                `json:"seed_id"`
	Width    float64            `json:"width"`
	Height   float64            `json:"height"`
	HasCycle bool               `json:"has_cycle"`
	Nodes    []PositionResponse `json:"nodes"`
	Edges    []EdgeResponse     `json:"edges"`
}

type LayoutNodeResponse struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Format string  `json:"format"`
	Year   int     `json:"year"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type LayoutResponse struct {
	ID        uint                 `json:"id"`
	SeedID    int                  `json:"seed_id"`
	Direction string               `json:"direction"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	Nodes     []LayoutNodeResponse `json:"nodes"`
	Edges     []EdgeResponse       `json:"edges"`
}

type LayoutSummaryResponse struct {
	ID        uint      `json:"id"`
	SeedID    int       `json:"seed_id"`
	Direction string    `json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}
