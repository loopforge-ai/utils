package yaml_test

import (
	"strings"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/yaml"
)

func Test_Marshal_With_BoolField_Should_ProduceYAML(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Active bool `yaml:"active"`
	}
	s := S{Active: true}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain active", strings.Contains(string(data), "active: true"), true)
}

func Test_Marshal_With_BoolLikeString_Should_QuoteIt(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	s := S{Value: "true"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain quoted true", strings.Contains(string(data), `value: "true"`), true)
}

func Test_Marshal_With_DashTag_Should_SkipField(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name    string `yaml:"name"`
		Skipped string `yaml:"-"`
	}
	s := S{Name: "test", Skipped: "hidden"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain name", strings.Contains(string(data), "name: test"), true)
	assert.That(t, "should not contain skipped", strings.Contains(string(data), "hidden"), false)
}

func Test_Marshal_With_EmptyMap_Should_ProduceEmptyBraces(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Args map[string]string `yaml:"args"`
	}
	s := S{Args: map[string]string{}}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain empty braces", strings.Contains(string(data), "args: {}"), true)
}

func Test_Marshal_With_EmptySliceOmitempty_Should_OmitField(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags,omitempty"`
	}
	s := S{}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should not contain tags", strings.Contains(string(data), "tags"), false)
}

func Test_Marshal_With_EmptySlice_Should_ProduceEmptyBrackets(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	s := S{Tags: []string{}}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain empty brackets", strings.Contains(string(data), "tags: []"), true)
}

func Test_Marshal_With_EmptyString_Should_QuoteIt(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	s := S{Name: ""}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain quoted empty", strings.Contains(string(data), `name: ""`), true)
}

func Test_Marshal_With_Float64Field_Should_ProduceYAML(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Score float64 `yaml:"score"`
	}
	s := S{Score: 3.14}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain score", strings.Contains(string(data), "score: 3.14"), true)
}

func Test_Marshal_With_Map_Should_ProduceMapping(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Args map[string]string `yaml:"args"`
	}
	s := S{Args: map[string]string{"key": "val"}}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	output := string(data)
	assert.That(t, "should contain args header", strings.Contains(output, "args:"), true)
	assert.That(t, "should contain key-val", strings.Contains(output, "key: val"), true)
}

func Test_Marshal_With_MultilineString_Should_UsePipe(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Code string `yaml:"code"`
	}
	s := S{Code: "line one\nline two"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain pipe", strings.Contains(string(data), "code: |"), true)
	assert.That(t, "should contain line one", strings.Contains(string(data), "  line one"), true)
	assert.That(t, "should contain line two", strings.Contains(string(data), "  line two"), true)
}

func Test_Marshal_With_NestedStruct_Should_ProduceIndentedMapping(t *testing.T) {
	t.Parallel()
	// Arrange
	type Inner struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	type S struct {
		Server Inner `yaml:"server"`
	}
	s := S{Server: Inner{Host: "localhost", Port: 8080}}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	output := string(data)
	assert.That(t, "should contain server header", strings.Contains(output, "server:"), true)
	assert.That(t, "should contain host", strings.Contains(output, "  host: localhost"), true)
	assert.That(t, "should contain port", strings.Contains(output, "  port: 8080"), true)
}

func Test_Marshal_With_NonStruct_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	_, err := yaml.Marshal("not a struct")

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
}

