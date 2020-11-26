package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/Kami-no/reporter/config"
	"github.com/Kami-no/reporter/controller"
)

type Project struct {
	Name      string
	Assignees map[string]controller.Assignee
}

func main() {
	log.Println("Preparing the report.")

	ctrl := controller.New(config.New())

	// Get projets data
	projects := make(map[int]Project)
	for pid, project := range ctrl.Config.Projects {
		assignees, err := ctrl.GetProjectAssignees(pid)
		if err != nil {
			log.Fatalln(err)
		}
		var prj Project
		prj.Name = project
		prj.Assignees = assignees

		projects[pid] = prj
	}

	// Render template
	tmpl := template.Must(template.ParseFiles("templates/confluence.gohtml"))
	var t bytes.Buffer
	if err := tmpl.Execute(&t, projects); err != nil {
		log.Fatalln(err)
	}
	report := t.String()

	// Post report
	if err := ctrl.Wiki.Update(report); err != nil {
		log.Fatal(err)
	}

	log.Println("The report has been published.")
}
