// Package rtdapi is a from-scratch client for Read the Docs public REST API v3.
// Endpoints and JSON shapes follow https://docs.readthedocs.com/platform/stable/api/v3.html
// (not BarnabyShearer/readthedocs or any third-party SDK/MCP).
package rtdapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://app.readthedocs.org/api/v3"
	BusinessBaseURL = "https://app.readthedocs.com/api/v3"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func New(baseURL, token string) *Client {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	return &Client{
		BaseURL:    base,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) Do(method, path string, query url.Values, body any) (int, []byte, error) {
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, u, rdr)
	if err != nil {
		return 0, nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Token "+c.Token)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(b))
		if len(msg) > 500 {
			msg = msg[:500] + "…"
		}
		return resp.StatusCode, b, fmt.Errorf("readthedocs %s %s: HTTP %d — %s", method, path, resp.StatusCode, msg)
	}
	return resp.StatusCode, b, nil
}

func (c *Client) JSON(method, path string, query url.Values, body any, out any) error {
	_, raw, err := c.Do(method, path, query, body)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

type Page struct {
	Count    int               `json:"count"`
	Next     *string           `json:"next"`
	Previous *string           `json:"previous"`
	Results  []json.RawMessage `json:"results"`
}

// List walks paginated collection endpoints (?limit=&offset=).
func (c *Client) List(path string, query url.Values, limit int) ([]json.RawMessage, int, error) {
	if query == nil {
		query = url.Values{}
	}
	if limit <= 0 {
		limit = 100
	}
	query.Set("limit", strconv.Itoa(limit))
	var all []json.RawMessage
	total := 0
	offset := 0
	for {
		query.Set("offset", strconv.Itoa(offset))
		var page Page
		if err := c.JSON(http.MethodGet, path, query, nil, &page); err != nil {
			return nil, 0, err
		}
		total = page.Count
		all = append(all, page.Results...)
		if page.Next == nil || len(page.Results) == 0 {
			break
		}
		offset += len(page.Results)
		if offset >= page.Count && page.Count > 0 {
			break
		}
		if len(all) > 10000 {
			break
		}
	}
	return all, total, nil
}

func q(kv map[string]string) url.Values {
	v := url.Values{}
	for k, val := range kv {
		if val != "" {
			v.Set(k, val)
		}
	}
	return v
}

// ---- Projects --------------------------------------------------------------

func (c *Client) ListProjects(filters map[string]string) ([]json.RawMessage, int, error) {
	return c.List("/projects/", q(filters), 50)
}

func (c *Client) GetProject(slug string, expand string) (json.RawMessage, error) {
	query := url.Values{}
	if expand != "" {
		query.Set("expand", expand)
	}
	_, raw, err := c.Do(http.MethodGet, "/projects/"+slug+"/", query, nil)
	return raw, err
}

func (c *Client) CreateProject(body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPost, "/projects/", nil, body)
	return raw, err
}

func (c *Client) UpdateProject(slug string, body map[string]any) error {
	_, _, err := c.Do(http.MethodPatch, "/projects/"+slug+"/", nil, body)
	return err
}

func (c *Client) SyncVersions(slug string) error {
	_, _, err := c.Do(http.MethodPost, "/projects/"+slug+"/sync-versions/", nil, map[string]any{})
	return err
}

func (c *Client) GetSuperproject(slug string) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, "/projects/"+slug+"/superproject/", nil, nil)
	return raw, err
}

// ---- Versions --------------------------------------------------------------

func (c *Client) ListVersions(slug string, filters map[string]string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/versions/", q(filters), 50)
}

func (c *Client) GetVersion(project, version string) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, "/projects/"+project+"/versions/"+version+"/", nil, nil)
	return raw, err
}

func (c *Client) UpdateVersion(project, version string, body map[string]any) error {
	_, _, err := c.Do(http.MethodPatch, "/projects/"+project+"/versions/"+version+"/", nil, body)
	return err
}

// ---- Builds ----------------------------------------------------------------

func (c *Client) ListBuilds(slug string, filters map[string]string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/builds/", q(filters), 50)
}

func (c *Client) GetBuild(slug string, id int, expand string) (json.RawMessage, error) {
	query := url.Values{}
	if expand != "" {
		query.Set("expand", expand)
	}
	_, raw, err := c.Do(http.MethodGet, fmt.Sprintf("/projects/%s/builds/%d/", slug, id), query, nil)
	return raw, err
}

func (c *Client) TriggerBuild(project, version string) (json.RawMessage, error) {
	if version == "" {
		version = "latest"
	}
	_, raw, err := c.Do(http.MethodPost, "/projects/"+project+"/versions/"+version+"/builds/", nil, map[string]any{})
	return raw, err
}

// ---- Subprojects -----------------------------------------------------------

func (c *Client) ListSubprojects(slug string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/subprojects/", nil, 50)
}

func (c *Client) GetSubproject(parent, alias string) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, "/projects/"+parent+"/subprojects/"+alias+"/", nil, nil)
	return raw, err
}