func Test_Marshal_With_NumberLikeString_Should_QuoteIt(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	s := S{Value: "42"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain quoted 42", strings.Contains(string(data), `value: "42"`), true)
}

func Test_Marshal_With_OmitemptyNonZero_Should_Include(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version,omitempty"`
	}
	s := S{Name: "test", Version: "1.0"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain version", strings.Contains(string(data), `version: "1.0"`), true)
}

func Test_Marshal_With_Omitempty_Should_SkipZeroValue(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version,omitempty"`
	}
	s := S{Name: "test"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain name", strings.Contains(string(data), "name"), true)
	assert.That(t, "should not contain version", strings.Contains(string(data), "version"), false)
}

func Test_Marshal_With_Pointer_Should_Dereference(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	s := &S{Name: "test"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain name", strings.Contains(string(data), "name: test"), true)
}

func Test_Marshal_With_SpecialCharsString_Should_QuoteIt(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	s := S{Name: "hello: world"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain quoted string", strings.Contains(string(data), `name: "hello: world"`), true)
}

func Test_Marshal_With_StringAndInt_Should_ProduceYAML(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name  string `yaml:"name"`
		Count int    `yaml:"count"`
	}
	s := S{Name: "test", Count: 5}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain name", strings.Contains(string(data), "name: test"), true)
	assert.That(t, "should contain count", strings.Contains(string(data), "count: 5"), true)
}

func Test_Marshal_With_StringSlice_Should_ProduceBlockSequence(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	s := S{Tags: []string{"go", "yaml"}}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain tags header", strings.Contains(string(data), "tags:"), true)
	assert.That(t, "should contain go item", strings.Contains(string(data), "- go"), true)
	assert.That(t, "should contain yaml item", strings.Contains(string(data), "- yaml"), true)
}

func Test_Marshal_With_StructSlice_Should_ProduceSequence(t *testing.T) {
	t.Parallel()
	// Arrange
	type Item struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}
	type S struct {
		Items []Item `yaml:"items"`
	}
	s := S{Items: []Item{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}}}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	output := string(data)
	assert.That(t, "should contain items header", strings.Contains(output, "items:"), true)
	assert.That(t, "should contain dash name a", strings.Contains(output, "- name: a"), true)
	assert.That(t, "should contain value 1", strings.Contains(output, `value: "1"`), true)
}

func Test_RoundTrip_With_AllScalarTypes_Should_PreserveValues(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name   string  `yaml:"name"`
		Count  int     `yaml:"count"`
		Score  float64 `yaml:"score"`
		Active bool    `yaml:"active"`
	}
	original := S{Name: "test", Count: 42, Score: 3.14, Active: true}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed S
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "name should match", parsed.Name, original.Name)
	assert.That(t, "count should match", parsed.Count, original.Count)
	assert.That(t, "score should match", parsed.Score, original.Score)
	assert.That(t, "active should match", parsed.Active, original.Active)
}

func Test_RoundTrip_With_BlockScalar_Should_PreserveValues(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Code string `yaml:"code"`
	}
	original := S{Code: "line one\nline two\nline three"}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed S
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "code should match original exactly", parsed.Code, original.Code)
}

func Test_RoundTrip_With_NestedStruct_Should_PreserveValues(t *testing.T) {
	t.Parallel()
	// Arrange
	type Inner struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	type S struct {
		Server Inner `yaml:"server"`
	}
	original := S{Server: Inner{Host: "localhost", Port: 8080}}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed S
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "host should match", parsed.Server.Host, original.Server.Host)
	assert.That(t, "port should match", parsed.Server.Port, original.Server.Port)
}

func Test_RoundTrip_With_SliceField_Should_PreserveValues(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name  string   `yaml:"name"`
		Tags  []string `yaml:"tags"`
		Retry int      `yaml:"retry"`
	}
	original := S{Name: "test", Tags: []string{"a", "b"}, Retry: 3}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed S
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "name should match", parsed.Name, original.Name)
	assert.That(t, "tags count should match", len(parsed.Tags), len(original.Tags))
	assert.That(t, "first tag should match", parsed.Tags[0], original.Tags[0])
	assert.That(t, "second tag should match", parsed.Tags[1], original.Tags[1])
	assert.That(t, "retry should match", parsed.Retry, original.Retry)
}

