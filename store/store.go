package store

import (
	"time"
	"visuilizer/media"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Entry{}, &Relation{}); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) SaveFranchise(entries []media.Entry, relations []media.Relation) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now()

	rows := make([]Entry, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, ToStoreEntry(e, now))
	}

	relRows := make([]Relation, 0, len(relations))
	for _, r := range relations {
		relRows = append(relRows, ToStoreRelation(r))
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"title", "title_english", "format", "episodes", "year", "fetched_at",
			}),
		}).CreateInBatches(rows, 256).Error
		if err != nil {
			return err
		}

		if len(relRows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(relRows, 100).Error
	})
}

func (s *Store) LoadFranchise(seedID int) ([]media.Entry, []media.Relation, error) {
	const franchiseIDs = `
	WITH RECURSIVE franchise(id) AS (
		SELECT ?
		UNION
		SELECT r.to_id FROM relations r JOIN franchise f ON r.from_id = f.id
	)
	SELECT id FROM franchise`

	var ids []int
	if err := s.db.Raw(franchiseIDs, seedID).Scan(&ids).Error; err != nil {
		return nil, nil, err
	}

	if len(ids) == 0 {
		return nil, nil, nil
	}

	var entries []Entry
	if err := s.db.Where("id IN ?", ids).Order("year asc, id asc").Find(&entries).Error; err != nil {
		return nil, nil, err
	}

	var relations []Relation
	if err := s.db.Where("from_id in ? AND to_id in ?", ids, ids).Find(&relations).Error; err != nil {
		return nil, nil, err
	}

	var mediaEntries []media.Entry
	var mediaRelations []media.Relation

	for _, e := range entries {
		mediaEntries = append(mediaEntries, media.Entry{
			ID:           e.ID,
			Title:        e.Title,
			TitleEnglish: e.TitleEnglish,
			Format:       media.Format(e.Format),
			Episodes:     e.Episodes,
			Year:         e.Year,
		})
	}

	for _, r := range relations {
		mediaRelations = append(mediaRelations, media.Relation{
			FromID: r.FromID,
			ToID:   r.ToID,
			Kind:   media.RelationKind(r.Kind),
		})
	}

	return mediaEntries, mediaRelations, nil
}

func (s *Store) LoadEntryByID(id int) (*media.Entry, []media.Relation, error) {
	var e *Entry
	var rels []Relation
	if err := s.db.Where("id = ?", id).First(&e).Error; err != nil {
		return nil, nil, err
	}
	if err := s.db.Where("from_id = ?", id).Find(&rels).Error; err != nil {
		return nil, nil, err
	}

	var mediaRels []media.Relation
	for _, r := range rels {
		mediaRels = append(mediaRels, media.Relation{
			FromID: r.FromID,
			ToID:   r.ToID,
			Kind:   media.RelationKind(r.Kind),
		})
	}

	return &media.Entry{
		ID:           e.ID,
		Title:        e.Title,
		TitleEnglish: e.TitleEnglish,
		Format:       media.Format(e.Format),
		Episodes:     e.Episodes,
		Year:         e.Year,
	}, mediaRels, nil
}

func (s *Store) FranchiseFetchedAt(seedID int) (time.Time, bool) {
	var entry Entry
	err := s.db.Select("fetched_at").First(&entry, seedID).Error
	if err != nil {
		return time.Time{}, false
	}
	return entry.FetchedAt, true
}

func ToStoreEntry(e media.Entry, fetchTime time.Time) Entry {
	return Entry{
		ID:           e.ID,
		Title:        e.Title,
		TitleEnglish: e.TitleEnglish,
		Format:       string(e.Format),
		Episodes:     e.Episodes,
		Year:         e.Year,
		FetchedAt:    fetchTime,
	}
}

func ToStoreRelation(r media.Relation) Relation {
	return Relation{
		FromID: r.FromID,
		ToID:   r.ToID,
		Kind:   string(r.Kind),
	}
}
