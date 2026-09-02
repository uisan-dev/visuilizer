package media

type Format string

const (
	TV          Format = "TV"
	TVShort     Format = "TV_SHORT"
	Movie       Format = "MOVIE"
	Special     Format = "SPECIAL"
	OVA         Format = "OVA"
	ONA         Format = "ONA"
	OtherFormat Format = "OTHER"
)

type RelationKind string

const (
	Prequel       RelationKind = "PREQUEL"
	Sequel        RelationKind = "SEQUEL"
	SideStory     RelationKind = "SIDE_STORY"
	Summary       RelationKind = "SUMMARY"
	Alternative   RelationKind = "ALTERNATIVE"
	Parent        RelationKind = "PARENT"
	SpinOff       RelationKind = "SPIN_OFF"
	OtherRelation RelationKind = "OTHER"
)

func (rk RelationKind) FollowForFranchise() bool {
	switch rk {
	case Prequel, Sequel, SideStory, Summary, Alternative, Parent, SpinOff:
		return true
	default:
		return false
	}
}

func (rk RelationKind) IsForward() bool {
	switch rk {
	case Sequel, SideStory, Summary:
		return true
	default:
		return false
	}
}

type Entry struct {
	ID           int
	Title        string
	TitleEnglish string
	Format       Format
	Episodes     int
	Year         int
}

func (f Format) IsAnime() bool {
	switch f {
	case TV, TVShort, Movie, Special, OVA, ONA:
		return true
	default:
		return false
	}
}

type Relation struct {
	FromID int
	ToID   int
	Kind   RelationKind
}
