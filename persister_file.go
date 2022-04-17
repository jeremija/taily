package guardlog

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/rand"
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

func (p PersisterFile) filename(watcherID ReaderID) string {
	return path.Join(p.dir, string(watcherID)+".json")
}

func (p PersisterFile) LoadState(ctx context.Context, watcherID ReaderID) (State, error) {
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

func (p PersisterFile) SaveState(ctx context.Context, watcherID ReaderID, state State) error {
	if err := os.MkdirAll(p.dir, 0755); err != nil {
		return errors.Trace(err)
	}

	tmp := make([]byte, 16)

	if _, err := rand.Read(tmp); err != nil {
		return errors.Trace(err)
	}

	// We store to a tmp filename first so we don't lose the old state in case we
	// fail to write it. After a successful write, we rename the tmp file to a
	// new one.
	filename := p.filename(watcherID)
	tmpFilename := p.filename(watcherID) + ".tmp" + hex.EncodeToString(tmp)

	f, err := os.OpenFile(tmpFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Trace(err)
	}

	defer f.Close()

	if err := json.NewEncoder(f).Encode(state); err != nil {
		return errors.Trace(err)
	}

	if err := os.Rename(tmpFilename, filename); err != nil {
		return errors.Trace(err)
	}

	return nil
}
