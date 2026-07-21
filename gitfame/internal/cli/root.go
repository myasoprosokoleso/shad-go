package cli

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	"gitlab.com/slon/shad-go/gitfame/internal/format"
	"gitlab.com/slon/shad-go/gitfame/internal/stat"
)

var (
	Root = &cobra.Command{
		Use:     "gitfame",
		Short:   "Count lines of code, commits and files per author or committer in a git repository",
		PreRunE: validateFlags,
		RunE:    run,
	}
	flags = &struct {
		repository            string
		revision              string
		orderBy               string
		useCommitter          bool
		format                string
		extensions, languages []string
		exclude, restrictTo   []string
	}{}
)

func init() {
	f := Root.Flags()

	f.StringVar(&flags.repository, "repository", ".", "Path to git repository")
	f.StringVar(&flags.revision, "revision", "HEAD", "Pointer to commit")
	f.StringVar(&flags.orderBy, "order-by", "lines", "Sorting key: {lines|commits|files}")
	f.BoolVar(&flags.useCommitter, "use-committer", false, "Use committer instead of author")
	f.StringVar(&flags.format, "format", "tabular", "Output format: {tabular|csv|json|json-lines}")
	f.StringSliceVar(&flags.extensions, "extensions", nil, "File extension filter, e.g., .go,.md")
	f.StringSliceVar(&flags.languages, "languages", nil, "Language filter, e.g., go,markdown")
	f.StringSliceVar(&flags.exclude, "exclude", nil, "Glob patterns for excluding files, e.g., foo/*,bar/*")
	f.StringSliceVar(&flags.restrictTo, "restrict-to", nil, "Glob patterns for including files")

	Root.SilenceUsage = true
}

func validateFlags(cmd *cobra.Command, args []string) error {
	switch flags.orderBy {
	case stat.Lines, stat.Commits, stat.Files:
	default:
		return fmt.Errorf("invalid --order-by: %s (want {%s|%s|%s})", flags.orderBy, stat.Lines, stat.Commits, stat.Files)
	}

	switch flags.format {
	case format.Tabular, format.CSV, format.JSON, format.JSONLines:
	default:
		return fmt.Errorf("invalid --format: %s (want {%s|%s|%s|%s})", flags.format, format.Tabular, format.CSV, format.JSON, format.JSONLines)
	}

	patterns := slices.Concat(flags.exclude, flags.restrictTo)
	for _, pattern := range patterns {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
	}

	return nil
}
