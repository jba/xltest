// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that
// can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

func main() {
	flag.Parse()
	if err := validateFiles(flag.Args()); err != nil {
		log.Fatal(err)
	}
}

const schemaFile = "../../../../test-schema.yaml"

func validateFiles(filenames []string) error {
	schema, err := readTestSchema(schemaFile)
	if err != nil {
		return err
	}

	var errs []error
	for _, filename := range filenames {
		errs = append(errs, validateFile(schema, filename))
	}
	return errors.Join(errs...)
}

func readTestSchema(filename string) (*jsonschema.Schema, error) {
	yamlData, err := os.ReadFile(filepath.FromSlash(filename))
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(yamlData, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return jsonschema.CompileString(filename, string(jsonData))
}

func validateFile(schema *jsonschema.Schema, filename string) error {
	m, err := readYAMLFile(filename)
	if err != nil {
		return err
	}
	if err := schema.Validate(m); err != nil {
		return fmt.Errorf("%s: %v", filename, err)
	}
	return nil
}

func readYAMLFile(filename string) (map[string]any, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return m, nil
}