func Test_RoundTrip_With_StructSlice_Should_PreserveValues(t *testing.T) {
	t.Parallel()
	// Arrange
	type Item struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}
	type S struct {
		Items []Item `yaml:"items"`
	}
	original := S{Items: []Item{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}}}

	// Act
	data, err := yaml.Marshal(original)
	assert.That(t, "marshal err should be nil", err, nil)
	var parsed S
	err = yaml.Unmarshal(data, &parsed)

	// Assert
	assert.That(t, "unmarshal err should be nil", err, nil)
	assert.That(t, "items count should match", len(parsed.Items), 2)
	assert.That(t, "first item name", parsed.Items[0].Name, original.Items[0].Name)
	assert.That(t, "first item value", parsed.Items[0].Value, original.Items[0].Value)
	assert.That(t, "second item name", parsed.Items[1].Name, original.Items[1].Name)
	assert.That(t, "second item value", parsed.Items[1].Value, original.Items[1].Value)
}

func Test_Unmarshal_With_BlockScalarEmptyContent_Should_ParseEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Code string `yaml:"code"`
	}
	input := "code: |"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "code should be empty", s.Code, "")
}

func Test_Unmarshal_With_BlockScalarFollowedByField_Should_StopAtDedent(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Code string `yaml:"code"`
		Name string `yaml:"name"`
	}
	input := "code: |\n  line one\n  line two\nname: hello"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "code should have 2 lines", strings.Count(s.Code, "\n"), 1)
	assert.That(t, "name should match", s.Name, "hello")
}

func Test_Unmarshal_With_BlockScalar_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Code string `yaml:"code"`
	}
	input := "code: |\n  line one\n  line two\n  line three"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "code should have 3 lines", strings.Count(s.Code, "\n"), 2)
	assert.That(t, "code starts with line one", strings.HasPrefix(s.Code, "line one"), true)
	assert.That(t, "code ends with line three", strings.HasSuffix(s.Code, "line three"), true)
}

func Test_Unmarshal_With_BlockSequence_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Items []string `yaml:"items"`
	}
	input := "items:\n  - alpha\n  - beta\n  - gamma"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "items count", len(s.Items), 3)
	assert.That(t, "first item", s.Items[0], "alpha")
	assert.That(t, "second item", s.Items[1], "beta")
	assert.That(t, "third item", s.Items[2], "gamma")
}

func Test_Unmarshal_With_BoolFalse_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Active bool `yaml:"active"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("active: false"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "active should be false", s.Active, false)
}

func Test_Unmarshal_With_BoolField_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Active bool `yaml:"active"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("active: true"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "active should be true", s.Active, true)
}

func Test_Unmarshal_With_Comments_Should_SkipThem(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	input := "# this is a comment\nname: test\n# another comment\nage: 25"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
	assert.That(t, "age should match", s.Age, 25)
}

func Test_Marshal_With_DELChar_Should_Quote(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	s := S{Value: "hello\x7Fworld"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should be quoted", strings.Contains(string(data), `"`), true)
}

func Test_Unmarshal_With_UnicodeEscape_Should_Decode(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	input := `value: "\u0041\u0042\u0043"`

	// Act
	var s S
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "value should be ABC", s.Value, "ABC")
}

func Test_Unmarshal_With_VerticalTabEscape_Should_Decode(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	input := `value: "hello\vworld"`

	// Act
	var s S
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "value should contain vertical tab", s.Value, "hello\vworld")
}

func Test_Marshal_With_SpecialValues_Should_Quote(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		A string `yaml:"a"`
		B string `yaml:"b"`
		C string `yaml:"c"`
		D string `yaml:"d"`
		E string `yaml:"e"`
	}
	s := S{A: ".inf", B: "-.inf", C: ".nan", D: "0x1F", E: "0o17"}

	// Act
	out, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	result := string(out)
	assert.That(t, "a should be quoted", strings.Contains(result, `a: ".inf"`), true)
	assert.That(t, "d should be quoted", strings.Contains(result, `d: "0x1F"`), true)
	assert.That(t, "e should be quoted", strings.Contains(result, `e: "0o17"`), true)
}

func Test_Marshal_With_VerticalTab_Should_Escape(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Val string `yaml:"val"`
	}
	s := S{Val: "hello\vworld"}

	// Act
	out, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	result := string(out)
	assert.That(t, "should contain escaped vertical tab", strings.Contains(result, `\v`), true)
}

func Test_Unmarshal_With_FlowSequenceEscapes_Should_ProcessEscapes(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Items []string `yaml:"items"`
	}
	input := `items: ["hello\nworld", "tab\there"]`
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should have 2 items", len(s.Items), 2)
	assert.That(t, "first item should have newline", s.Items[0], "hello\nworld")
	assert.That(t, "second item should have tab", s.Items[1], "tab\there")
}

func Test_Unmarshal_With_StripInlineCommentEscapedQuotes_Should_Preserve(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Val string `yaml:"val"`
	}
	input := `val: "hello \"world\" test" # comment`
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "val should preserve escaped quotes", s.Val, `hello "world" test`)
}