func (c *Client) CreateSubproject(parent string, body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPost, "/projects/"+parent+"/subprojects/", nil, body)
	return raw, err
}

func (c *Client) DeleteSubproject(parent, alias string) error {
	_, _, err := c.Do(http.MethodDelete, "/projects/"+parent+"/subprojects/"+alias+"/", nil, nil)
	return err
}

// ---- Translations ----------------------------------------------------------

func (c *Client) ListTranslations(slug string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/translations/", nil, 50)
}

// ---- Redirects -------------------------------------------------------------

func (c *Client) ListRedirects(slug string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/redirects/", nil, 50)
}

func (c *Client) GetRedirect(slug string, id int) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, fmt.Sprintf("/projects/%s/redirects/%d/", slug, id), nil, nil)
	return raw, err
}

func (c *Client) CreateRedirect(slug string, body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPost, "/projects/"+slug+"/redirects/", nil, body)
	return raw, err
}

func (c *Client) UpdateRedirect(slug string, id int, body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPut, fmt.Sprintf("/projects/%s/redirects/%d/", slug, id), nil, body)
	return raw, err
}

func (c *Client) DeleteRedirect(slug string, id int) error {
	_, _, err := c.Do(http.MethodDelete, fmt.Sprintf("/projects/%s/redirects/%d/", slug, id), nil, nil)
	return err
}

// ---- Environment variables -------------------------------------------------

func (c *Client) ListEnvVars(slug string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/environmentvariables/", nil, 50)
}

func (c *Client) GetEnvVar(slug string, id int) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, fmt.Sprintf("/projects/%s/environmentvariables/%d/", slug, id), nil, nil)
	return raw, err
}

func (c *Client) CreateEnvVar(slug string, body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPost, "/projects/"+slug+"/environmentvariables/", nil, body)
	return raw, err
}

func (c *Client) DeleteEnvVar(slug string, id int) error {
	_, _, err := c.Do(http.MethodDelete, fmt.Sprintf("/projects/%s/environmentvariables/%d/", slug, id), nil, nil)
	return err
}

// ---- Sharing (Business) ----------------------------------------------------

func (c *Client) ListSharing(slug string) ([]json.RawMessage, int, error) {
	return c.List("/projects/"+slug+"/sharing/", nil, 50)
}

func (c *Client) GetSharing(slug string, id int) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, fmt.Sprintf("/projects/%s/sharing/%d/", slug, id), nil, nil)
	return raw, err
}

func (c *Client) CreateSharing(slug string, body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPost, "/projects/"+slug+"/sharing/", nil, body)
	return raw, err
}

func (c *Client) UpdateSharing(slug string, id int, body map[string]any) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodPatch, fmt.Sprintf("/projects/%s/sharing/%d/", slug, id), nil, body)
	return raw, err
}

func (c *Client) DeleteSharing(slug string, id int) error {
	_, _, err := c.Do(http.MethodDelete, fmt.Sprintf("/projects/%s/sharing/%d/", slug, id), nil, nil)
	return err
}

// ---- Organizations (Business) ----------------------------------------------

func (c *Client) ListOrganizations() ([]json.RawMessage, int, error) {
	return c.List("/organizations/", nil, 50)
}

func (c *Client) GetOrganization(slug string) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, "/organizations/"+slug+"/", nil, nil)
	return raw, err
}

func (c *Client) ListOrganizationProjects(slug string) ([]json.RawMessage, int, error) {
	return c.List("/organizations/"+slug+"/projects/", nil, 50)
}

func (c *Client) ListOrganizationTeams(slug, expand string) ([]json.RawMessage, int, error) {
	return c.List("/organizations/"+slug+"/teams/", q(map[string]string{"expand": expand}), 50)
}

// ---- Remote VCS ------------------------------------------------------------

func (c *Client) ListRemoteOrganizations(filters map[string]string) ([]json.RawMessage, int, error) {
	return c.List("/remote/organizations/", q(filters), 50)
}

func (c *Client) ListRemoteRepositories(filters map[string]string) ([]json.RawMessage, int, error) {
	return c.List("/remote/repositories/", q(filters), 50)
}

// ---- Embed -----------------------------------------------------------------

func (c *Client) Embed(params map[string]string) (json.RawMessage, error) {
	_, raw, err := c.Do(http.MethodGet, "/embed/", q(params), nil)
	return raw, err
}

// ExtractInt pulls a numeric id from common JSON keys (id, pk).
func ExtractInt(raw json.RawMessage, keys ...string) int {
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return 0
	}
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			return int(v)
		case json.Number:
			n, _ := v.Int64()
			return int(n)
		case string:
			n, _ := strconv.Atoi(v)
			return n
		}
	}
	return 0
}

func ExtractString(raw json.RawMessage, keys ...string) string {
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return ""
	}
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func NestedString(raw json.RawMessage, path ...string) string {
	var cur any
	if json.Unmarshal(raw, &cur) != nil {
		return ""
	}
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[p]
	}
	s, _ := cur.(string)
	return s
}
