package worker

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyScriptID = errors.New("worker script ID cannot be empty")
)

type WorkerScriptID string

func (id WorkerScriptID) String() string {
	return string(id)
}

func (id WorkerScriptID) Validate() error {
	if strings.TrimSpace(string(id)) == "" {
		return ErrEmptyScriptID
	}
	return nil
}

type WorkerScript struct {
	ID         WorkerScriptID `json:"id"`
	CreatedOn  time.Time      `json:"created_on"`
	ModifiedOn time.Time      `json:"modified_on"`
	UsageModel string         `json:"usage_model"`
	Logpush    bool           `json:"logpush"`
	HasAssets  bool           `json:"has_assets"`
}

func NewWorkerScript(
	id string,
	createdOn time.Time,
	modifiedOn time.Time,
	usageModel string,
	logpush bool,
	hasAssets bool,
) (WorkerScript, error) {
	scriptID := WorkerScriptID(strings.TrimSpace(id))
	if err := scriptID.Validate(); err != nil {
		return WorkerScript{}, err
	}
	if usageModel == "" {
		usageModel = "standard"
	}
	return WorkerScript{
		ID:         scriptID,
		CreatedOn:  createdOn,
		ModifiedOn: modifiedOn,
		UsageModel: usageModel,
		Logpush:    logpush,
		HasAssets:  hasAssets,
	}, nil
}
