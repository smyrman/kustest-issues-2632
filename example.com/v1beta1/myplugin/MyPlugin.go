package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type plugin struct {
	Metadata interface{} `yaml:"metadata"`
	Data     struct {
		Username string `yaml:"username"`
	} `yaml:"data"`
}

func (d *plugin) Generate(ctx context.Context) (interface{}, error) {
	return map[string]interface{}{
		"kind":       "ConfigMap",
		"apiVersion": "v1",
		"metadata":   d.Metadata, // pass-through.
		"data": map[string]string{
			"username": d.Data.Username,
		},
	}, nil
}

func main() {
	GeneratorMain(&plugin{})
}

// Would usually live in a shared package; copied in for example.

func init() {
	log.SetFlags(0)
}

// Generator describes the interface for a generator plugin.
type Generator interface {
	// Generate returns a type that can be YAML encoded to a Kubernetes
	// manifest.
	Generate(ctx context.Context) (interface{}, error)
}

// GeneratorMain runs a generator plugin by parsing the command-line argument
// and exiting with a non-zero exit code on error. plugin must be a pointer
// type.
func GeneratorMain(plugin Generator) {
	flag.Parse()
	if n := flag.NArg(); n != 1 {
		log.Printf("expecting exactly one argument, got %d", n)
		os.Exit(2)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Printf("open input: %s", err)
		os.Exit(1)
	}
	defer f.Close()
	if err := RunGenerator(os.Stdout, f, plugin); err != nil {
		log.Printf("generate: %s", err)
		os.Exit(1)
	}
}

// RunGenerator runs a generator plugin with r as input, expecting output to be
// written to w. plugin must be a pointer type.
func RunGenerator(w io.Writer, r io.Reader, plugin Generator) error {
	dec := yaml.NewDecoder(r)
	//dec.KnownFields(true) // Skip for reproduction test.
	if err := dec.Decode(plugin); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	target, err := plugin.Generate(context.Background())
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	if err := enc.Encode(target); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	return nil
}
