package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kami-no/reporter/config"
)

type Service interface {
	getCurrent(pid int) (current, error)
	Update(body string) error
}

type confluenceSvc struct {
	cfg *config.CfgService
}

var _ Service = (*confluenceSvc)(nil)

// Create Confluence service
func New(cfg *config.CfgService) *confluenceSvc {
	return &confluenceSvc{
		cfg: cfg,
	}
}

type Post struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
}

type current struct {
	Title   string `json:"title"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
}

func (c *confluenceSvc) getCurrent(pid int) (current, error) {
	var result current

	uri := fmt.Sprintf("%v/rest/api/content/%v", c.cfg.Endpoint, c.cfg.Page)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return result, err
	}
	req.SetBasicAuth(c.cfg.User, c.cfg.Pass)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

func (c *confluenceSvc) Update(body string) error {
	cur, err := c.getCurrent(c.cfg.Page)
	if err != nil {
		return err
	}

	post := Post{
		ID:    fmt.Sprint(c.cfg.Page),
		Type:  "page",
		Title: cur.Title,
	}
	post.Body.Storage.Value = body
	post.Body.Storage.Representation = "storage"

	post.Version.Number = cur.Version.Number + 1

	jsonValue, err := json.Marshal(post)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%v/rest/api/content/%v", c.cfg.Endpoint, c.cfg.Page)

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.cfg.User, c.cfg.Pass)
	req.Header.Set("Content-Type", "application/json")
	mod := &http.Client{}

	resp, err := mod.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to post: %v", resp.StatusCode)
	}

	return nil
}
