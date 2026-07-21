package format

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"gitlab.com/slon/shad-go/gitfame/internal/stat"
)

const (
	Tabular   = "tabular"
	CSV       = "csv"
	JSON      = "json"
	JSONLines = "json-lines"
)

func Print(stats []stat.OutputStats, formatFlag string) error {
	switch formatFlag {
	case CSV:
		return printCSV(stats)
	case JSON:
		return printJSON(stats)
	case JSONLines:
		return printJSONLines(stats)
	default:
		// unknown flags are already handled in validateFlags
		return printTabular(stats)
	}
}

func printTabular(stats []stat.OutputStats) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.StripEscape)

	fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")
	esc := []byte{tabwriter.Escape}
	for _, s := range stats {
		fmt.Fprintf(w, "%s%s%s\t%d\t%d\t%d\n", esc, s.Name, esc, s.Lines, s.Commits, s.Files)
	}

	return w.Flush()
}

func printCSV(stats []stat.OutputStats) error {
	records := make([][]string, 0, len(stats))
	records = append(records, []string{"Name", "Lines", "Commits", "Files"})
	for _, s := range stats {
		records = append(records, []string{s.Name, fmt.Sprintf("%d", s.Lines), fmt.Sprintf("%d", s.Commits), fmt.Sprintf("%d", s.Files)})
	}

	w := csv.NewWriter(os.Stdout)
	return w.WriteAll(records)
}

func printJSON(stats []stat.OutputStats) error {
	if err := json.NewEncoder(os.Stdout).Encode(stats); err != nil {
		return err
	}
	return nil
}

func printJSONLines(stats []stat.OutputStats) error {
	enc := json.NewEncoder(os.Stdout)
	for _, s := range stats {
		if err := enc.Encode(s); err != nil {
			return err
		}
	}
	return nil
}
