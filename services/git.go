package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/DoESLiverpool/status/database"

	"github.com/caarlos0/env"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// We pull the args from the system environment
type githubSettings struct {
	// If set to anything not blank and not false it will disable the github service
	Disabled string `env:"GITHUB_DISABLED"`

	// The access token to have higher api limits
	AuthToken string `env:"GITHUB_TOKEN"`

	// The repository to fetch the data from
	Org  string `env:"GITHUB_ORG"`
	Repo string `env:"GITHUB_REPO"`

	// All issues with a label will be marked as broken services if left blank
	LabelPrefix string `env:"GITHUB_LABEL_PREFIX"`
	LabelBroken string `env:"GITHUB_LABEL_BROKEN"`
}

// We only keep the data that is "important" to the system.
type githubLabel struct {
	ID          *int64  `json:"id" binding:"required"`
	Name        *string `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// Again we remove any extra info from the issue that isn't needed
type githubIssue struct {
	ID         *int64         `json:"id" binding:"required"`
	Title      *string        `json:"title" binding:"required"`
	Labels     []*githubLabel `json:"labels"`
	State      *string        `json:"state"`
	DateOpened *time.Time     `json:"dateopened"`
}

// UpdateGit will update the database with the latest settings based on the api
func UpdateGit() ([]*database.Service, error) {
	// Parse environment variables into github settings
	var settings = githubSettings{}
	env.Parse(&settings)

	if settings.Disabled != "" && settings.Disabled != "false" {
		return nil, nil
	}

	if settings.AuthToken == "" {
		return nil, errors.New("Missing github auth token")
	}

	if settings.Org == "" {
		return nil, errors.New("Missing github organisation")
	}

	if settings.Repo == "" {
		return nil, errors.New("Missing github repository")
	}

	labels, err := getLabels(settings)

	if err != nil {
		return nil, err
	}

	issues, err := getOpenIssues(settings, labels)

	if err != nil {
		return nil, err
	}

	services := getGitServices(labels, issues)

	if services == nil {
		return nil, errors.New("No services returned by github")
	}

	return services, nil
}

// Get all the labels for the repository
func getLabels(details githubSettings) ([]*githubLabel, error) {
	client, ctx := getClient(details)

	var opts = github.ListOptions{
		Page:    0,
		PerPage: 50,
	}

	var allLabels []*github.Label
	keepPaging := true

	// Loop through all of the label pages
	for keepPaging {
		labels, _, err := client.Issues.ListLabels(*ctx, details.Org, details.Repo, &opts)

		allLabels = append(allLabels, labels...)

		if err != nil {
			return nil, err
		}

		opts.Page++

		if opts.Page*opts.PerPage != len(allLabels) {
			keepPaging = false
		}
	}

	var labels []*githubLabel

	// Convert the github labels into our label
	// This is done for memory footprint and json output
	for _, gitLabel := range allLabels {

		// Ensure the label has the prefix we are looking for
		if strings.HasPrefix(*gitLabel.Name, details.LabelPrefix) {
			labels = append(labels, &githubLabel{
				ID:          gitLabel.ID,
				Name:        gitLabel.Name,
				Description: gitLabel.Description,
			})
		}
	}

	return labels, nil
}

// Get all the issues for the labels set. The issue also require having the GITHUB_LABEL_BROKEN label to appear
func getOpenIssues(details githubSettings, labels []*githubLabel) ([]*githubIssue, error) {
	client, ctx := getClient(details)

	var allIssues []*github.Issue

	// Loop through all the labels looking for issues
	// Only finds issues that have both the GITHUB_LABEL_BROKEN label and the label in the array
	for i := 0; i < len(labels); i++ {
		labelNames := []string{details.LabelBroken, *labels[i].Name}

		var opts = github.ListOptions{
			Page:    0,
			PerPage: 50,
		}

		keepPaging := true

		// Loop through all the issue pages for the labels specified
		for keepPaging {
			issues, _, err := client.Issues.ListByRepo(*ctx, details.Org, details.Repo, &github.IssueListByRepoOptions{
				Labels: labelNames,

				ListOptions: opts,
			})

			allIssues = append(allIssues, issues...)

			if err != nil {
				return nil, err
			}

			opts.Page++

			if opts.Page*opts.PerPage != len(allIssues) {
				keepPaging = false
			}
		}
	}

	var issues []*githubIssue

	// Convert github issues into our issues.
	// This is done for memory footprint and json output
	for _, gitIssue := range allIssues {

		issue := &githubIssue{
			ID:         gitIssue.ID,
			Title:      gitIssue.Title,
			State:      gitIssue.State,
			DateOpened: gitIssue.CreatedAt,
		}

		// Don't add the issue if it is already in the array.
		// This can happen if the issue belongs to multiple labels
		if !issueExists(issue, issues) {

			// Convert github labels into our labels.
			for _, gitLabel := range gitIssue.Labels {
				label := &githubLabel{
					ID:          gitLabel.ID,
					Name:        gitLabel.Name,
					Description: gitLabel.Description,
				}

				issue.Labels = append(issue.Labels, label)
			}

			issues = append(issues, issue)
		}
	}

	return issues, nil
}

// Return false if the issue is already in the array. We match based on ID
func issueExists(is *githubIssue, iss []*githubIssue) bool {
	for _, i := range iss {
		if *i.ID == *is.ID {
			return true
		}
	}

	return false
}

// Return an authorised github client along with the context it is using
func getClient(details githubSettings) (*github.Client, *context.Context) {
	if details.AuthToken == "" {
		panic("Missing auth token")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: details.AuthToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client, &ctx
}

// Gets all of the github services. Marks broken ones and sets their timestamp
func getGitServices(labels []*githubLabel, issues []*githubIssue) []*database.Service {
	var services []*database.Service

	// Create a service for each label
	for _, label := range labels {
		service := &database.Service{
			ID:    *label.ID,
			Name:  *label.Name,
			State: database.WorkingState,
			Since: time.Now(),
		}

		if label.Description != nil {
			service.Description = *label.Description
		}

		services = append(services, service)
	}

	// Loop through each issue
	for _, issue := range issues {

		// Find the service for each label and mark as broken
		for _, label := range issue.Labels {
			service := getServiceByID(services, *label.ID)

			if service != nil {
				service.State = database.BrokenState
				service.Since = *issue.DateOpened
			}
		}
	}

	return services
}