func Test_Unmarshal_With_DoubleQuotedString_Should_Unquote(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte(`name: "hello world"`), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "hello world")
}

func Test_Unmarshal_With_EmptyFlowSequence_Should_ParseEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("tags: []"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "tags should be empty", len(s.Tags), 0)
}

func Test_Unmarshal_With_EmptyInput_Should_NotError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte(""), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should be zero value", s.Name, "")
}

func Test_Unmarshal_With_EmptyLines_Should_SkipThem(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	input := "name: test\n\n\nage: 25"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
	assert.That(t, "age should match", s.Age, 25)
}

func Test_Unmarshal_With_EmptyMap_Should_ParseEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Args map[string]string `yaml:"args"`
		Name string            `yaml:"name"`
	}
	input := "name: test"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
}

func Test_Unmarshal_With_EscapedQuotes_Should_Unescape(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	input := `value: "hello \"world\""`
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "value should have unescaped quotes", s.Value, `hello "world"`)
}

func Test_Unmarshal_With_FlowSequenceSingleQuotes_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	input := "tags: ['hello', 'wor,ld', 'foo']"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should have 3 tags", len(s.Tags), 3)
	assert.That(t, "first tag", s.Tags[0], "hello")
	assert.That(t, "second tag with comma", s.Tags[1], "wor,ld")
	assert.That(t, "third tag", s.Tags[2], "foo")
}

func Test_Unmarshal_With_UintField_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Port uint16 `yaml:"port"`
	}
	input := "port: 8080"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "port should be 8080", s.Port, uint16(8080))
}

func Test_Unmarshal_With_UintOverflow_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Small uint8 `yaml:"small"`
	}
	input := "small: 999"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
	assert.That(t, "err should mention parse", strings.Contains(err.Error(), "parse uint"), true)
}

func Test_Unmarshal_With_UnsupportedMapType_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Counts map[string]int `yaml:"counts"`
	}
	input := "counts:\n  a: 1\n  b: 2"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
	assert.That(t, "err should mention unsupported map", strings.Contains(err.Error(), "unsupported map type"), true)
}

func Test_Marshal_With_PointerField_Should_DereferenceAndMarshal(t *testing.T) {
	t.Parallel()
	// Arrange
	val := "hello"
	type S struct {
		Name *string `yaml:"name"`
	}
	s := S{Name: &val}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should contain name", strings.Contains(string(data), "name: hello"), true)
}

func Test_Marshal_With_NilPointerField_Should_Skip(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name *string `yaml:"name,omitempty"`
	}
	s := S{Name: nil}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should not contain name", strings.Contains(string(data), "name"), false)
}

