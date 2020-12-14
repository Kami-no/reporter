package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Kami-no/reporter/config"
)

type Service interface {
	getCurrent(pid int) (Current, error)
	searchPage(title string) (int, error)
	postNew(title string, body string) error
	postUpdate(pid int, body string) error
	Publish(body string) error
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

type NewPost struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Space struct {
		Key string `json:"key"`
	} `json:"space"`
	Ancestors []Ancestor `json:"ancestors"`
	Body      struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
}

type Ancestor struct {
	ID string `json:"id"`
}

type Current struct {
	Title   string `json:"title"`
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
}

type SearchResults struct {
	Size    int `json:"size"`
	Results []struct {
		ID string `json:"id"`
	} `json:"results"`
}

// Search page by title
func (c *confluenceSvc) searchPage(title string) (int, error) {
	var results SearchResults
	var output int

	uri := fmt.Sprintf("%v/rest/api/content?spaceKey=%v&title=%v", c.cfg.Endpoint, c.cfg.Space, title)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return output, err
	}
	req.SetBasicAuth(c.cfg.User, c.cfg.Pass)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return output, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return output, err
		}
		body := string(bodyBytes)

		return output, fmt.Errorf("failed to post: %v %v", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return output, err
	}

	switch results.Size {
	case 0:
		return 0, nil
	case 1:
		output, err = strconv.Atoi(results.Results[0].ID)
		if err != nil {
			return output, err
		}
		return output, nil
	}

	return output, fmt.Errorf("Result size: %v", results.Size)
}

// Post new page as a child to existing
func (c *confluenceSvc) postNew(title string, body string) error {
	post := NewPost{
		Type:  "page",
		Title: title,
	}
	post.Space.Key = c.cfg.Space
	post.Body.Storage.Value = body
	post.Body.Storage.Representation = "storage"
	post.Ancestors = append(post.Ancestors, Ancestor{ID: fmt.Sprint(c.cfg.Page)})

	jsonValue, err := json.Marshal(post)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%v/rest/api/content/", c.cfg.Endpoint)

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonValue))
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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		body := string(bodyBytes)

		return fmt.Errorf("failed to post: %v %v", resp.StatusCode, body)
	}

	return nil
}

// Get existing page current revision info
func (c *confluenceSvc) getCurrent(pid int) (Current, error) {
	var result Current

	uri := fmt.Sprintf("%v/rest/api/content/%v", c.cfg.Endpoint, pid)

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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}
		body := string(bodyBytes)

		return result, fmt.Errorf("failed to post: %v %v", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

// Update existing page
func (c *confluenceSvc) postUpdate(pid int, body string) error {
	cur, err := c.getCurrent(pid)
	if err != nil {
		return err
	}

	post := Post{
		ID:    fmt.Sprint(pid),
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

	uri := fmt.Sprintf("%v/rest/api/content/%v", c.cfg.Endpoint, pid)

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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		body := string(bodyBytes)

		return fmt.Errorf("failed to post: %v %v", resp.StatusCode, body)
	}

	return nil
}

func (c *confluenceSvc) Publish(body string) error {
	title := time.Now().Format("2006-01-02")

	pid, err := c.searchPage(title)
	if err != nil {
		return err
	}

	fmt.Print(pid)

	if pid == 0 {
		if err := c.postNew(title, body); err != nil {
			return err
		}
	} else {
		if err := c.postUpdate(pid, body); err != nil {
			return err
		}
	}

	return nil
}
