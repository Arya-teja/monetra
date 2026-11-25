package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\n\nMONETRA_TEST_A=plain\nMONETRA_TEST_B=\"quoted value\"\r\nMONETRA_TEST_C=from-file\nMONETRA_TEST_URL=postgres://u:p@h/db?sslmode=disable\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MONETRA_TEST_C", "from-env") // already set: must win over the file

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("loadDotEnv() error = %v", err)
	}

	want := map[string]string{
		"MONETRA_TEST_A":   "plain",
		"MONETRA_TEST_B":   "quoted value",
		"MONETRA_TEST_C":   "from-env",
		"MONETRA_TEST_URL": "postgres://u:p@h/db?sslmode=disable",
	}
	for key, value := range want {
		if got := os.Getenv(key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
		os.Unsetenv(key)
	}
}

func TestLoadDotEnv_MissingFileIsIgnored(t *testing.T) {
	if err := loadDotEnv(filepath.Join(t.TempDir(), "does-not-exist")); err != nil {
		t.Fatalf("loadDotEnv() error = %v, want nil", err)
	}
}

func TestLoadDotEnv_RejectsLineWithoutEquals(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("NOT_A_PAIR\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err == nil {
		t.Fatal("loadDotEnv() error = nil, want an error")
	}
}
