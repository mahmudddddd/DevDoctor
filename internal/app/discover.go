package app

import (
	"context"
	"fmt"

	"github.com/mahmudddddd/DevDoctor/internal/detect"
	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
	"github.com/mahmudddddd/DevDoctor/internal/version"
)

type discoverer interface {
	Discover(context.Context, string) (model.ProjectSummary, error)
}

// DiscoveryService orchestrates safe project discovery and report creation.
type DiscoveryService struct {
	discoverer discoverer
}

// NewDiscoveryService creates a discovery service with the default safe file policy.
func NewDiscoveryService() DiscoveryService {
	return DiscoveryService{
		discoverer: detect.NewProjectDetector(privacy.NewFilePolicy()),
	}
}

// Diagnose safely inspects a project and returns a versioned discovery report.
func (s DiscoveryService) Diagnose(ctx context.Context, path string) (model.ProjectReport, error) {
	if s.discoverer == nil {
		return model.ProjectReport{}, fmt.Errorf("project discoverer is not configured")
	}

	project, err := s.discoverer.Discover(ctx, path)
	if err != nil {
		return model.ProjectReport{}, err
	}

	return model.ProjectReport{
		SchemaVersion: model.ReportSchemaVersion,
		ToolVersion:   version.Version,
		Project:       project,
	}, nil
}
