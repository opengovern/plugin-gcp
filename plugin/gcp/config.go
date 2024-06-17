package gcp

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2/google"
)

type GCP struct {
	credentials *google.Credentials
	ProjectID   string `json:"quota_project_id"`
	Scopes      []string
}

func NewGCP(scopes []string) GCP {
	return GCP{
		ProjectID: "",
		Scopes:    scopes,
	}
}

func (g *GCP) GetCredentials(ctx context.Context) error {
	var err error

	g.credentials, err = google.FindDefaultCredentials(
		ctx,
		g.Scopes...,
	)
	if err != nil {
		return err
	}

	g.ProjectID = g.credentials.ProjectID

	json.Unmarshal(g.credentials.JSON, g) //this will store project id from credentials

	return nil
}

func (g *GCP) Identify() map[string]string {

	identification := map[string]string{
		"project_id": g.ProjectID,
	}

	return identification
}
