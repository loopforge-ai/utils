//go:build integration

package yaml_test

import (
	"strings"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/yaml"
)

func Test_RoundTrip_With_BlockScalarTemplate_Should_PreserveCode(t *testing.T) {
	t.Parallel()
	// Arrange
	type Skill struct {
		Name     string `yaml:"name"`
		Template string `yaml:"template"`
	}
	original := Skill{
		Name:     "codegen",
		Template: "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}",
	}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed Skill
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "name", parsed.Name, original.Name)
	assert.That(t, "template should contain original", strings.Contains(parsed.Template, original.Template), true)
}

func Test_RoundTrip_With_SkillFrontmatter_Should_PreserveAllFields(t *testing.T) {
	t.Parallel()
	// Arrange
	type Arg struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	type Skill struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Version     string   `yaml:"version"`
		Tags        []string `yaml:"tags"`
		Args        []Arg    `yaml:"args"`
	}
	original := Skill{
		Name:        "round_trip_skill",
		Description: "Tests round-trip fidelity",
		Version:     "2.0.0",
		Tags:        []string{"integration", "test"},
		Args: []Arg{
			{Name: "input", Description: "The input file"},
			{Name: "output", Description: "The output file"},
		},
	}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed Skill
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "name", parsed.Name, original.Name)
	assert.That(t, "description", parsed.Description, original.Description)
	assert.That(t, "version", parsed.Version, original.Version)
	assert.That(t, "tags count", len(parsed.Tags), len(original.Tags))
	assert.That(t, "first tag", parsed.Tags[0], original.Tags[0])
	assert.That(t, "args count", len(parsed.Args), len(original.Args))
	assert.That(t, "first arg name", parsed.Args[0].Name, original.Args[0].Name)
	assert.That(t, "first arg desc", parsed.Args[0].Description, original.Args[0].Description)
	assert.That(t, "second arg name", parsed.Args[1].Name, original.Args[1].Name)
}

func Test_Unmarshal_With_BlockScalarTemplate_Should_ParseCode(t *testing.T) {
	t.Parallel()
	// Arrange
	type Skill struct {
		Name     string `yaml:"name"`
		Template string `yaml:"template"`
	}
	input := `name: codegen
template: |
  package {{ .pkg }}

  func main() {
      fmt.Println("Hello")
  }`
	var s Skill

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name", s.Name, "codegen")
	assert.That(t, "template should contain package", strings.Contains(s.Template, "package {{ .pkg }}"), true)
	assert.That(t, "template should contain func main", strings.Contains(s.Template, "func main()"), true)
	assert.That(t, "template should contain Println", strings.Contains(s.Template, `fmt.Println("Hello")`), true)
}

func Test_Unmarshal_With_MixedFieldTypes_Should_ParseCorrectly(t *testing.T) {
	t.Parallel()
	// Arrange
	type Config struct {
		Name        string            `yaml:"name"`
		Debug       bool              `yaml:"debug"`
		MaxRetries  int               `yaml:"max_retries"`
		Timeout     float64           `yaml:"timeout"`
		Tags        []string          `yaml:"tags"`
		Env         map[string]string `yaml:"env"`
		Description string            `yaml:"description"`
	}
	input := `name: my_service
debug: true
max_retries: 5
timeout: 30.5
tags:
  - production
  - critical
env:
  GO_ENV: production
  LOG_LEVEL: info
description: |
  A multi-line description
  that spans several lines
  for testing purposes`
	var c Config

	// Act
	err := yaml.Unmarshal([]byte(input), &c)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name", c.Name, "my_service")
	assert.That(t, "debug", c.Debug, true)
	assert.That(t, "max_retries", c.MaxRetries, 5)
	assert.That(t, "timeout", c.Timeout, 30.5)
	assert.That(t, "tags count", len(c.Tags), 2)
	assert.That(t, "first tag", c.Tags[0], "production")
	assert.That(t, "env count", len(c.Env), 2)
	assert.That(t, "GO_ENV", c.Env["GO_ENV"], "production")
	assert.That(t, "description contains multi-line", strings.Contains(c.Description, "A multi-line description"), true)
	assert.That(t, "description contains testing", strings.Contains(c.Description, "for testing purposes"), true)
}

func Test_Unmarshal_With_SkillFrontmatter_Should_ParseAllFields(t *testing.T) {
	t.Parallel()
	// Arrange
	type Arg struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	type Skill struct {
		Name        string            `yaml:"name"`
		Description string            `yaml:"description"`
		Version     string            `yaml:"version"`
		Tags        []string          `yaml:"tags"`
		Args        []Arg             `yaml:"args"`
		Defaults    map[string]string `yaml:"defaults"`
	}
	input := `name: my_skill
description: A test skill for integration testing
version: 1.0.0
tags: [go, testing, yaml]
args:
  - name: pkg
    description: The package name
  - name: module
    description: The module path
defaults:
  pkg: myapp
  module: example.com/myapp`
	var s Skill

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name", s.Name, "my_skill")
	assert.That(t, "description", s.Description, "A test skill for integration testing")
	assert.That(t, "version", s.Version, "1.0.0")
	assert.That(t, "tags count", len(s.Tags), 3)
	assert.That(t, "first tag", s.Tags[0], "go")
	assert.That(t, "args count", len(s.Args), 2)
	assert.That(t, "first arg name", s.Args[0].Name, "pkg")
	assert.That(t, "first arg desc", s.Args[0].Description, "The package name")
	assert.That(t, "second arg name", s.Args[1].Name, "module")
	assert.That(t, "defaults count", len(s.Defaults), 2)
	assert.That(t, "default pkg", s.Defaults["pkg"], "myapp")
	assert.That(t, "default module", s.Defaults["module"], "example.com/myapp")
}
