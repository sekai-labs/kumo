package pages

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyProjectID   = errors.New("pages project ID cannot be empty")
	ErrEmptyProjectName = errors.New("pages project name cannot be empty")
)

type PagesProjectID string

func (id PagesProjectID) String() string {
	return string(id)
}

func (id PagesProjectID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyProjectID
	}
	return nil
}

type PagesProject struct {
	ID               PagesProjectID `json:"id"`
	Name             string         `json:"name"`
	Subdomain        string         `json:"subdomain"`
	ProductionBranch string         `json:"production_branch"`
	Domains          []string       `json:"domains"`
	CreatedOn        time.Time      `json:"created_on"`
}

func NewPagesProject(
	id string,
	name string,
	subdomain string,
	productionBranch string,
	domains []string,
	createdOn time.Time,
) (PagesProject, error) {
	projID := PagesProjectID(strings.TrimSpace(id))
	if err := projID.Validate(); err != nil {
		return PagesProject{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return PagesProject{}, ErrEmptyProjectName
	}
	if productionBranch == "" {
		productionBranch = "main"
	}
	return PagesProject{
		ID:               projID,
		Name:             name,
		Subdomain:        strings.TrimSpace(subdomain),
		ProductionBranch: productionBranch,
		Domains:          domains,
		CreatedOn:        createdOn,
	}, nil
}
