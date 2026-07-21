//go:build !solution

package httpgauge

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

type Gauge struct {
	stats *concurrentMap
}

func New() *Gauge {
	return &Gauge{stats: newConcurrentMap()}
}

func (g *Gauge) Snapshot() map[string]int {
	return g.stats.GetCopy()
}

// ServeHTTP returns accumulated statistics in text format ordered by pattern.
//
// For example:
//
//	/a 10
//	/b 5
//	/c/{id} 7
func (g *Gauge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snap := g.Snapshot()

	type entry struct {
		pattern string
		count   int
	}
	entries := make([]entry, 0, len(snap))
	for pattern, count := range snap {
		entries = append(entries, entry{pattern: pattern, count: count})
	}

	slices.SortFunc(entries, func(a, b entry) int {
		return cmp.Compare(a.pattern, b.pattern)
	})
	for _, stat := range entries {
		fmt.Fprintln(w, stat.pattern, stat.count)
	}
}

func (g *Gauge) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// RoutePattern() can change the pattern before the handler is called
			pattern := chi.RouteContext(r.Context()).RoutePattern()
			g.stats.Inc(pattern)
		}()

		next.ServeHTTP(w, r)
	})
}
