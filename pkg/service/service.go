package service

import (
	"github.com/grid-x/gxctl/pkg/client"
	cfg "github.com/grid-x/gxctl/pkg/config"
	"github.com/grid-x/gxctl/pkg/printer"
)

// Service contains all commonly used APIs
type Service struct {
	Client  *client.APIClient
	Printer *printer.Printer
	Config  *cfg.Config
}

// New returns a new service
func New() (*Service, error) {
	return &Service{
		Config:  &cfg.Config{},
		Printer: printer.NewPrinter(),
	}, nil
}

// Init services for usage by actions
func (s *Service) Init() error {
	if err := s.Config.Read(); err != nil {
		return err
	}

	s.Client = client.NewAPIClient(s.Config.UseStaging, &s.Config.Auth, s.Config.Profile)

	return nil
}
