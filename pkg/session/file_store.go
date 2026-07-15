package session

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"ticTacSolved/task/pkg/errs"
)

type FileStore struct {
	path string
}

var _ Store = (*FileStore)(nil)

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Load() (Data, error) {
	raw, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return Data{}, nil
	}
	if err != nil {
		return Data{}, errs.Wrap(
			errs.CodeStorageFailure,
			"failed to read session file",
			err,
		)
	}

	var data Data
	if err = json.Unmarshal(raw, &data); err != nil {
		return Data{}, errs.Wrap(
			errs.CodeStorageFailure,
			"failed to decode session file",
			err,
		)
	}

	return data, nil
}

func (s *FileStore) Save(data Data) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errs.Wrap(
			errs.CodeStorageFailure,
			"failed to encode session data",
			err,
		)
	}
	if err = os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return errs.Wrap(
			errs.CodeStorageFailure,
			"failed to create session dir",
			err,
		)
	}
	if err = os.WriteFile(s.path, raw, 0o600); err != nil {
		return errs.Wrap(
			errs.CodeStorageFailure,
			"failed to write session file",
			err,
		)
	}
	return nil
}

func (s *FileStore) Clear() error {
	err := os.Remove(s.path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return errs.Wrap(
			errs.CodeStorageFailure,
			"failed to remove session file",
			err,
		)
	}
	return nil
}
