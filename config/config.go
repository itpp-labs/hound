package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultMsBetweenPoll          = 30000
	defaultMaxConcurrentIndexers  = 2
	defaultMaxConcurrentSearchers = 1000
	defaultMaxReposInFirstResult  = 10
	defaultMaxReposInNextResult   = 30
	defaultPushEnabled            = false
	defaultPollEnabled            = true
	defaultTitle                  = "Hound"
	defaultVcs                    = "git"
	defaultBaseUrl                = "{url}/blob/{rev}/{path}{anchor}"
	defaultAnchor                 = "#L{line}"
	defaultHealthCheckURI         = "/healthz"
)

type UrlPattern struct {
	BaseUrl string `json:"base-url"`
	Anchor  string `json:"anchor"`
}

type PatternLink struct {
	Pattern string `json:"pattern"`
	Link    string `json:"link"`
}

type Repo struct {
	Name              string         `json:"name"`
	Url               string         `json:"url"`
	MsBetweenPolls    int            `json:"ms-between-poll"`
	Vcs               string         `json:"vcs"`
	VcsConfigMessage  *SecretMessage `json:"vcs-config"`
	UrlPattern        *UrlPattern    `json:"url-pattern"`
	ExcludeDotFiles   bool           `json:"exclude-dot-files"`
	EnablePollUpdates *bool          `json:"enable-poll-updates"`
	EnablePushUpdates *bool          `json:"enable-push-updates"`
	PatternLinks      []PatternLink  `json:"pattern-links"`
}

// Used for interpreting the config value for fields that use *bool. If a value
// is present, that value is returned. Otherwise, the default is returned.
func optionToBool(val *bool, def bool) bool {
	if val == nil {
		return def
	}
	return *val
}

// Are polling based updates enabled on this repo?
func (r *Repo) PollUpdatesEnabled() bool {
	return optionToBool(r.EnablePollUpdates, defaultPollEnabled)
}

// Are push based updates enabled on this repo?
func (r *Repo) PushUpdatesEnabled() bool {
	return optionToBool(r.EnablePushUpdates, defaultPushEnabled)
}

type Config struct {
	DbPath                 string            `json:"dbpath"`
	Title                  string            `json:"title"`
	Favicon                *Favicon          `json:"favicon"`
	Repos                  []*Repo           `json:"repos"`
	MaxConcurrentIndexers  int               `json:"max-concurrent-indexers"`
	MaxConcurrentSearchers int               `json:"max-concurrent-searchers"`
	MaxReposInFirstResult  int               `json:"max-repos-in-first-result"`
	MaxReposInNextResult   int               `json:"max-repos-in-next-result"`
	HealthCheckURI         string            `json:"health-check-uri"`
	InitSearch             map[string]string `json:"init-search"`
	Ads                    []*string         `json:"ads"`
}

type Favicon struct {
	Image   []byte
	ModTime time.Time
}

func (f *Favicon) UnmarshalJSON(b []byte) error {
	if b == nil {
		return errors.New("Favicon: UnmarshalJSON on nil pointer")
	}
	unquoted := string(b[1 : len(b)-2])
	data, err := base64.RawStdEncoding.DecodeString(unquoted)
	if err != nil {
		panic(err)
	}
	f.Image = append(((*f).Image)[0:0], data...)
	f.ModTime = time.Now()
	return nil
}

type ClientConfig struct {
	InitSearch map[string]string
}

// SecretMessage is just like json.RawMessage but it will not
// marshal its value as JSON. This is to ensure that vcs-config
// is not marshalled into JSON and send to the UI.
type SecretMessage []byte

// This always marshals to an empty object.
func (s *SecretMessage) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}

// See http://golang.org/pkg/encoding/json/#RawMessage.UnmarshalJSON
func (s *SecretMessage) UnmarshalJSON(b []byte) error {
	if b == nil {
		return errors.New("SecretMessage: UnmarshalJSON on nil pointer")
	}
	*s = append((*s)[0:0], b...)
	return nil
}

// Get the JSON encode vcs-config for this repo. This returns nil if
// the repo doesn't declare a vcs-config.
func (r *Repo) VcsConfig() []byte {
	if r.VcsConfigMessage == nil {
		return nil
	}
	return *r.VcsConfigMessage
}

// Populate missing config values with default values.
func initRepo(r *Repo) {
	if r.MsBetweenPolls == 0 {
		r.MsBetweenPolls = defaultMsBetweenPoll
	}

	if r.Vcs == "" {
		r.Vcs = defaultVcs
	}

	if r.UrlPattern == nil {
		r.UrlPattern = &UrlPattern{
			BaseUrl: defaultBaseUrl,
			Anchor:  defaultAnchor,
		}
	} else {
		if r.UrlPattern.BaseUrl == "" {
			r.UrlPattern.BaseUrl = defaultBaseUrl
		}

		if r.UrlPattern.Anchor == "" {
			r.UrlPattern.Anchor = defaultAnchor
		}
	}
}

// Populate missing config values with default values.
func initConfig(c *Config) {
	if c.MaxConcurrentIndexers == 0 {
		c.MaxConcurrentIndexers = defaultMaxConcurrentIndexers
	}
	if c.MaxConcurrentSearchers == 0 {
		c.MaxConcurrentSearchers = defaultMaxConcurrentSearchers
	}
	if c.MaxReposInFirstResult == 0 {
		c.MaxReposInFirstResult = defaultMaxReposInFirstResult
	}
	if c.MaxReposInNextResult == 0 {
		c.MaxReposInNextResult = defaultMaxReposInNextResult
	}

	if c.HealthCheckURI == "" {
		c.HealthCheckURI = defaultHealthCheckURI
	}
}

func (c *Config) LoadFromFile(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := json.NewDecoder(r).Decode(c); err != nil {
		return err
	}

	if c.Title == "" {
		c.Title = defaultTitle
	}

	if !filepath.IsAbs(c.DbPath) {
		path, err := filepath.Abs(
			filepath.Join(filepath.Dir(filename), c.DbPath))
		if err != nil {
			return err
		}
		c.DbPath = path
	}

	for _, repo := range c.Repos {
		initRepo(repo)
	}

	initConfig(c)

	return nil
}

func (c *Config) ToJsonString() (string, error) {
	client := ClientConfig{c.InitSearch}
	b, err := json.Marshal(client)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
func get(dict map[string]string, key, dflt string) string {
	if value, ok := dict[key]; ok {
		return value
	}
	return dflt
}

func (c *Config) ToOpenSearchParams() (string, error) {
	// This must be the same as in App.jsx (see const initParams = ...)
	// Exception is for InitSearch.q which is not used here
	i := get(c.InitSearch, "i", "nope")
	files := get(c.InitSearch, "files", "")
	excludeFiles := get(c.InitSearch, "excludeFiles", "")
	repos := get(c.InitSearch, "repos", ".*")
	params := url.Values{}
	params.Add("i", i)
	params.Add("files", files)
	params.Add("excludeFiles", excludeFiles)
	params.Add("repos", repos)

	return params.Encode(), nil
}
