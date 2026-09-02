package store

import "time"

type Entry struct {
	ID           int    `gorm:"primaryKey;autoIncrement:false"`
	Title        string `gorm:"not null;index"`
	TitleEnglish string
	Format       string `gorm:"index"`
	Episodes     int
	Year         int

	FetchedAt time.Time `gorm:"not null"`
}

type Relation struct {
	FromID int    `gorm:"primaryKey;autoIncrement:false"`
	ToID   int    `gorm:"primaryKey;autoIncrement:false"`
	Kind   string `gorm:"primaryKey"`
}

type SavedLayout struct {
	ID        uint   `gorm:"primaryKey"`
	SeedID    int    `gorm:"index;not null"`
	Direction string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SavedNode struct {
	LayoutID uint    `gorm:"primaryKey;autoIncrement:false"`
	EntryID  int     `gorm:"primaryKey;autoIncrement:false"`
	X        float64 `gorm:"not null"`
	Y        float64 `gorm:"not null"`
}

type SaveLayoutRequest struct {
	Direction string `json:"direction" binding:"required,oneof=vertical horizontal"`
	Nodes     []struct {
		ID int     `json:"id" binding:"required"`
		X  float64 `json:"x"`
		Y  float64 `json:"y"`
	} `json:"nodes" binding:"required,min=1,max=500,dive"`
}
