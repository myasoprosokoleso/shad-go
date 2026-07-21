package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"

	"gitlab.com/slon/shad-go/firewall/internal/rules"
)

type Transport struct {
	base  http.RoundTripper
	rules []rules.Rule
}

func NewTransport(rules []rules.Rule) *Transport {
	return &Transport{base: http.DefaultTransport, rules: rules}
}

// RoundTrip must return err == nil if it obtained a response,
// regardless of the response's HTTP status code.
// A non-nil err should be reserved for failure to obtain a response.
//
// RoundTrip should not modify the request, except for
// consuming and closing the Request's Body.
// RoundTrip must always close the body, including on errors.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	rule := t.getRule(req.URL.Path)
	if rule == nil {
		return t.base.RoundTrip(req)
	}

	if ok, err := checkRequest(rule, req); !ok {
		// req.Body may be nil, for example, for HEAD requests
		if req.Body != nil {
			req.Body.Close()
		}
		return makeForbiddenResponse(), err
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if ok, err := checkResponse(rule, resp); !ok {
		resp.Body.Close()
		return makeForbiddenResponse(), err
	}

	return resp, nil
}

func checkRequest(rule *rules.Rule, req *http.Request) (bool, error) {
	ua := req.Header.Get("User-Agent")
	for _, re := range rule.ForbiddenUserAgents {
		if re.MatchString(ua) {
			return false, nil
		}
	}

	for k, values := range req.Header {
		for _, v := range values {
			header := fmt.Sprintf("%s: %s", k, v)
			for _, re := range rule.ForbiddenHeaders {
				if re.MatchString(header) {
					return false, nil
				}
			}
		}
	}

	for _, rh := range rule.RequiredHeaders {
		if req.Header.Get(rh) == "" {
			return false, nil
		}
	}

	if rule.MaxRequestLengthBytes > 0 && req.ContentLength > rule.MaxRequestLengthBytes {
		return false, nil
	}

	if len(rule.ForbiddenRequestRe) > 0 && req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return false, err
		}

		for _, re := range rule.ForbiddenRequestRe {
			if re.Match(body) {
				return false, nil
			}
		}

		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	return true, nil
}

func checkResponse(rule *rules.Rule, resp *http.Response) (bool, error) {
	if rule.MaxResponseLengthBytes > 0 && resp.ContentLength > rule.MaxResponseLengthBytes {
		return false, nil
	}

	if slices.Contains(rule.ForbiddenResponseCodes, resp.StatusCode) {
		return false, nil
	}

	if len(rule.ForbiddenResponseRe) > 0 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		for _, re := range rule.ForbiddenResponseRe {
			if re.Match(body) {
				return false, nil
			}
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	return true, nil
}

func (t *Transport) getRule(endpoint string) *rules.Rule {
	for i, r := range t.rules {
		if r.Endpoint == endpoint {
			return &t.rules[i]
		}
	}
	return nil
}

func makeForbiddenResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(bytes.NewBuffer([]byte("Forbidden"))),
		Header:     make(http.Header),
	}
}
