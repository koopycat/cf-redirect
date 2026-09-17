package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePrecedenceFlagsThenEnvThenFile(t *testing.T) {
	env := map[string]string{AccountIDEnv: "env-account"}
	lookup := func(key string) (string, bool) { value, ok := env[key]; return value, ok }
	got, err := resolve(" flag-account ", "", lookup, func() (Config, error) {
		return Config{AccountID: "file-account", ListID: " file-list "}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AccountID != "flag-account" || got.ListID != "file-list" {
		t.Fatalf("unexpected config: %#v", got)
	}
}

func TestResolveSkipsFileWhenFlagsAndEnvironmentAreComplete(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == ListIDEnv {
			return "env-list", true
		}
		return "", false
	}
	got, err := resolve("flag-account", "", lookup, func() (Config, error) {
		return Config{}, errors.New("corrupt config must not be loaded")
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != (Config{AccountID: "flag-account", ListID: "env-list"}) {
		t.Fatalf("unexpected config: %#v", got)
	}
}

func TestResolveAccountIDPrecedenceFlagThenEnvThenFile(t *testing.T) {
	t.Run("flag", func(t *testing.T) {
		lookup := func(string) (string, bool) { return "env-account", true }
		got, err := resolveAccountID(" flag-account ", lookup, func() (Config, error) {
			return Config{}, errors.New("config must not be loaded")
		})
		if err != nil || got != "flag-account" {
			t.Fatalf("resolveAccountID() = %q, %v", got, err)
		}
	})

	t.Run("environment", func(t *testing.T) {
		lookup := func(key string) (string, bool) {
			if key == AccountIDEnv {
				return " env-account ", true
			}
			return "", false
		}
		got, err := resolveAccountID("", lookup, func() (Config, error) {
			return Config{}, errors.New("config must not be loaded")
		})
		if err != nil || got != "env-account" {
			t.Fatalf("resolveAccountID() = %q, %v", got, err)
		}
	})

	t.Run("persisted config", func(t *testing.T) {
		got, err := resolveAccountID("", func(string) (string, bool) { return "", false }, func() (Config, error) {
			return Config{AccountID: " file-account "}, nil
		})
		if err != nil || got != "file-account" {
			t.Fatalf("resolveAccountID() = %q, %v", got, err)
		}
	})
}

func TestResolveAccountIDRequiresAccount(t *testing.T) {
	if _, err := resolveAccountID("", func(string) (string, bool) { return "", false }, func() (Config, error) {
		return Config{}, nil
	}); err == nil {
		t.Fatal("expected missing account error")
	}
}

func TestResolveCompletesPersistedAccountWithRuntimeList(t *testing.T) {
	load := func() (Config, error) { return Config{AccountID: " persisted-account "}, nil }

	t.Run("flag", func(t *testing.T) {
		got, err := resolve("", " flag-list ", func(string) (string, bool) { return "", false }, load)
		if err != nil {
			t.Fatal(err)
		}
		if got != (Config{AccountID: "persisted-account", ListID: "flag-list"}) {
			t.Fatalf("unexpected config: %#v", got)
		}
	})

	t.Run("environment", func(t *testing.T) {
		lookup := func(key string) (string, bool) {
			if key == ListIDEnv {
				return " env-list ", true
			}
			return "", false
		}
		got, err := resolve("", "", lookup, load)
		if err != nil {
			t.Fatal(err)
		}
		if got != (Config{AccountID: "persisted-account", ListID: "env-list"}) {
			t.Fatalf("unexpected config: %#v", got)
		}
	})
}

func TestResolveRequiresBothIDs(t *testing.T) {
	lookup := func(string) (string, bool) { return "", false }
	if _, err := resolve("", "list", lookup, func() (Config, error) { return Config{}, nil }); err == nil {
		t.Fatal("expected missing account error")
	}
	if _, err := resolve("", "", lookup, func() (Config, error) {
		return Config{AccountID: "persisted-account"}, nil
	}); err == nil {
		t.Fatal("expected missing list error")
	}
}

func TestLoadRejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	for name, content := range map[string]string{
		"unknown":  `{"account_id":"account","list_id":"list","extra":true}`,
		"trailing": `{"account_id":"account","list_id":"list"} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatal("expected strict JSON error")
			}
		})
	}
}

func TestSaveIsAtomicAndPrivate(t *testing.T) {
	for name, input := range map[string]Config{
		"account and list": {AccountID: " account ", ListID: " list "},
		"account only":     {AccountID: " account ", ListID: "   "},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "nested", "config.json")
			if err := Save(path, input); err != nil {
				t.Fatal(err)
			}
			got, err := Load(path)
			if err != nil || got != (Config{AccountID: "account", ListID: strings.TrimSpace(input.ListID)}) {
				t.Fatalf("Load() = %#v, %v", got, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0o600 {
				t.Fatalf("config mode = %o", info.Mode().Perm())
			}
			matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".config-*"))
			if err != nil || len(matches) != 0 {
				t.Fatalf("temporary files remain: %v, %v", matches, err)
			}
		})
	}
}

func TestSaveRequiresAccount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := Save(path, Config{ListID: "list"}); err == nil {
		t.Fatal("expected missing account error")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config file exists after rejected save: %v", err)
	}
}
