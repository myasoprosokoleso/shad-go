package git

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
)

type BlameResult map[string]*commit

type commit struct {
	Author, Committer string
	Lines             int
	Files             map[string]struct{}
}

func ListFiles(repository, revision string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", revision)
	cmd.Dir = repository

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	var files []string
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return files, nil
}

func Blame(repository, revision, file string) (BlameResult, error) {
	cmd := exec.Command("git", "blame", "-p", revision, "--", file)
	cmd.Dir = repository

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	res, err := parseBlame(stdout, file)
	stdout.Close()
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return parseEmptyBlame(res, repository, revision, file)
	}

	return res, nil
}

func parseBlame(stdout io.ReadCloser, file string) (BlameResult, error) {
	reader := bufio.NewReader(stdout)
	commits := make(BlameResult) // map: hash -> *commit

	var currCommit *commit
	for {
		line, err := reader.ReadString('\n')
		switch {
		case err != nil && err != io.EOF:
			return nil, err
		case err == io.EOF && line == "":
			return commits, nil
		case err != io.EOF:
			line = strings.TrimSuffix(line, "\n")
		}

		switch {
		case strings.HasPrefix(line, "author "):
			currCommit.Author = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "committer "):
			currCommit.Committer = strings.TrimPrefix(line, "committer ")
		case strings.HasPrefix(line, "\t"):
			currCommit.Lines++
		case hasHashPrefix(line):
			hash := line[:40]
			c, ok := commits[hash]
			if !ok {
				c = &commit{Files: map[string]struct{}{file: {}}}
				commits[hash] = c
			}
			currCommit = c
		}
	}
}

func hasHashPrefix(s string) bool {
	hash, _, found := strings.Cut(s, " ")
	if found && len(hash) == 40 {
		return true
	}
	return false
}

func parseEmptyBlame(commits BlameResult, repository, revision, file string) (BlameResult, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H%n%an%n%cn", revision, "--", file)
	cmd.Dir = repository

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	hash, author, committer := lines[0], lines[1], lines[2]
	commits[hash] = &commit{
		Author:    author,
		Committer: committer,
		Files:     map[string]struct{}{file: {}},
	}

	return commits, nil
}
