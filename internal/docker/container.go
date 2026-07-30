package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Container struct {
	Name   string
	Image  string
	DBType string
	Mounts []Mount
}

type Mount struct {
	Destination string
	Source      string
}

var errNotRunning = errors.New("docker daemon not available or not running")

func Inspect(name string) (*Container, error) {
	out, err := exec.Command("docker", "inspect", name).Output()
	if err != nil {
		return nil, fmt.Errorf("inspecting %q: %w", name, err)
	}

	var raw []struct {
		Name   string `json:"Name"`
		Config struct {
			Image string `json:"Image"`
		} `json:"Config"`
	Mounts []struct {
		Destination string `json:"Destination"`
		Source      string `json:"Source"`
		Type        string `json:"Type"`
	} `json:"Mounts"`
	}

	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing inspect: %w", err)
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("container %q not found", name)
	}

	c := raw[0]
	return &Container{
		Name:   strings.TrimPrefix(c.Name, "/"),
		Image:  c.Config.Image,
		DBType: detectDBType(c.Config.Image),
		Mounts: toMounts(c.Mounts),
	}, nil
}

func Detect() (*Container, error) {
	out, err := exec.Command("docker", "ps",
		"--format", "{{.Names}}|{{.Image}}").Output()
	if err != nil {
		return nil, errors.Join(errNotRunning, err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		name, image := parts[0], parts[1]
		if detectDBType(image) == "" {
			continue
		}

		mounts, err := getMounts(name)
		if err != nil {
			continue
		}

		return &Container{
			Name:   name,
			Image:  image,
			DBType: detectDBType(image),
			Mounts: mounts,
		}, nil
	}

	return nil, fmt.Errorf("no MySQL/MariaDB container found running")
}

func DataDir(c *Container) string {
	for _, m := range c.Mounts {
		if m.Destination == "/var/lib/mysql" {
			return m.Source
		}
	}
	return ""
}

func Stop(name string) error {
	out, err := exec.Command("docker", "stop", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("stop %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	return nil
}

func Start(name string) error {
	out, err := exec.Command("docker", "start", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("start %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	return nil
}

func WaitForHealthy(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := healthStatus(name)
		if err == nil && (status == "healthy" || status == "running") {
			return nil
		}

		running, err := runningStatus(name)
		if err == nil && running {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for container %q after %v", name, timeout)
}

func healthStatus(name string) (string, error) {
	out, err := exec.Command("docker", "inspect",
		"--format", "{{.State.Health.Status}}", name).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func runningStatus(name string) (bool, error) {
	out, err := exec.Command("docker", "inspect",
		"--format", "{{.State.Status}}", name).Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "running", nil
}

// -- helpers --

func detectDBType(image string) string {
	lower := strings.ToLower(image)
	for _, t := range []string{"mariadb", "mysql"} {
		if strings.Contains(lower, t) {
			return t
		}
	}
	return ""
}

func getMounts(container string) ([]Mount, error) {
	out, err := exec.Command("docker", "inspect",
		"--format", "{{json .Mounts}}", container).Output()
	if err != nil {
		return nil, err
	}

	var raw []struct {
		Destination string `json:"Destination"`
		Source      string `json:"Source"`
		Type        string `json:"Type"`
	}

	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}

	return toMounts(raw), nil
}

func toMounts(raw []struct {
	Destination string `json:"Destination"`
	Source      string `json:"Source"`
	Type        string `json:"Type"`
}) []Mount {
	var mounts []Mount
	for _, m := range raw {
		if m.Type != "" && m.Type != "bind" {
			continue
		}
		mounts = append(mounts, Mount{
			Destination: m.Destination,
			Source:      m.Source,
		})
	}
	return mounts
}
