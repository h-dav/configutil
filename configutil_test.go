package configutil_test

import (
	"errors"
	"flag"
	"os"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h-dav/configutil"
)

// ---------------------------------------------------------------------------
// Set — basic functionality
// ---------------------------------------------------------------------------

func TestSet(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		type Config struct {
			String string  `config:"STRING"`
			Int    int     `config:"INT"`
			Float  float64 `config:"FLOAT"`
			Bool   bool    `config:"BOOL"`
		}

		t.Setenv("STRING", "value")
		t.Setenv("INT", "10")
		t.Setenv("FLOAT", "1.2")
		t.Setenv("BOOL", "true")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		want := Config{String: "value", Int: 10, Float: 1.2, Bool: true}
		if diff := cmp.Diff(want, cfg); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("default values", func(t *testing.T) {
		type Config struct {
			Default string `config:"NON_EXISTENT,default=fallback"`
		}

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Default != "fallback" {
			t.Errorf("got %q, want %q", cfg.Default, "fallback")
		}
	})

	t.Run("explicit zero values not overwritten by defaults", func(t *testing.T) {
		type Config struct {
			Str  string `config:"ZERO_STR,default=fallback"`
			Num  int    `config:"ZERO_NUM,default=42"`
			Flag bool   `config:"ZERO_BOOL,default=true"`
		}

		t.Setenv("ZERO_STR", "")
		t.Setenv("ZERO_NUM", "0")
		t.Setenv("ZERO_BOOL", "false")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}

		if cfg.Str != "" {
			t.Errorf("Str: got %q, want empty", cfg.Str)
		}
		if cfg.Num != 0 {
			t.Errorf("Num: got %d, want 0", cfg.Num)
		}
		if cfg.Flag != false {
			t.Errorf("Flag: got %v, want false", cfg.Flag)
		}
	})

	t.Run("required field missing", func(t *testing.T) {
		type Config struct {
			Required string `config:"MISSING_REQ,required"`
		}

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrRequired) {
			t.Errorf("got %v, want ErrRequired", err)
		}
	})

	t.Run("required field present", func(t *testing.T) {
		type Config struct {
			Required string `config:"PRESENT_REQ,required"`
		}

		t.Setenv("PRESENT_REQ", "ok")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Required != "ok" {
			t.Errorf("got %q, want %q", cfg.Required, "ok")
		}
	})

	t.Run("text replacement single", func(t *testing.T) {
		type Config struct {
			Host string `config:"HOST"`
			URL  string `config:"URL"`
		}

		t.Setenv("HOST", "localhost")
		t.Setenv("URL", "http://${HOST}:8080")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.URL != "http://localhost:8080" {
			t.Errorf("got %q, want %q", cfg.URL, "http://localhost:8080")
		}
	})

	t.Run("text replacement multiple", func(t *testing.T) {
		type Config struct {
			Host string `config:"HOST"`
			Port string `config:"PORT"`
			URL  string `config:"URL"`
		}

		t.Setenv("HOST", "localhost")
		t.Setenv("PORT", "9090")
		t.Setenv("URL", "http://${HOST}:${PORT}/api")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.URL != "http://localhost:9090/api" {
			t.Errorf("got %q, want %q", cfg.URL, "http://localhost:9090/api")
		}
	})

	t.Run("text replacement missing var", func(t *testing.T) {
		type Config struct {
			URL string `config:"URL"`
		}

		t.Setenv("URL", "http://${MISSING_HOST}:8080")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrReplacement) {
			t.Errorf("got %v, want ErrReplacement", err)
		}
	})

	t.Run("text replacement with empty value", func(t *testing.T) {
		type Config struct {
			Host string `config:"HOST"`
			URL  string `config:"URL"`
		}

		t.Setenv("HOST", "")
		t.Setenv("URL", "http://${HOST}:8080")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.URL != "http://:8080" {
			t.Errorf("got %q, want %q", cfg.URL, "http://:8080")
		}
	})

	t.Run("slices", func(t *testing.T) {
		type Config struct {
			Strings []string  `config:"STRINGS"`
			Ints    []int     `config:"INTS"`
			Floats  []float64 `config:"FLOATS"`
		}

		t.Setenv("STRINGS", "a,b,c")
		t.Setenv("INTS", "1,2,3")
		t.Setenv("FLOATS", "1.1,2.2")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}

		if !slices.Equal(cfg.Strings, []string{"a", "b", "c"}) {
			t.Errorf("Strings: got %v", cfg.Strings)
		}
		if !slices.Equal(cfg.Ints, []int{1, 2, 3}) {
			t.Errorf("Ints: got %v", cfg.Ints)
		}
		if !slices.Equal(cfg.Floats, []float64{1.1, 2.2}) {
			t.Errorf("Floats: got %v", cfg.Floats)
		}
	})

	t.Run("slice with spaces", func(t *testing.T) {
		type Config struct {
			Values []string `config:"SPACE_VALS"`
		}

		t.Setenv("SPACE_VALS", "a, ,b")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if len(cfg.Values) != 3 || cfg.Values[1] != "" {
			t.Errorf("got %v", cfg.Values)
		}
	})

	t.Run("nested struct with prefix", func(t *testing.T) {
		type Config struct {
			Server struct {
				Port int `config:"PORT"`
			} `config:",prefix=SERVER_"`
		}

		t.Setenv("SERVER_PORT", "8080")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("got %d, want 8080", cfg.Server.Port)
		}
	})

	t.Run("deeply nested structs", func(t *testing.T) {
		type Config struct {
			Server struct {
				Database struct {
					User string `config:"USER"`
				} `config:",prefix=DB_"`
			} `config:",prefix=SERVER_"`
		}

		t.Setenv("SERVER_DB_USER", "admin")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Server.Database.User != "admin" {
			t.Errorf("got %q, want %q", cfg.Server.Database.User, "admin")
		}
	})

	t.Run("unexported fields are skipped", func(t *testing.T) {
		type Config struct {
			Public  string `config:"PUBLIC"`
			private string `config:"PRIVATE"` //nolint:unused
		}

		t.Setenv("PUBLIC", "visible")
		t.Setenv("PRIVATE", "hidden")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Public != "visible" {
			t.Errorf("got %q, want %q", cfg.Public, "visible")
		}
	})

	t.Run("field without tag is skipped", func(t *testing.T) {
		type Config struct {
			Tagged   string `config:"TAGGED"`
			Untagged string
		}

		t.Setenv("TAGGED", "value")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Tagged != "value" {
			t.Errorf("Tagged: got %q", cfg.Tagged)
		}
		if cfg.Untagged != "" {
			t.Errorf("Untagged: got %q, want empty", cfg.Untagged)
		}
	})

	t.Run("malformed default value", func(t *testing.T) {
		type Config struct {
			Port int `config:"UNSET_PORT,default=notanumber"`
		}

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrTag) {
			t.Errorf("got %v, want ErrTag", err)
		}
		var mde *configutil.MalformedDefaultError
		if !errors.As(err, &mde) {
			t.Fatalf("errors.As MalformedDefaultError failed")
		}
		if mde.Default != "notanumber" {
			t.Errorf("Default = %q, want %q", mde.Default, "notanumber")
		}
		if mde.FieldName != "Port" {
			t.Errorf("FieldName = %q, want %q", mde.FieldName, "Port")
		}
	})

	t.Run("conversion error int", func(t *testing.T) {
		type Config struct {
			Port int `config:"BAD_PORT"`
		}

		t.Setenv("BAD_PORT", "notanumber")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrConversion) {
			t.Errorf("got %v, want ErrConversion", err)
		}
	})

	t.Run("conversion error bool", func(t *testing.T) {
		type Config struct {
			Flag bool `config:"BAD_BOOL"`
		}

		t.Setenv("BAD_BOOL", "notabool")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrConversion) {
			t.Errorf("got %v, want ErrConversion", err)
		}
	})

	t.Run("conversion error float", func(t *testing.T) {
		type Config struct {
			Rate float64 `config:"BAD_FLOAT"`
		}

		t.Setenv("BAD_FLOAT", "notafloat")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrConversion) {
			t.Errorf("got %v, want ErrConversion", err)
		}
	})

	t.Run("conversion error int slice", func(t *testing.T) {
		type Config struct {
			Nums []int `config:"BAD_INTS"`
		}

		t.Setenv("BAD_INTS", "1,two,3")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrConversion) {
			t.Errorf("got %v, want ErrConversion", err)
		}
	})

	t.Run("conversion error float slice", func(t *testing.T) {
		type Config struct {
			Rates []float64 `config:"BAD_FLOATS"`
		}

		t.Setenv("BAD_FLOATS", "1.1,bad,3.3")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrConversion) {
			t.Errorf("got %v, want ErrConversion", err)
		}
	})

	t.Run("required field with explicit zero value is not an error", func(t *testing.T) {
		type Config struct {
			Flag bool `config:"REQ_BOOL_ZERO,required"`
			Num  int  `config:"REQ_INT_ZERO,required"`
		}

		t.Setenv("REQ_BOOL_ZERO", "false")
		t.Setenv("REQ_INT_ZERO", "0")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Errorf("Set() unexpected error = %v", err)
		}
		if cfg.Flag != false {
			t.Errorf("Flag: got %v, want false", cfg.Flag)
		}
		if cfg.Num != 0 {
			t.Errorf("Num: got %d, want 0", cfg.Num)
		}
	})

	t.Run("required field unset still errors", func(t *testing.T) {
		type Config struct {
			Name string `config:"UNSET_REQ_STR,required"`
		}

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrRequired) {
			t.Errorf("got %v, want ErrRequired", err)
		}
	})

	t.Run("named string type", func(t *testing.T) {
		type Hostname string
		type Config struct {
			Host Hostname `config:"NAMED_HOST"`
		}

		t.Setenv("NAMED_HOST", "localhost")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if cfg.Host != "localhost" {
			t.Errorf("got %q, want %q", cfg.Host, "localhost")
		}
	})

	t.Run("named int type", func(t *testing.T) {
		type Port int
		type Config struct {
			Port Port `config:"NAMED_PORT"`
		}

		t.Setenv("NAMED_PORT", "9090")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if cfg.Port != 9090 {
			t.Errorf("got %d, want 9090", cfg.Port)
		}
	})

	t.Run("named float type", func(t *testing.T) {
		type Score float64
		type Config struct {
			Score Score `config:"NAMED_SCORE"`
		}

		t.Setenv("NAMED_SCORE", "9.5")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if cfg.Score != 9.5 {
			t.Errorf("got %v, want 9.5", cfg.Score)
		}
	})

	t.Run("uint types", func(t *testing.T) {
		type Config struct {
			U   uint   `config:"UINT"`
			U32 uint32 `config:"UINT32"`
			U64 uint64 `config:"UINT64"`
		}

		t.Setenv("UINT", "1")
		t.Setenv("UINT32", "2")
		t.Setenv("UINT64", "3")

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if cfg.U != 1 || cfg.U32 != 2 || cfg.U64 != 3 {
			t.Errorf("got %v/%v/%v, want 1/2/3", cfg.U, cfg.U32, cfg.U64)
		}
	})

	t.Run("unknown tag option errors", func(t *testing.T) {
		type Config struct {
			Port int `config:"PORT,rquired"`
		}

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrTag) {
			t.Errorf("got %v, want ErrTag", err)
		}
		var mte *configutil.MalformedTagError
		if !errors.As(err, &mte) {
			t.Fatalf("errors.As MalformedTagError failed")
		}
	})

	t.Run("unsupported field type", func(t *testing.T) {
		type Config struct {
			Ch chan int `config:"CHAN"`
		}

		t.Setenv("CHAN", "value")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrUnsupported) {
			t.Errorf("got %v, want ErrUnsupported", err)
		}
	})

	t.Run("unsupported slice element type", func(t *testing.T) {
		type Config struct {
			Flags []bool `config:"BOOL_SLICE"`
		}

		t.Setenv("BOOL_SLICE", "true,false")

		var cfg Config
		err := configutil.Set(&cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrUnsupported) {
			t.Errorf("got %v, want ErrUnsupported", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Set — invalid config types
// ---------------------------------------------------------------------------

func TestSet_InvalidConfig(t *testing.T) {
	t.Run("not a pointer", func(t *testing.T) {
		type Config struct {
			Value string `config:"VALUE"`
		}
		var cfg Config
		err := configutil.Set(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrInvalidConfig) {
			t.Errorf("got %v, want ErrInvalidConfig", err)
		}
	})

	t.Run("pointer to non-struct", func(t *testing.T) {
		var val string
		err := configutil.Set(&val)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrInvalidConfig) {
			t.Errorf("got %v, want ErrInvalidConfig", err)
		}
	})
}

// ---------------------------------------------------------------------------
// WithFilepath
// ---------------------------------------------------------------------------

func TestSet_WithFilepath(t *testing.T) {
	t.Run("loads values from env file", func(t *testing.T) {
		type Config struct {
			Key string `config:"KEY"`
		}

		tmp := t.TempDir() + "/test.env"
		if err := os.WriteFile(tmp, []byte("KEY=from_file\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		var cfg Config
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}
		if cfg.Key != "from_file" {
			t.Errorf("got %q, want %q", cfg.Key, "from_file")
		}
	})

	t.Run("int field from file", func(t *testing.T) {
		type Config struct {
			Port int `config:"PORT"`
		}

		tmp := t.TempDir() + "/test.env"
		if err := os.WriteFile(tmp, []byte("PORT=3000\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		var cfg Config
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}
		if cfg.Port != 3000 {
			t.Errorf("got %d, want 3000", cfg.Port)
		}
	})

	t.Run("default used when file empty", func(t *testing.T) {
		type Config struct {
			Val string `config:"DEFAULT_VAL,default=fallback"`
		}

		tmp := t.TempDir() + "/empty.env"
		if err := os.WriteFile(tmp, []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		var cfg Config
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}
		if cfg.Val != "fallback" {
			t.Errorf("got %q, want %q", cfg.Val, "fallback")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		type Config struct {
			Key string `config:"KEY"`
		}

		var cfg Config
		err := configutil.Set(&cfg, configutil.WithFilepath("/nonexistent.env"))
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrFile) {
			t.Errorf("got %v, want ErrFile", err)
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		type Config struct {
			Key string `config:"KEY"`
		}

		var cfg Config
		err := configutil.Set(&cfg, configutil.WithFilepath("config.txt"))
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, configutil.ErrFile) {
			t.Errorf("got %v, want ErrFile", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Precedence: Default < File < Env < Flag
// ---------------------------------------------------------------------------

func TestPrecedence(t *testing.T) {
	if flag.Lookup("PREC_FLAG") == nil {
		flag.String("PREC_FLAG", "", "test flag for precedence")
	}

	type Config struct {
		Value string `config:"PREC_FLAG,default=default_val"`
	}

	t.Run("default only", func(t *testing.T) {
		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "default_val" {
			t.Errorf("got %q, want %q", cfg.Value, "default_val")
		}
	})

	t.Run("file overwrites default", func(t *testing.T) {
		tmp := t.TempDir() + "/prec.env"
		if err := os.WriteFile(tmp, []byte("PREC_FLAG=file_val\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		var cfg Config
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "file_val" {
			t.Errorf("got %q, want %q", cfg.Value, "file_val")
		}
	})

	t.Run("env overwrites file", func(t *testing.T) {
		tmp := t.TempDir() + "/prec.env"
		if err := os.WriteFile(tmp, []byte("PREC_FLAG=file_val\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		t.Setenv("PREC_FLAG", "env_val")

		var cfg Config
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "env_val" {
			t.Errorf("got %q, want %q", cfg.Value, "env_val")
		}
	})

	t.Run("flag overwrites env", func(t *testing.T) {
		t.Setenv("PREC_FLAG", "env_val")
		if err := flag.Set("PREC_FLAG", "flag_val"); err != nil {
			t.Fatal(err)
		}

		var cfg Config
		if err := configutil.Set(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.Value != "flag_val" {
			t.Errorf("got %q, want %q", cfg.Value, "flag_val")
		}
	})

	t.Run("slice precedence env over file", func(t *testing.T) {
		type SliceCfg struct {
			Values []string `config:"PREC_SLICE"`
		}

		tmp := t.TempDir() + "/prec.env"
		if err := os.WriteFile(tmp, []byte("PREC_SLICE=a,b,c\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		t.Setenv("PREC_SLICE", "env1,env2")

		var cfg SliceCfg
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}

		if !slices.Equal(cfg.Values, []string{"env1", "env2"}) {
			t.Errorf("got %v, want [env1 env2]", cfg.Values)
		}
	})

	t.Run("nested struct precedence", func(t *testing.T) {
		type Nested struct {
			Server struct {
				Port int `config:"PORT"`
			} `config:",prefix=SERVER_"`
		}

		tmp := t.TempDir() + "/prec.env"
		if err := os.WriteFile(tmp, []byte("SERVER_PORT=8080\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		t.Setenv("SERVER_PORT", "9090")

		var cfg Nested
		if err := configutil.Set(&cfg, configutil.WithFilepath(tmp)); err != nil {
			t.Fatal(err)
		}
		if cfg.Server.Port != 9090 {
			t.Errorf("got %d, want 9090", cfg.Server.Port)
		}
	})
}