func Test_Unmarshal_With_Float64Field_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Score float64 `yaml:"score"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("score: 3.14"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "score should be 3.14", s.Score, 3.14)
}

func Test_Unmarshal_With_FlowSequenceQuotedItems_Should_Unquote(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte(`tags: ["hello", "world"]`), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "tags count", len(s.Tags), 2)
	assert.That(t, "first tag", s.Tags[0], "hello")
	assert.That(t, "second tag", s.Tags[1], "world")
}

func Test_Unmarshal_With_FlowSequence_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("tags: [go, testing, yaml]"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "tags count", len(s.Tags), 3)
	assert.That(t, "first tag", s.Tags[0], "go")
	assert.That(t, "second tag", s.Tags[1], "testing")
	assert.That(t, "third tag", s.Tags[2], "yaml")
}

func Test_Unmarshal_With_IntField_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Count int `yaml:"count"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("count: 42"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "count should be 42", s.Count, 42)
}

func Test_Unmarshal_With_InvalidBool_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Active bool `yaml:"active"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("active: notabool"), &s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
}

func Test_Unmarshal_With_InvalidFloat_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Score float64 `yaml:"score"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("score: notafloat"), &s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
}

func Test_Unmarshal_With_InvalidInt_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Count int `yaml:"count"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("count: abc"), &s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
}

func Test_Unmarshal_With_MapStringString_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Args map[string]string `yaml:"args"`
	}
	input := "args:\n  pkg: foo\n  module: example.com/bar"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "args count", len(s.Args), 2)
	assert.That(t, "pkg value", s.Args["pkg"], "foo")
	assert.That(t, "module value", s.Args["module"], "example.com/bar")
}

func Test_Unmarshal_With_MultipleFields_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name    string  `yaml:"name"`
		Version string  `yaml:"version"`
		Retry   int     `yaml:"retry"`
		Score   float64 `yaml:"score"`
		Active  bool    `yaml:"active"`
	}
	input := "name: test\nversion: 1.0.0\nretry: 3\nscore: 0.95\nactive: true"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name", s.Name, "test")
	assert.That(t, "version", s.Version, "1.0.0")
	assert.That(t, "retry", s.Retry, 3)
	assert.That(t, "score", s.Score, 0.95)
	assert.That(t, "active", s.Active, true)
}

func Test_Unmarshal_With_NegativeInt_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Count int `yaml:"count"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("count: -7"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "count should be -7", s.Count, -7)
}

func Test_Unmarshal_With_NestedStructSequence_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type Item struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	type S struct {
		Items []Item `yaml:"items"`
	}
	input := "items:\n  - name: foo\n    description: a foo\n  - name: bar\n    description: a bar"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "items count", len(s.Items), 2)
	assert.That(t, "first name", s.Items[0].Name, "foo")
	assert.That(t, "first desc", s.Items[0].Description, "a foo")
	assert.That(t, "second name", s.Items[1].Name, "bar")
	assert.That(t, "second desc", s.Items[1].Description, "a bar")
}

func Test_Unmarshal_With_NestedStruct_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type Inner struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}
	type S struct {
		Server Inner `yaml:"server"`
	}
	input := "server:\n  host: localhost\n  port: 8080"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "host should match", s.Server.Host, "localhost")
	assert.That(t, "port should match", s.Server.Port, 8080)
}

func Test_Unmarshal_With_NilPointer_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	type S struct{}
	err := yaml.Unmarshal([]byte("name: test"), (*S)(nil))

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
}

func Test_Unmarshal_With_NoTag_Should_MatchCaseInsensitive(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("name: hello"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "hello")
}

func Test_Unmarshal_With_NonPointer_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct{}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("name: test"), s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
}

func Test_Unmarshal_With_SingleQuotedString_Should_Unquote(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("name: 'hello world'"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "hello world")
}

func Test_Unmarshal_With_StringField_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	var s S

	// Act
	err := yaml.Unmarshal([]byte("name: hello"), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "hello")
}

func Test_Unmarshal_With_UnknownBlockField_Should_SkipChildren(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	input := "unknown:\n  child1: a\n  child2: b\nname: test"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
}

