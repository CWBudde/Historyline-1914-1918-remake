package dosai

import (
	"container/heap"

	"github.com/CWBudde/Historyline-1914-1918-remake/internal/hexgrid"
)

// dijkstra is the least route cost from every hex of costs to a start hex:
// a step costs the entry cost of the hex it enters, a start hex costs its
// start value. Hexes missing from costs are closed.
func dijkstra(costs map[hexgrid.Coord]int, start map[hexgrid.Coord]int) map[hexgrid.Coord]int {
	dist := make(map[hexgrid.Coord]int, len(costs))
	q := &queue{}
	for c, v := range start {
		dist[c] = v
		heap.Push(q, item{c, v})
	}
	for q.Len() > 0 {
		it := heap.Pop(q).(item)
		if it.d > dist[it.c] {
			continue
		}
		// Reaching it.c from a neighbour costs its entry cost.
		d := it.d + costs[it.c]
		for _, n := range hexgrid.Around(it.c) {
			if _, ok := costs[n]; !ok {
				continue
			}
			if old, ok := dist[n]; !ok || d < old {
				dist[n] = d
				heap.Push(q, item{n, d})
			}
		}
	}
	return dist
}

type item struct {
	c hexgrid.Coord
	d int
}

type queue []item

func (q queue) Len() int           { return len(q) }
func (q queue) Less(i, j int) bool { return q[i].d < q[j].d }
func (q queue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *queue) Push(x any)        { *q = append(*q, x.(item)) }
func (q *queue) Pop() any {
	old := *q
	it := old[len(old)-1]
	*q = old[:len(old)-1]
	return it
}
