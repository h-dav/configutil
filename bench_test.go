package configutil

import (
	"os"
	"testing"
)

func BenchmarkParseTag_Simple(b *testing.B) {
	for b.Loop() {
		_, _ = parseTag("PORT,required")
	}
}

func BenchmarkParseTag_Complex(b *testing.B) {
	for b.Loop() {
		_, _ = parseTag("PORT,required,default=8080,prefix=API_")
	}
}

func BenchmarkParseTag_Whitespace(b *testing.B) {
	for b.Loop() {
		_, _ = parseTag("  PORT  ,  required  ,  default = 8080  ,  prefix = API_  ")
	}
}

func BenchmarkFileSource_Load(b *testing.B) {
	tmp := b.TempDir() + "/bench.env"
	content := "KEY1=VAL1\nKEY2=VAL2\nKEY3=VAL3\n# Comment\nKEY4=VAL4 # inline comment\n"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		b.Fatal(err)
	}

	src := fileSource{filepath: tmp}
	for b.Loop() {
		_, _ = src.Load()
	}
}

func BenchmarkEnvSource_Load(b *testing.B) {
	src := environmentVariableSource{}
	for b.Loop() {
		_, _ = src.Load()
	}
}

func BenchmarkSet_Small(b *testing.B) {
	type Config struct {
		Key string `config:"SMALL_KEY"`
	}
	b.Setenv("SMALL_KEY", "value")

	var cfg Config
	for b.Loop() {
		_ = Set(&cfg)
	}
}

func BenchmarkSet_Medium(b *testing.B) {
	type Config struct {
		K1 string `config:"K1"`
		K2 int    `config:"K2"`
		K3 bool   `config:"K3"`
		K4 string `config:"K4,default=def"`
		K5 []int  `config:"K5"`
	}
	b.Setenv("K1", "v1")
	b.Setenv("K2", "123")
	b.Setenv("K3", "true")
	b.Setenv("K5", "1,2,3")

	var cfg Config
	for b.Loop() {
		_ = Set(&cfg)
	}
}

func BenchmarkSet_Large(b *testing.B) {
	type Nested struct {
		N1 string `config:"N1"`
		N2 int    `config:"N2"`
	}
	type Config struct {
		K1  string   `config:"K1"`
		K2  int      `config:"K2"`
		K3  bool     `config:"K3"`
		K4  string   `config:"K4,default=def"`
		K5  []int    `config:"K5"`
		K6  float64  `config:"K6"`
		K7  string   `config:"K7,required"`
		K8  Nested   `config:",prefix=PRE_"`
		K9  []string `config:"K9"`
		K10 string   `config:"K10"`
	}
	b.Setenv("K1", "v1")
	b.Setenv("K2", "123")
	b.Setenv("K3", "true")
	b.Setenv("K5", "1,2,3")
	b.Setenv("K6", "1.23")
	b.Setenv("K7", "req")
	b.Setenv("PRE_N1", "nv1")
	b.Setenv("PRE_N2", "456")
	b.Setenv("K9", "a,b,c")
	b.Setenv("K10", "v10")

	var cfg Config
	for b.Loop() {
		_ = Set(&cfg)
	}
}
