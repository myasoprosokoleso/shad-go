package stat

import (
	"cmp"
	"slices"

	"gitlab.com/slon/shad-go/gitfame/internal/git"
)

const (
	Lines   = "lines"
	Commits = "commits"
	Files   = "files"
)

type OutputStats struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

type aggStats struct {
	lines               int
	commitHashes, files map[string]struct{}
}

func AggregateBlameRes(blameResCh <-chan git.BlameResult, useCommitter bool) map[string]*aggStats {
	stats := make(map[string]*aggStats) // map: name -> *aggStats
	for res := range blameResCh {
		for hash, commit := range res {
			name := commit.Author
			if useCommitter {
				name = commit.Committer
			}

			if s, ok := stats[name]; !ok {
				stats[name] = &aggStats{
					lines:        commit.Lines,
					commitHashes: map[string]struct{}{hash: {}},
					files:        commit.Files,
				}
			} else {
				s.lines += commit.Lines
				s.commitHashes[hash] = struct{}{}
				for f := range commit.Files {
					s.files[f] = struct{}{}
				}
			}
		}
	}

	return stats
}

func MakeOutputStats(stats map[string]*aggStats, orderBy string) []OutputStats {
	res := make([]OutputStats, 0, len(stats))
	for name, s := range stats {
		res = append(res, OutputStats{
			Name:    name,
			Lines:   s.lines,
			Commits: len(s.commitHashes),
			Files:   len(s.files),
		})
	}

	slices.SortFunc(res, func(a, b OutputStats) int {
		keyA, keyB := getSortKey(a, orderBy), getSortKey(b, orderBy)
		for i := range keyA {
			if n := cmp.Compare(keyB[i], keyA[i]); n != 0 {
				return n
			}
		}
		return cmp.Compare(a.Name, b.Name)
	})

	return res
}

func getSortKey(stats OutputStats, orderBy string) [3]int {
	switch orderBy {
	case Commits:
		return [3]int{stats.Commits, stats.Lines, stats.Files}
	case Files:
		return [3]int{stats.Files, stats.Lines, stats.Commits}
	default:
		// unknown flags are already handled in validateFlags
		return [3]int{stats.Lines, stats.Commits, stats.Files}
	}
}
