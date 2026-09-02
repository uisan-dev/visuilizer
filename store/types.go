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
