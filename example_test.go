package configutil_test

import (
	"fmt"
	"os"

	"github.com/h-dav/configutil"
)

func ExampleSet() {
	type Config struct {
		Value string `config:"VALUE"`
	}

	if err := os.Setenv("VALUE", "value"); err != nil {
		panic(err)
	}
	defer func() { _ = os.Unsetenv("VALUE") }()

	var cfg Config

	if err := configutil.Set(&cfg); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Value)
	// Output:
	// value
}

func ExampleSet_advanced() {
	type Config struct {
		Service string `config:"SERVICE,required"`
		Port    int    `config:"PORT,default=8080"`
	}

	if err := os.Setenv("SERVICE", "auth"); err != nil {
		panic(err)
	}
	defer func() { _ = os.Unsetenv("SERVICE") }()
	// PORT is not set, so it will use the default value.

	var cfg Config

	if err := configutil.Set(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("Service: %s, Port: %d\n", cfg.Service, cfg.Port)
	// Output:
	// Service: auth, Port: 8080
}
