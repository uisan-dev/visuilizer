package api

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

type MediaResponse struct {
	Entry     EntryResponse      `json:"entry"`
	Relations []RelationResponse `json:"relations"`
}

type FranchiseResponse struct {
	SeedID    int                `json:"seed_id"`
	Entries   []EntryResponse    `json:"entries"`
	Relations []RelationResponse `json:"relations"`
}
