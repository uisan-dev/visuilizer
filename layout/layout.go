package layout

import (
	"cmp"
	"slices"
	"visuilizer/media"
)

func Build(nodes []Node, relations []media.Relation) Graph {
	rows := make(map[int][]Node)
	maxDepth := 0
	for _, n := range nodes {
		rows[n.Depth] = append(rows[n.Depth], n)
		if n.Depth > maxDepth {
			maxDepth = n.Depth
		}
	}

	widest := 0
	for _, row := range rows {
		if len(row) > widest {
			widest = len(row)
		}
	}

	canvasWidth := float64(widest)*ColumnWidth + PadX*2
	canvasHeight := float64(maxDepth+1)*RowHeight + PadY*2

	positions := make(map[int]Position, len(nodes))
	out := make([]Position, 0, len(nodes))

	for depth := 0; depth <= maxDepth; depth++ {
		row := rows[depth]
		if len(row) == 0 {
			continue
		}

		slices.SortFunc(row, func(a, b Node) int {
			if c := cmp.Compare(a.Entry.Year, b.Entry.Year); c != 0 {
				return c
			}
			return cmp.Compare(a.Entry.ID, b.Entry.ID)
		})

		rowWidth := float64(len(row)) * ColumnWidth
		left := (canvasWidth - rowWidth) / 2

		for col, n := range row {
			p := Position{
				Entry:   n.Entry,
				Depth:   n.Depth,
				InCycle: n.InCycle,
				X:       left + float64(col)*ColumnWidth + (ColumnWidth-NodeWidth)/2,
				Y:       PadY + float64(depth)*RowHeight,
			}
			positions[n.Entry.ID] = p
			out = append(out, p)
		}
	}

	edges := make([]Edge, 0, len(relations))
	seen := make(map[[2]int]bool)

	for _, r := range relations {
		if !r.Kind.IsForward() {
			continue
		}

		from, ok := positions[r.FromID]
		if !ok {
			continue
		}

		to, ok := positions[r.ToID]
		if !ok {
			continue
		}

		key := [2]int{r.FromID, r.ToID}
		if seen[key] {
			continue
		}
		seen[key] = true

		edges = append(edges, Edge{
			FromID: r.FromID,
			ToID:   r.ToID,
			Kind:   r.Kind,
			X1:     from.X + NodeWidth/2,
			Y1:     from.Y + NodeHeight,
			X2:     to.X + NodeWidth/2,
			Y2:     to.Y,
		})
	}

	return Graph{
		Nodes:  out,
		Edges:  edges,
		Width:  canvasWidth,
		Height: canvasHeight,
	}
}

func AssignDepths(entries []media.Entry, relations []media.Relation) ([]Node, error) {
	known := make(map[int]bool, len(entries))

	for _, e := range entries {
		known[e.ID] = true
	}

	out := make(map[int][]int)
	inDegree := make(map[int]int, len(entries))

	for _, e := range entries {
		inDegree[e.ID] = 0
	}

	seen := make(map[[2]int]bool)
	for _, r := range relations {
		if !r.Kind.IsForward() || !known[r.FromID] || !known[r.ToID] {
			continue
		}

		if r.FromID == r.ToID {
			continue
		}

		key := [2]int{r.FromID, r.ToID}
		if seen[key] {
			continue
		}

		seen[key] = true

		out[r.FromID] = append(out[r.FromID], r.ToID)
		inDegree[r.ToID]++
	}

	depth := make(map[int]int, len(entries))
	var queue []int

	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}

	processed := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		processed++

		for _, next := range out[id] {
			if depth[id]+1 > depth[next] {
				depth[next] = depth[id] + 1
			}
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	maxDepth := 0
	for _, d := range depth {
		if d > maxDepth {
			maxDepth = d
		}
	}

	nodes := make([]Node, 0, len(entries))

	for _, e := range entries {
		n := Node{Entry: e, Depth: depth[e.ID]}
		if inDegree[e.ID] > 0 {
			n.InCycle = true
			n.Depth = maxDepth + 1
		}
		nodes = append(nodes, n)
	}

	if processed != len(entries) {
		return nodes, ErrCycle
	}

	return nodes, nil
}
