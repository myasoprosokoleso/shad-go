package ciletters

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed letter.tmpl
var letterTemplateText string

func jobSection(job Job) string {
	logLines := strings.Split(job.RunnerLog, "\n")
	log := strings.Join(logLines[max(0, len(logLines)-10):], "\n")

	var builder strings.Builder

	builder.WriteString("\n        Stage: ")
	builder.WriteString(job.Stage)
	builder.WriteString(", Job ")
	builder.WriteString(job.Name)
	builder.WriteString("\n            ")
	builder.WriteString(strings.ReplaceAll(log, "\n", "\n            "))
	builder.WriteByte('\n')

	return builder.String()
}

func MakeLetter(n *Notification) (string, error) {
	var letterTemplate = template.Must(template.New("letter").Funcs(template.FuncMap{
		"shortHash":  func(hash string) string { return hash[:8] },
		"jobSection": jobSection,
	}).Parse(letterTemplateText))

	var buffer bytes.Buffer
	if err := letterTemplate.Execute(&buffer, n); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
