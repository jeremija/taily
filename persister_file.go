package guardlog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/juju/errors"
)

type PersisterFile struct {
	dir string
}

func NewPersisterFile(dir string) *PersisterFile {
	return &PersisterFile{
		dir: dir,
	}
}

var _ Persister = PersisterFile{}

func (p PersisterFile) filename(watcherID WatcherID) string {
	return path.Join(p.dir, string(watcherID)+".json")
}

func (p PersisterFile) LoadState(ctx context.Context, watcherID WatcherID) (State, error) {
	f, err := os.Open(p.filename(watcherID))
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}

		return State{}, errors.Trace(err)
	}

	defer f.Close()

	var state State

	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return State{}, errors.Trace(err)
	}

	return state, nil
}

func (p PersisterFile) SaveState(ctx context.Context, watcherID WatcherID, state State) error {
	fmt.Println("SaveState", watcherID, state)

	if err := os.MkdirAll(p.dir, 0755); err != nil {
		return errors.Trace(err)
	}

	f, err := os.OpenFile(p.filename(watcherID), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Trace(err)
	}

	defer f.Close()

	if err := json.NewEncoder(f).Encode(state); err != nil {
		return errors.Trace(err)
	}

	return nil
}
