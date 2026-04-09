package configutil

import (
	"os"
	"testing"
)

func Test_parseEnvLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantErr   bool
	}{
		{name: "simple", line: "KEY=value", wantKey: "KEY", wantValue: "value"},
		{name: "inline comment", line: "KEY=value # comment", wantKey: "KEY", wantValue: "value"},
		{name: "double quoted with hash", line: `KEY="value # not a comment"`, wantKey: "KEY", wantValue: "value # not a comment"},
		{name: "single quoted with hash", line: `KEY='value # not a comment'`, wantKey: "KEY", wantValue: "value # not a comment"},
		{name: "no equals", line: "INVALID", wantErr: true},
		{name: "hash without space", line: "KEY=value#notcomment", wantKey: "KEY", wantValue: "value#notcomment"},
		{name: "empty value", line: "KEY=", wantKey: "KEY", wantValue: ""},
		{name: "spaces around equals", line: "KEY = value", wantKey: "KEY", wantValue: "value"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, value, err := parseEnvLine(tc.line)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseEnvLine() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if key != tc.wantKey || value != tc.wantValue {
				t.Errorf("parseEnvLine() = (%q, %q), want (%q, %q)", key, value, tc.wantKey, tc.wantValue)
			}
		})
	}
}

func Test_parseEnvFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		tmp := t.TempDir() + "/test.env"
		content := "KEY1=val1\n# comment\nKEY2=val2\n\nKEY3=val3 # inline\n"
		if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		m, err := parseEnvFile(tmp)
		if err != nil {
			t.Fatal(err)
		}

		want := map[string]string{"KEY1": "val1", "KEY2": "val2", "KEY3": "val3"}
		for k, wantV := range want {
			if got := m[k]; got != wantV {
				t.Errorf("key %q: got %q, want %q", k, got, wantV)
			}
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := parseEnvFile("/nonexistent/path.env")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func Test_fileSource_invalidExtension(t *testing.T) {
	src := fileSource{filepath: "config.txt"}
	_, err := src.Load()
	if err == nil {
		t.Error("expected error for non-.env extension")
	}
}

func Test_fileSource_caseInsensitiveExtension(t *testing.T) {
	tmp := t.TempDir() + "/config.ENV"
	if err := os.WriteFile(tmp, []byte("KEY=val\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := fileSource{filepath: tmp}
	m, err := src.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}
	if m["KEY"] != "val" {
		t.Errorf("got %q, want %q", m["KEY"], "val")
	}
}

func Test_flagSource_unparsed(t *testing.T) {
	// flag.Parsed() is true in tests (go test calls flag.Parse()),
	// so we test that Load returns a non-nil map.
	src := flagSource{}
	m, err := src.Load()
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Error("expected non-nil map")
	}
}

func Test_environmentVariableSource(t *testing.T) {
	t.Setenv("CONFIGUTIL_TEST_KEY", "testval")

	src := environmentVariableSource{}
	m, err := src.Load()
	if err != nil {
		t.Fatal(err)
	}

	if m["CONFIGUTIL_TEST_KEY"] != "testval" {
		t.Errorf("got %q, want %q", m["CONFIGUTIL_TEST_KEY"], "testval")
	}
}