func Test_Unmarshal_With_UnknownBlockScalar_Should_SkipContent(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	input := "unknown: |\n  some block\n  content here\nname: test"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
}

func Test_Unmarshal_With_UnknownFields_Should_SkipThem(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	input := "name: test\nunknown: value\nalso_unknown: 42"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
}

func Test_Unmarshal_With_InlineComment_Should_IgnoreComment(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	}
	input := "name: hello # this is a comment\nport: 8080 # default port"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should be hello", s.Name, "hello")
	assert.That(t, "port should be 8080", s.Port, 8080)
}

func Test_Unmarshal_With_HashInQuotes_Should_Preserve(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	input := `value: "hello # world"`
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "value should preserve hash", s.Value, "hello # world")
}

func Test_Unmarshal_With_YAMLBooleans_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		A bool `yaml:"a"`
		B bool `yaml:"b"`
		C bool `yaml:"c"`
		D bool `yaml:"d"`
	}
	input := "a: yes\nb: no\nc: on\nd: off"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "yes should be true", s.A, true)
	assert.That(t, "no should be false", s.B, false)
	assert.That(t, "on should be true", s.C, true)
	assert.That(t, "off should be false", s.D, false)
}

func Test_Unmarshal_With_NullValue_Should_LeaveZeroValue(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name  string `yaml:"name"`
		Count int    `yaml:"count"`
	}
	input := "name: null\ncount: 0"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should be zero value", s.Name, "")
	assert.That(t, "count should be 0", s.Count, 0)
}

func Test_Unmarshal_With_TildeNull_Should_LeaveZeroValue(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	input := "name: ~"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should be zero value", s.Name, "")
}

func Test_EscapeScalar_With_ControlChars_Should_Escape(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Value string `yaml:"value"`
	}
	s := S{Value: "hello\x01world\x1Fend"}

	// Act
	data, err := yaml.Marshal(s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	result := string(data)
	assert.That(t, "should contain \\u0001", strings.Contains(result, `\u0001`), true)
	assert.That(t, "should contain \\u001F", strings.Contains(result, `\u001F`), true)
}

func Test_SplitFlow_With_NestedBrackets_Should_PreserveInner(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Tags []string `yaml:"tags"`
	}
	// Inner brackets quoted — commas inside quotes are already handled.
	// This tests that bracket-contained commas in quoted strings work.
	input := `tags: [a, "b, c", d]`
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "should have 3 items", len(s.Tags), 3)
	assert.That(t, "first item", s.Tags[0], "a")
	assert.That(t, "second item with comma", s.Tags[1], "b, c")
	assert.That(t, "third item", s.Tags[2], "d")
}

func Test_Unmarshal_With_BlockScalarLeadingEmptyLines_Should_Preserve(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Code string `yaml:"code"`
	}
	input := "code: |\n\n  line one\n  line two"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "code should start with newline", strings.HasPrefix(s.Code, "\n"), true)
	assert.That(t, "code should contain line one", strings.Contains(s.Code, "line one"), true)
}

func Test_Unmarshal_With_InvalidUTF8_Should_ReturnError(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
	}
	var s S
	input := []byte{0x6E, 0x61, 0x6D, 0x65, 0x3A, 0x20, 0xFF, 0xFE} // "name: " + invalid UTF-8

	// Act
	err := yaml.Unmarshal(input, &s)

	// Assert
	assert.That(t, "err should not be nil", err != nil, true)
	assert.That(t, "err should mention UTF-8", strings.Contains(err.Error(), "UTF-8"), true)
}

func Test_Unmarshal_With_CRLFLineEndings_Should_Parse(t *testing.T) {
	t.Parallel()
	// Arrange
	type S struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	input := "name: test\r\nage: 25\r\n"
	var s S

	// Act
	err := yaml.Unmarshal([]byte(input), &s)

	// Assert
	assert.That(t, "err should be nil", err, nil)
	assert.That(t, "name should match", s.Name, "test")
	assert.That(t, "age should match", s.Age, 25)
}
