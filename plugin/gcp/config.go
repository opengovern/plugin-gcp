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
	credentials, err := google.FindDefaultCredentials(
		ctx,
		g.Scopes...,
	)
	if err != nil {
		return err
	}

	json.Unmarshal(credentials.JSON, g) //this will store project id from credentials

	g.credentials = credentials
	// log.Println(g.ProjectID)
	return nil
}
