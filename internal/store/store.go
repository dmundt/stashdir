package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	term "github.com/AlecAivazis/survey/v2/terminal"
)

type DB struct {
	Path  string   `json:"-"`
	Items []string `json:"items"`
}

// sortInPlace maintains a case-insensitive alphabetical order of Items.
func (d *DB) sortInPlace() {
	sort.Slice(d.Items, func(i, j int) bool {
		return strings.ToLower(d.Items[i]) < strings.ToLower(d.Items[j])
	})
}

func Open() (*DB, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	home := filepath.Join(cfgDir, "stashdir")
	if err := os.MkdirAll(home, 0o755); err != nil {
		return nil, err
	}
	p := filepath.Join(home, "config.json")
	db := &DB{Path: p}
	_ = db.load()
	return db, nil
}

func (d *DB) load() error {
	b, err := os.ReadFile(d.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := json.Unmarshal(b, d); err != nil {
		return err
	}
	d.sortInPlace()
	return nil
}

func (d *DB) save() error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.Path, b, 0o644)
}

func normalize(p string) string {
	p = strings.TrimSpace(p)
	p = filepath.Clean(p)
	// On Windows case-insensitive file system, normalize to lowercase for de-dup
	if isWindows() {
		p = strings.ToLower(p)
	}
	return p
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func (d *DB) Add(p string) error {
	p = normalize(p)
	if p == "" {
		return fmt.Errorf("empty path")
	}
	for _, it := range d.Items {
		if normalize(it) == p {
			return nil // already exists
		}
	}
	d.Items = append(d.Items, p)
	d.sortInPlace()
	return d.save()
}

func (d *DB) List() []string {
	out := make([]string, len(d.Items))
	copy(out, d.Items)
	return out
}

func (d *DB) RemoveIndex(i int) error {
	if i < 0 || i >= len(d.Items) {
		return fmt.Errorf("index out of range")
	}
	d.Items = append(d.Items[:i], d.Items[i+1:]...)
	return d.save()
}

func (d *DB) RemovePath(p string) error {
	p = normalize(p)
	for i, it := range d.Items {
		if normalize(it) == p {
			d.Items = append(d.Items[:i], d.Items[i+1:]...)
			return d.save()
		}
	}
	return fmt.Errorf("path not found")
}

func (d *DB) SelectInteractive() (string, error) {
	items := d.List()
	if len(items) == 0 {
		return "", nil
	}
	prompt := &survey.Select{
		Message:  "Select directory:",
		Options:  items,
		PageSize: 12,
	}
	var choice string
	if err := survey.AskOne(prompt, &choice); err != nil {
		// If the user cancels (ESC/Ctrl+C), return empty without error so callers can exit quietly
		if errors.Is(err, term.InterruptErr) {
			return "", nil
		}
		return "", err
	}
	return choice, nil
}
