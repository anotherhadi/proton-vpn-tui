package backend

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type Option struct {
	ID          string
	Description string
}

type Setting struct {
	Description string
	Options     []Option
	Value       string
}

type Config map[string]Setting

func LoadConfig() (c Config, err error) {
	c = make(Config)
	ctx := context.Background()
	res, err := Run(ctx, "config", "list")
	if err != nil {
		return
	}

	inOptions := false
	for line := range strings.SplitSeq(res.Stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			inOptions = false
			continue
		}
		if strings.HasPrefix(fields[0], "---") {
			inOptions = true
			continue
		}
		if !inOptions {
			continue
		}

		c[fields[0]] = Setting{Value: strings.Join(fields[1:], " ")}
	}

	for setting := range c {
		desc, options, e := getInfo(setting)
		if e != nil {
			err = e
			return
		}
		entry := c[setting]
		entry.Description = desc
		entry.Options = options
		c[setting] = entry
	}

	if err = c.SaveCache(); err != nil {
		return
	}

	return
}

func LoadConfigCache() (c Config, err error) {
	path, err := configCachePath()
	if err != nil {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &c)
	return
}

func (c Config) SaveCache() error {
	path, err := configCachePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func configCachePath() (string, error) {
	cacheHome, err := xdgCacheHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheHome, "proton-vpn-tui", "config.json"), nil
}

func (c Config) CycleSettingUp(setting string) (err error) {
	entry, ok := c[setting]
	if !ok || len(entry.Options) == 0 {
		return errors.New("setting unknown")
	}

	idx := -1
	for i, opt := range entry.Options {
		if opt.ID == entry.Value {
			idx = i
			break
		}
	}

	next := idx + 1
	if next >= len(entry.Options) {
		next = 0
	}

	return c.setSetting(setting, entry.Options[next].ID)
}

func (c Config) CycleSettingDown(setting string) (err error) {
	entry, ok := c[setting]
	if !ok || len(entry.Options) == 0 {
		return errors.New("setting unknown")
	}

	idx := -1
	for i, opt := range entry.Options {
		if opt.ID == entry.Value {
			idx = i
			break
		}
	}

	prev := idx - 1
	if prev < 0 {
		prev = len(entry.Options) - 1
	}

	return c.setSetting(setting, entry.Options[prev].ID)
}

func (c Config) setSetting(setting string, value string) error {
	ctx := context.Background()
	_, err := Run(ctx, "config", "set", setting, value)
	if err != nil {
		return err
	}
	entry := c[setting]
	entry.Value = value
	c[setting] = entry
	return c.SaveCache()
}

func getInfo(setting string) (description string, options []Option, err error) {
	ctx := context.Background()
	res, err := Run(ctx, "config", "set", setting, "--help")
	if err != nil {
		return
	}

	inUsage := false
	inDescription := false
	inOptions := false
	descDone := false
	var descLines []string
	for line := range strings.SplitSeq(res.Stdout, "\n") {
		line = strings.TrimSpace(stripControl(line))
		if strings.HasPrefix(line, "Usage:") {
			inUsage = true
			continue
		}
		if line == "Values:" {
			inOptions = true
			continue
		}
		if line == "" {
			if inUsage {
				inUsage = false
				inDescription = !descDone
			} else if len(descLines) > 0 {
				descDone = true
				inDescription = false
			}
			inOptions = false
			continue
		}
		if inUsage {
			continue
		}
		if inOptions {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			options = append(options, Option{ID: fields[0], Description: strings.Join(fields[1:], " ")})
			continue
		}
		if inDescription {
			descLines = append(descLines, line)
		}
	}
	description = strings.Join(descLines, " ")

	return
}

func stripControl(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}
