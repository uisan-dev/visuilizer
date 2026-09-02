package layout

import (
	"fmt"
	"html"
	"math"
	"strings"
	"visuilizer/media"
	"visuilizer/store"
)

func Render(g Graph, direction string) []byte {
	var b strings.Builder

	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" font-family="system-ui, sans-serif">`,
		g.Width, g.Height, g.Width, g.Height)

	b.WriteString(`<defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto"><path d="M0,0 L10,5 L0,10 z" fill="#999"/></marker></defs>`)
	fmt.Fprintf(&b, `<rect width="%.0f" height="%.0f" fill="#fdfdfc"/>`, g.Width, g.Height)

	for _, e := range g.Edges {
		var c1x, c1y, c2x, c2y float64
		if direction == "horizontal" {
			dx := (e.X2 - e.X1) / 2
			c1x, c1y = e.X1+dx, e.Y1
			c2x, c2y = e.X2-dx, e.Y2
		} else {
			dy := (e.Y2 - e.Y1) / 2
			c1x, c1y = e.X1, e.Y1+dy
			c2x, c2y = e.X2, e.Y2-dy
		}

		fmt.Fprintf(&b,
			`<path d="M%.1f,%.1f C%.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="none" stroke="%s" stroke-width="1.5" stroke-dasharray="%s" marker-end="url(#arrow)"/>`,
			e.X1, e.Y1,
			c1x, c1y,
			c2x, c2y,
			e.X2, e.Y2,
			edgeColor(e.Kind), edgeDash(e.Kind))
	}

	for _, n := range g.Nodes {
		fill, stroke := nodeColors(n)

		fmt.Fprintf(&b,
			`<rect x="%.1f" y="%.1f" width="%.0f" height="%.0f" rx="6" fill="%s" stroke="%s" stroke-width="1"/>`,
			n.X, n.Y, NodeWidth, NodeHeight, fill, stroke)

		fmt.Fprintf(&b,
			`<text x="%.1f" y="%.1f" font-size="11" fill="#1a1a19">%s</text>`,
			n.X+10, n.Y+19, html.EscapeString(truncate(n.Entry.Title, 32)))

		meta := fmt.Sprintf("%d · %s", n.Entry.Year, n.Entry.Format)
		fmt.Fprintf(&b,
			`<text x="%.1f" y="%.1f" font-size="10" fill="#73726c">%s</text>`,
			n.X+10, n.Y+34, html.EscapeString(meta))

		fmt.Fprintf(&b, `<title>%s</title>`, html.EscapeString(n.Entry.Title))
	}

	b.WriteString(`</svg>`)
	return []byte(b.String())
}
func GraphFromSaved(entries []media.Entry, relations []media.Relation, positions []store.SavedNode, direction string) Graph {
	byID := make(map[int]media.Entry, len(entries))
	for _, e := range entries {
		byID[e.ID] = e
	}

	nodes := make(map[int]Position, len(positions))
	out := make([]Position, 0, len(positions))

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for _, p := range positions {
		entry, ok := byID[p.EntryID]
		if !ok {
			continue
		}

		pos := Position{Entry: entry, X: p.X, Y: p.Y}
		nodes[p.EntryID] = pos
		out = append(out, pos)

		minX = math.Min(minX, p.X)
		minY = math.Min(minY, p.Y)
		maxX = math.Max(maxX, p.X+NodeWidth)
		maxY = math.Max(maxY, p.Y+NodeHeight)
	}

	if len(out) == 0 {
		return Graph{Width: PadX * 2, Height: PadY * 2}
	}

	// Shift so the top-left node sits at the padding offset.
	dx := PadX - minX
	dy := PadY - minY
	for i := range out {
		out[i].X += dx
		out[i].Y += dy
	}
	for id, p := range nodes {
		p.X += dx
		p.Y += dy
		nodes[id] = p
	}

	edges := make([]Edge, 0, len(relations))
	seen := make(map[[2]int]bool)

	for _, r := range relations {
		if !r.Kind.IsForward() {
			continue
		}

		from, ok := nodes[r.FromID]
		if !ok {
			continue
		}
		to, ok := nodes[r.ToID]
		if !ok {
			continue
		}

		key := [2]int{r.FromID, r.ToID}
		if seen[key] {
			continue
		}
		seen[key] = true

		e := Edge{FromID: r.FromID, ToID: r.ToID, Kind: r.Kind}
		if direction == "horizontal" {
			e.X1 = from.X + NodeWidth
			e.Y1 = from.Y + NodeHeight/2
			e.X2 = to.X
			e.Y2 = to.Y + NodeHeight/2
		} else {
			e.X1 = from.X + NodeWidth/2
			e.Y1 = from.Y + NodeHeight
			e.X2 = to.X + NodeWidth/2
			e.Y2 = to.Y
		}
		edges = append(edges, e)
	}

	return Graph{
		Nodes:  out,
		Edges:  edges,
		Width:  maxX - minX + PadX*2,
		Height: maxY - minY + PadY*2,
	}
}

func nodeColors(n Position) (fill, stroke string) {
	if n.InCycle {
		return "#fcebeb", "#e24b4a"
	}

	switch n.Entry.Format {
	case media.TV, media.TVShort:
		return "#e6f1fb", "#378add"
	case media.Movie:
		return "#eeedfe", "#7f77dd"
	case media.OVA, media.ONA:
		return "#e1f5ee", "#1d9e75"
	default:
		return "#f1efe8", "#888780"
	}
}

func edgeColor(kind media.RelationKind) string {
	if kind == media.Sequel {
		return "#888780"
	}
	return "#b4b2a9"
}

func edgeDash(kind media.RelationKind) string {
	if kind == media.Sequel {
		return "none"
	}
	return "4 3"
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
