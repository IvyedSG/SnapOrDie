package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Snapshot struct {
	Name      string    `json:"name"`
	Created   time.Time `json:"created"`
	DataDir   string    `json:"data_dir"`
	Container string    `json:"container"`
	SizeBytes int64     `json:"size_bytes"`
}

type manifest struct {
	Snapshots []Snapshot `json:"snapshots"`
	Default   string     `json:"default"`
}

type Manager struct {
	SnapDir   string
	DataDir   string
	Container string
}

func NewManager(dataDir, container string) *Manager {
	snapDir := filepath.Join(filepath.Dir(dataDir), ".snapordie")
	return &Manager{
		SnapDir:   snapDir,
		DataDir:   dataDir,
		Container: container,
	}
}

func (m *Manager) Save(name string) (*Snapshot, error) {
	if err := os.MkdirAll(m.SnapDir, 0755); err != nil {
		return nil, fmt.Errorf("creating snap dir: %w", err)
	}

	if name == "" {
		name = time.Now().Format("20060102-150405")
	}

	dst := filepath.Join(m.SnapDir, name)

	if _, err := os.Stat(dst); err == nil {
		return nil, fmt.Errorf("snapshot %q already exists", name)
	}

	if err := cloneDir(m.DataDir, dst); err != nil {
		return nil, fmt.Errorf("cloning data dir: %w", err)
	}

	size, err := dirSize(dst)
	if err != nil {
		size = 0
	}

	snap := Snapshot{
		Name:      name,
		Created:   time.Now(),
		DataDir:   m.DataDir,
		Container: m.Container,
		SizeBytes: size,
	}

	if err := m.addToManifest(snap); err != nil {
		return nil, fmt.Errorf("updating manifest: %w", err)
	}

	return &snap, nil
}

func (m *Manager) Reset(name string) error {
	if name == "" {
		man, err := m.readManifest()
		if err != nil {
			return fmt.Errorf("no snapshots found")
		}
		if man.Default == "" {
			if len(man.Snapshots) == 0 {
				return fmt.Errorf("no snapshots found — run 'save' first")
			}
			name = man.Snapshots[len(man.Snapshots)-1].Name
		} else {
			name = man.Default
		}
	}

	src := filepath.Join(m.SnapDir, name)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %q not found", name)
	}

	if err := os.RemoveAll(m.DataDir); err != nil {
		return fmt.Errorf("removing data dir: %w", err)
	}

	if err := cloneDir(src, m.DataDir); err != nil {
		return fmt.Errorf("restoring snapshot: %w", err)
	}

	return nil
}

func (m *Manager) List() ([]Snapshot, error) {
	man, err := m.readManifest()
	if err != nil {
		return nil, nil
	}

	// Re-scan to get accurate sizes
	for i := range man.Snapshots {
		snapPath := filepath.Join(m.SnapDir, man.Snapshots[i].Name)
		size, err := dirSize(snapPath)
		if err == nil {
			man.Snapshots[i].SizeBytes = size
		}
	}

	sort.Slice(man.Snapshots, func(i, j int) bool {
		return man.Snapshots[i].Created.After(man.Snapshots[j].Created)
	})
	return man.Snapshots, nil
}

func (m *Manager) Info(name string) (*Snapshot, error) {
	man, err := m.readManifest()
	if err != nil {
		return nil, fmt.Errorf("no snapshots found")
	}

	for _, s := range man.Snapshots {
		if s.Name == name {
			snapPath := filepath.Join(m.SnapDir, s.Name)
			size, err := dirSize(snapPath)
			if err == nil {
				s.SizeBytes = size
			}
			return &s, nil
		}
	}
	return nil, fmt.Errorf("snapshot %q not found", name)
}

func (m *Manager) Remove(name string) error {
	snapPath := filepath.Join(m.SnapDir, name)
	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		return fmt.Errorf("snapshot %q not found", name)
	}

	if err := os.RemoveAll(snapPath); err != nil {
		return fmt.Errorf("removing snapshot: %w", err)
	}

	man, err := m.readManifest()
	if err == nil {
		var kept []Snapshot
		for _, s := range man.Snapshots {
			if s.Name != name {
				kept = append(kept, s)
			}
		}
		man.Snapshots = kept
		if man.Default == name {
			man.Default = ""
		}
		m.writeManifest(man)
	}

	return nil
}

func (m *Manager) addToManifest(snap Snapshot) error {
	man, err := m.readManifest()
	if err != nil {
		man = &manifest{}
	}
	man.Snapshots = append(man.Snapshots, snap)
	if man.Default == "" {
		man.Default = snap.Name
	}
	return m.writeManifest(man)
}

func (m *Manager) readManifest() (*manifest, error) {
	path := filepath.Join(m.SnapDir, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var man manifest
	if err := json.Unmarshal(data, &man); err != nil {
		return nil, err
	}
	return &man, nil
}

func (m *Manager) writeManifest(man *manifest) error {
	path := filepath.Join(m.SnapDir, "manifest.json")
	data, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func cloneDir(src, dst string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("cp", "-c", src, dst)
	case "linux":
		cmd = exec.Command("cp", "--reflink=always", "-a", src, dst)
	default:
		cmd = exec.Command("cp", "-a", src, dst)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to regular copy
		fallback := exec.Command("cp", "-a", src, dst)
		if fbOut, fbErr := fallback.CombinedOutput(); fbErr != nil {
			return fmt.Errorf("clone failed (CoW + fallback): %s: %w; fallback: %s: %v",
				strings.TrimSpace(string(out)), err,
				strings.TrimSpace(string(fbOut)), fbErr)
		}
	}
	return nil
}

func dirSize(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func HumanSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
