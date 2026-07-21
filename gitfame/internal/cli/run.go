package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"gitlab.com/slon/shad-go/gitfame/internal/filter"
	"gitlab.com/slon/shad-go/gitfame/internal/format"
	"gitlab.com/slon/shad-go/gitfame/internal/git"
	"gitlab.com/slon/shad-go/gitfame/internal/stat"
)

const maxChanBufSize = 1 << 7 // 128

func run(cmd *cobra.Command, args []string) error {
	files, err := git.ListFiles(flags.repository, flags.revision)
	if err != nil {
		return fmt.Errorf("git ls-tree: %w", err)
	}

	bufSize := min(len(files), maxChanBufSize)
	filesCh := make(chan string, bufSize)
	go func() {
		for _, f := range files {
			if filter.MatchFlags(f, flags.extensions, flags.languages, flags.exclude, flags.restrictTo) {
				filesCh <- f
			}
		}
		close(filesCh)
	}()

	var blameEG errgroup.Group
	blameResCh := make(chan git.BlameResult, bufSize)
	jobsCnt := runtime.NumCPU()
	for range jobsCnt {
		blameEG.Go(func() error {
			return processBlameJob(filesCh, flags.repository, flags.revision, blameResCh)
		})
	}

	blameErrCh := make(chan error, 1)
	go func() {
		blameErrCh <- blameEG.Wait()
		close(blameResCh)
	}()

	stats := stat.AggregateBlameRes(blameResCh, flags.useCommitter)
	if blameErr := <-blameErrCh; blameErr != nil {
		return fmt.Errorf("git blame: %w", blameErr)
	}
	outputStats := stat.MakeOutputStats(stats, flags.orderBy)

	if err := format.Print(outputStats, flags.format); err != nil {
		return fmt.Errorf("print %s: %w", flags.format, err)
	}

	return nil
}

func processBlameJob(filesCh <-chan string, repository, revision string, blameResCh chan<- git.BlameResult) error {
	for f := range filesCh {
		res, err := git.Blame(repository, revision, f)
		if err != nil {
			return fmt.Errorf("file %q: %w", f, err)
		}
		blameResCh <- res
	}

	return nil
}
