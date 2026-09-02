package importer

import (
	"fmt"
	"log"
	"sync"
	"time"
	"visuilizer/anilist"
	"visuilizer/config"
	"visuilizer/store"
)

type Importer struct {
	client *anilist.Client
	store  *store.Store

	mu   sync.Mutex
	jobs map[int]*Job
}

func NewImporter(client *anilist.Client, st *store.Store) *Importer {
	return &Importer{
		client: client,
		store:  st,
		jobs:   make(map[int]*Job),
	}
}

func (i *Importer) Start(seedID int) (Job, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	var existing *Job
	var ok bool

	if existing, ok = i.jobs[seedID]; ok {
		if existing.Status == StatusRunning {
			return *existing, false
		}
		if existing.Status == StatusDone && time.Since(*existing.EndedAt) < config.MaxJobInMemoryPeriod {
			return *existing, false
		}
	}
	if fetchedAt, ok := i.store.FranchiseFetchedAt(seedID); ok && time.Since(fetchedAt) < config.MaxJobInMemoryPeriod {
		return Job{
			SeedID:  seedID,
			Status:  StatusDone,
			EndedAt: &fetchedAt,
		}, false
	}

	job := &Job{
		SeedID:    seedID,
		Status:    StatusRunning,
		StartedAt: time.Now(),
	}

	i.jobs[seedID] = job

	go i.run(seedID)

	return *job, true
}

func (i *Importer) Status(seedID int) (Job, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	job, ok := i.jobs[seedID]

	if !ok {
		return Job{}, false
	}

	return *job, true
}

func (i *Importer) run(seedID int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("importing %d (panic): %v\n", seedID, r)
			i.finish(seedID, 0, fmt.Errorf("internal error"))
		}
	}()

	entries, relations, errs := i.client.FetchFranchise(seedID)
	for _, e := range errs {
		log.Printf("importing %d: %v\n", seedID, e)
	}

	if len(entries) == 0 {
		i.finish(seedID, 0, fmt.Errorf("no entries found for %d", seedID))
		return
	}

	if err := i.store.SaveFranchise(entries, relations); err != nil {
		i.finish(seedID, 0, err)
		return
	}

	i.finish(seedID, len(entries), nil)
}

func (i *Importer) finish(seedID, count int, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	job, ok := i.jobs[seedID]
	if !ok {
		return
	}

	now := time.Now()
	job.EndedAt = &now
	job.Entries = count

	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()
		return
	}
	job.Status = StatusDone
}
