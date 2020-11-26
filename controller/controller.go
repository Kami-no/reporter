package controller

import (
	"github.com/Kami-no/reporter/config"
	"github.com/Kami-no/reporter/services/confluence"
	"github.com/Kami-no/reporter/services/gitlabsvc"
)

// Controller
type Controller struct {
	Config *config.Config
	Git    gitlabsvc.Service
	Wiki   confluence.Service
}

// Generate controller
func New(config *config.Config) *Controller {
	return &Controller{
		Config: config,
		Git:    gitlabsvc.New(config),
		Wiki:   confluence.New(&config.Confluence),
	}
}

type Assignee struct {
	Issues []gitlabsvc.Issue
}

func (c *Controller) GetProjectAssignees(pid int) (map[string]Assignee, error) {
	output := make(map[string]Assignee)

	issues, err := c.Git.GetProjectClosedIssues(pid)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		assignee := output[issue.Assignee]
		assignee.Issues = append(assignee.Issues, issue)
		output[issue.Assignee] = assignee
	}

	return output, nil
}
