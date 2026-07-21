package rules

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
)

type Rule struct {
	Endpoint               string
	ForbiddenUserAgents    []*regexp.Regexp
	ForbiddenHeaders       []*regexp.Regexp
	RequiredHeaders        []string
	MaxRequestLengthBytes  int64
	MaxResponseLengthBytes int64
	ForbiddenResponseCodes []int
	ForbiddenRequestRe     []*regexp.Regexp
	ForbiddenResponseRe    []*regexp.Regexp
}

type yamlRule struct {
	Endpoint               string   `yaml:"endpoint"`
	ForbiddenUserAgents    []string `yaml:"forbidden_user_agents"`
	ForbiddenHeaders       []string `yaml:"forbidden_headers"`
	RequiredHeaders        []string `yaml:"required_headers"`
	MaxRequestLengthBytes  int64    `yaml:"max_request_length_bytes"`
	MaxResponseLengthBytes int64    `yaml:"max_response_length_bytes"`
	ForbiddenResponseCodes []int    `yaml:"forbidden_response_codes"`
	ForbiddenRequestRe     []string `yaml:"forbidden_request_re"`
	ForbiddenResponseRe    []string `yaml:"forbidden_response_re"`
}

func Load(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config struct {
		Rules []yamlRule `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return compile(config.Rules)
}

func compile(rules []yamlRule) ([]Rule, error) {
	res := make([]Rule, len(rules))
	var err error
	for i, r := range rules {
		if res[i], err = compileRule(r); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func compileRule(rule yamlRule) (Rule, error) {
	res := Rule{
		Endpoint:               rule.Endpoint,
		RequiredHeaders:        rule.RequiredHeaders,
		MaxRequestLengthBytes:  rule.MaxRequestLengthBytes,
		MaxResponseLengthBytes: rule.MaxResponseLengthBytes,
		ForbiddenResponseCodes: rule.ForbiddenResponseCodes,
	}

	var err error
	for _, field := range []struct {
		expr     []string
		compiled *[]*regexp.Regexp
	}{
		{expr: rule.ForbiddenUserAgents, compiled: &res.ForbiddenUserAgents},
		{expr: rule.ForbiddenHeaders, compiled: &res.ForbiddenHeaders},
		{expr: rule.ForbiddenRequestRe, compiled: &res.ForbiddenRequestRe},
		{expr: rule.ForbiddenResponseRe, compiled: &res.ForbiddenResponseRe},
	} {
		if *field.compiled, err = compileExpressions(field.expr); err != nil {
			return res, err
		}
	}

	return res, nil
}

func compileExpressions(expressions []string) ([]*regexp.Regexp, error) {
	res := make([]*regexp.Regexp, len(expressions))
	var err error
	for i, expr := range expressions {
		if res[i], err = regexp.Compile(expr); err != nil {
			return nil, err
		}
	}
	return res, nil
}
