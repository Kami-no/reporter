package gitlabsvc

import (
	"fmt"
	"time"

	"github.com/xanzy/go-gitlab"

	"github.com/Kami-no/reporter/config"
)

type Service interface {
	GetProjectUpdatedIssues(project int) ([]Issue, error)
	getProjectIssues(pid int, issueOpts *gitlab.ListProjectIssuesOptions) ([]Issue, error)
}

type gitlabSvc struct {
	cfg *config.Config
}

var _ Service = (*gitlabSvc)(nil)

// Create GitLab service
func New(cfg *config.Config) *gitlabSvc {
	return &gitlabSvc{
		cfg: cfg,
	}
}

type Issue struct {
	Title    string
	State    string
	Assignee string
	URL      string
	ID       string
}

func (g *gitlabSvc) GetProjectUpdatedIssues(pid int) ([]Issue, error) {
	// Setup filter
	issueOpts := &gitlab.ListProjectIssuesOptions{
		OrderBy:      gitlab.String("updated_at"),
		UpdatedAfter: gitlab.Time(time.Now().AddDate(0, 0, -7)),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	out, err := g.getProjectIssues(pid, issueOpts)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (g *gitlabSvc) getProjectIssues(pid int, issueOpts *gitlab.ListProjectIssuesOptions) ([]Issue, error) {
	// Establish GitLab connection
	git, err := gitlab.NewBasicAuthClient(
		g.cfg.GitLab.User,
		g.cfg.GitLab.Pass,
		gitlab.WithBaseURL(g.cfg.GitLab.Endpoint))
	if err != nil {
		return nil, err
	}

	var output []Issue

	// Process all issues without limitations
	for {
		issues, r, err := git.Issues.ListProjectIssues(pid, issueOpts)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			if issue.Assignee != nil {
				i := Issue{
					Title:    issue.Title,
					Assignee: issue.Assignee.Name,
					State:    issue.State,
					URL:      issue.WebURL,
					ID:       fmt.Sprintf("%04d", issue.IID),
				}
				output = append(output, i)
			}
		}

		// Switch to the next page
		if r.CurrentPage >= r.TotalPages {
			break
		}
		issueOpts.Page = r.NextPage
	}

	return output, nil
}
