package persister

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os"
	"path"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// File is an implementation of Persister that keeps track of all
// state on local disk.
type File struct {
	dir string
}

// NewFile creates a new instance of File. The dir parameter
// must be writable by the current process.
func NewFile(dir string) *File {
	return &File{
		dir: dir,
	}
}

// Assert that File implements types.Persister.
var _ types.Persister = File{}

// filename returns a filename for watcherID.
func (p File) filename(watcherID types.ReaderID) string {
	return path.Join(p.dir, string(watcherID)+".json")
}

// LoadState implements Persister.
func (p File) LoadState(ctx context.Context, watcherID types.ReaderID) (types.State, error) {
	f, err := os.Open(p.filename(watcherID))
	if err != nil {
		if os.IsNotExist(err) {
			return types.State{}, nil
		}

		return types.State{}, errors.Trace(err)
	}

	defer f.Close()

	var state types.State

	if err := json.NewDecoder(f).Decode(&state); err != nil {
		return types.State{}, errors.Trace(err)
	}

	return state, nil
}

// SaveState implements Persister.
func (p File) SaveState(ctx context.Context, watcherID types.ReaderID, state types.State) error {
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
