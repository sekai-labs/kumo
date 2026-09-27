package ruleset

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyRulesetID   = errors.New("ruleset ID cannot be empty")
	ErrEmptyRulesetName = errors.New("ruleset name cannot be empty")
)

type RulesetID string

func (id RulesetID) String() string {
	return string(id)
}

func (id RulesetID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyRulesetID
	}
	return nil
}

type Ruleset struct {
	ID          RulesetID `json:"id"`
	Name        string    `json:"name"`
	Phase       string    `json:"phase"`
	Kind        string    `json:"kind"`
	LastUpdated time.Time `json:"last_updated"`
	Description string    `json:"description,omitempty"`
	RulesCount  int       `json:"rules_count"`
}

func NewRuleset(
	id string,
	name string,
	phase string,
	kind string,
	lastUpdated time.Time,
	description string,
	rulesCount int,
) (Ruleset, error) {
	rID := RulesetID(strings.TrimSpace(id))
	if err := rID.Validate(); err != nil {
		return Ruleset{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Ruleset{}, ErrEmptyRulesetName
	}
	if phase == "" {
		phase = "http_request_firewall_custom"
	}
	if kind == "" {
		kind = "zone"
	}
	return Ruleset{
		ID:          rID,
		Name:        name,
		Phase:       phase,
		Kind:        kind,
		LastUpdated: lastUpdated,
		Description: strings.TrimSpace(description),
		RulesCount:  rulesCount,
	}, nil
}
