// Package yaml provides a minimal YAML parser and serializer.
// It supports the subset of YAML used by skill frontmatter:
// scalars (string, bool, int, float64), block and flow sequences,
// block mappings, nested structs, map[string]string, and block scalars (|).
package yaml

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxNestingDepth is the maximum allowed recursion depth during unmarshal.
const maxNestingDepth = 100

var (
	// escapeMap maps single-character YAML escape codes to their byte values.
	escapeMap = map[byte]byte{
		'0':  0,
		'"':  '"',
		'\\': '\\',
		'b':  '\b',
		'f':  '\f',
		'n':  '\n',
		'r':  '\r',
		't':  '\t',
		'v':  '\v',
	}

	// reverseEscapeMap maps bytes to their YAML escape sequences for serialization.
	reverseEscapeMap = map[byte]string{
		0:    `\0`,
		'\b': `\b`,
		'\f': `\f`,
		'\n': `\n`,
		'\r': `\r`,
		'\t': `\t`,
		'\v': `\v`,
		'"':  `\"`,
		'\\': `\\`,
	}

	// yamlSpecialChars contains characters that require quoting in YAML scalars.
	yamlSpecialChars = [256]bool{
		'!': true, '"': true, '#': true, '%': true, '&': true, '\'': true,
		'*': true, ',': true, '-': true, ':': true, '>': true, '?': true,
		'@': true, '[': true, ']': true, '`': true, '{': true, '|': true,
		'}': true,
	}
)

// Marshal serializes v into YAML bytes.
// v must be a struct or pointer to a struct.
func Marshal(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, errors.New("yaml: Marshal requires a struct")
	}

	var b strings.Builder
	marshalStruct(&b, rv, 0)
	return []byte(b.String()), nil
}

// Unmarshal parses YAML data and stores the result in the value pointed to by v.
// v must be a pointer to a struct or map[string]string.
func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("yaml: Unmarshal requires a non-nil pointer")
	}
	if !utf8.Valid(data) {
		return errors.New("yaml: input is not valid UTF-8")
	}

	normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	if _, err := unmarshalMapping(lines, 0, rv.Elem(), 0); err != nil {
		return fmt.Errorf("yaml: unmarshal: %w", err)
	}
	return nil
}

// --- Unmarshal internals ---

// unmarshalMapping parses block mapping lines starting at index i with the given
// base indent, setting fields on dst. Returns the next line index to process.
func unmarshalMapping(lines []string, i int, dst reflect.Value, depth int) (int, error) {
	if depth > maxNestingDepth {
		return i, errors.New("yaml: maximum nesting depth exceeded")
	}
	dst = derefPointer(dst)
	baseIndent := -1

	for i < len(lines) {
		line := lines[i]

		// Skip empty lines and comments.
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			i++
			continue
		}

		indent := lineIndent(line)

		// Set base indent from first mapping line.
		if baseIndent < 0 {
			baseIndent = indent
		}

		// If dedented, we're done with this mapping.
		if indent < baseIndent {
			return i, nil
		}

		var err error
		i, err = unmarshalMappingLine(lines, i, indent, trimmed, dst, depth)
		if err != nil {
			return i, err
		}
	}

	return i, nil
}

// derefPointer dereferences a pointer value, allocating if nil.
func derefPointer(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// unmarshalMappingLine processes a single non-empty, non-comment mapping line.
func unmarshalMappingLine(lines []string, i int, indent int, trimmed string, dst reflect.Value, depth int) (int, error) {
	colonIdx := findKeySeparator(trimmed)
	if colonIdx < 0 {
		return i + 1, nil
	}

	key := trimmed[:colonIdx]
	rest := ""
	if colonIdx+1 < len(trimmed) {
		rest = strings.TrimSpace(trimmed[colonIdx+1:])
	}

	rest = stripInlineComment(rest)

	field, fieldType, found := findField(dst, key)
	if !found {
		return skipUnknownKey(lines, i, indent, rest), nil
	}

	return unmarshalField(lines, i, indent, rest, key, field, fieldType, depth)
}

// skipUnknownKey advances past an unknown mapping key.
func skipUnknownKey(lines []string, i int, indent int, rest string) int {
	if rest == "" || rest == "|" {
		// Block value or block scalar — skip children.
		i++
		for i < len(lines) {
			t := strings.TrimSpace(lines[i])
			if t == "" {
				i++
				continue
			}
			if lineIndent(lines[i]) <= indent {
				break
			}
			i++
		}
	} else {
		i++
	}
	return i
}

// unmarshalField handles setting a single field value during unmarshal.
func unmarshalField(lines []string, i int, indent int, rest string, key string, field reflect.Value, fieldType reflect.Type, depth int) (int, error) {
	// Handle block scalar (|).
	if rest == "|" {
		i++
		val, next := parseBlockScalar(lines, i, indent)
		if field.IsValid() && field.Kind() == reflect.String {
			field.SetString(val)
		}
		return next, nil
	}

	// Handle empty value — block sequence or nested mapping.
	if rest == "" {
		return unmarshalBlockChild(lines, i+1, indent, key, field, fieldType, depth)
	}

	// Null literal — leave field at zero value.
	if isNullLiteral(rest) {
		return i + 1, nil
	}

	// Inline flow sequence: [a, b, c].
	if strings.HasPrefix(rest, "[") && strings.HasSuffix(rest, "]") {
		inner := rest[1 : len(rest)-1]
		items := splitFlow(inner)
		if err := setSliceFromStrings(field, fieldType, items); err != nil {
			return i, fmt.Errorf("yaml: field %q: %w", key, err)
		}
		return i + 1, nil
	}

	// Scalar value.
	if err := setScalar(field, rest); err != nil {
		return i, fmt.Errorf("yaml: field %q: %w", key, err)
	}
	return i + 1, nil
}

// unmarshalBlockChild handles block sequences and nested mappings for an empty-value key.
func unmarshalBlockChild(lines []string, i int, indent int, key string, field reflect.Value, fieldType reflect.Type, depth int) (int, error) {
	if i >= len(lines) {
		return i, nil
	}
	nextTrimmed := strings.TrimSpace(lines[i])
	nextIndent := lineIndent(lines[i])
	if nextIndent <= indent {
		return i, nil
	}
	if strings.HasPrefix(nextTrimmed, "- ") {
		next, err := parseBlockSequence(lines, i, field, fieldType)
		if err != nil {
			return i, fmt.Errorf("yaml: field %q: %w", key, err)
		}
		return next, nil
	}
	next, err := unmarshalNested(lines, i, field, fieldType, depth+1)
	if err != nil {
		return i, fmt.Errorf("yaml: field %q: %w", key, err)
	}
	return next, nil
}

// parseBlockScalar collects a literal block scalar (|) starting at line i.
func parseBlockScalar(lines []string, i int, parentIndent int) (string, int) {
	if i >= len(lines) {
		return "", i
	}

	blockIndent := -1
	var b strings.Builder
	first := true

	for i < len(lines) {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			i = handleEmptyBlockLine(&b, &first, blockIndent, i)
			continue
		}

		ind := lineIndent(line)
		if blockIndent < 0 {
			if ind <= parentIndent {
				break
			}
			blockIndent = ind
		}

		if ind < blockIndent {
			break
		}

		if !first {
			b.WriteByte('\n')
		}
		first = false
		if len(line) > blockIndent {
			b.WriteString(line[blockIndent:])
		}
		i++
	}

	return strings.TrimRight(b.String(), "\n"), i
}

// handleEmptyBlockLine handles empty lines within a block scalar.
func handleEmptyBlockLine(b *strings.Builder, first *bool, _ int, i int) int {
	if !*first {
		b.WriteByte('\n')
	}
	*first = false
	return i + 1
}

// parseBlockSequence parses a block sequence (lines starting with "- ").
func parseBlockSequence(lines []string, i int, field reflect.Value, ft reflect.Type) (int, error) { //nolint:cyclop // natural loop complexity from error handling
	if i >= len(lines) {
		return i, nil
	}

	seqIndent := lineIndent(lines[i])
	elemType := ft.Elem()

	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}

		ind := lineIndent(lines[i])
		if ind < seqIndent {
			break
		}
		if ind > seqIndent {
			i++
			continue
		}

		if !strings.HasPrefix(trimmed, "- ") {
			break
		}

		itemStr := strings.TrimSpace(trimmed[2:])

		switch elemType.Kind() { //nolint:exhaustive // only string and struct sequences supported
		case reflect.String:
			field.Set(reflect.Append(field, reflect.ValueOf(unquote(itemStr))))
			i++
		case reflect.Struct:
			var err error
			i, err = parseSequenceStruct(lines, i, seqIndent, itemStr, elemType, &field)
			if err != nil {
				return i, err
			}
		default:
			i++
		}
	}

	return i, nil
}

// setKeyValue extracts a key: value pair from s using findKeySeparator and sets the
// corresponding field on elem. Returns an error if the scalar assignment fails.
func setKeyValue(elem reflect.Value, s string) error {
	ci := findKeySeparator(s)
	if ci < 0 {
		return nil
	}
	k := s[:ci]
	v := ""
	if ci+1 < len(s) {
		v = strings.TrimSpace(s[ci+1:])
	}
	v = stripInlineComment(v)
	f, _, found := findField(elem, k)
	if !found {
		return nil
	}
	return setScalar(f, v)
}

func parseSequenceStruct(lines []string, i int, seqIndent int, itemStr string, elemType reflect.Type, field *reflect.Value) (int, error) {
	elem := reflect.New(elemType).Elem()

	if err := setKeyValue(elem, itemStr); err != nil {
		return i, err
	}

	// Parse subsequent indented lines as part of this struct.
	i++
	itemIndent := seqIndent + 2
	for i < len(lines) {
		lt := strings.TrimSpace(lines[i])
		if lt == "" {
			i++
			continue
		}
		if lineIndent(lines[i]) < itemIndent {
			break
		}
		if err := setKeyValue(elem, lt); err != nil {
			return i, err
		}
		i++
	}

	field.Set(reflect.Append(*field, elem))
	return i, nil
}

// unmarshalNested handles nested structs, map[string]string, and slices of structs.
func unmarshalNested(lines []string, i int, field reflect.Value, ft reflect.Type, depth int) (int, error) {
	if depth > maxNestingDepth {
		return i, errors.New("yaml: maximum nesting depth exceeded")
	}
	switch ft.Kind() { //nolint:exhaustive // only struct, map, slice are relevant
	case reflect.Map:
		if ft.Key().Kind() == reflect.String && ft.Elem().Kind() == reflect.String {
			if field.IsNil() {
				field.Set(reflect.MakeMap(ft))
			}
			return unmarshalStringMap(lines, i, field)
		}
		return i, fmt.Errorf("yaml: unsupported map type %s", ft)
	case reflect.Slice:
		return parseBlockSequence(lines, i, field, ft)
	case reflect.Struct:
		return unmarshalMapping(lines, i, field, depth)
	default:
		return i, fmt.Errorf("yaml: unsupported nested type %s", ft.Kind())
	}
}

// unmarshalStringMap parses a mapping into a map[string]string.
func unmarshalStringMap(lines []string, i int, dst reflect.Value) (int, error) {
	if i >= len(lines) {
		return i, nil
	}
	baseIndent := lineIndent(lines[i])

	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}
		if lineIndent(lines[i]) < baseIndent {
			break
		}
		ci := findKeySeparator(trimmed)
		if ci < 0 {
			break
		}
		k := trimmed[:ci]
		v := ""
		if ci+1 < len(trimmed) {
			v = strings.TrimSpace(trimmed[ci+1:])
		}
		v = stripInlineComment(v)
		dst.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(unquote(v)))
		i++
	}
	return i, nil
}

// --- Field lookup ---

// findKeySeparator finds the colon that separates a YAML key from its value.
// It looks for ": " (colon-space) first, then a trailing colon at end-of-line.
// This avoids splitting on colons inside values (e.g., "url: http://example.com").
func findKeySeparator(s string) int {
	idx := strings.Index(s, ": ")
	if idx >= 0 {
		return idx
	}
	if len(s) > 0 && s[len(s)-1] == ':' {
		return len(s) - 1
	}
	return -1
}

// findField locates a struct field by YAML tag name.
func findField(v reflect.Value, yamlName string) (reflect.Value, reflect.Type, bool) {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, nil, false
	}
	t := v.Type()
	for j := range t.NumField() {
		sf := t.Field(j)
		tag := sf.Tag.Get("yaml")
		name := tagName(tag)
		if name == yamlName {
			return v.Field(j), sf.Type, true
		}
		// Fallback: match field name case-insensitively.
		if name == "" && strings.EqualFold(sf.Name, yamlName) {
			return v.Field(j), sf.Type, true
		}
	}
	return reflect.Value{}, nil, false
}

// tagName extracts the field name from a struct tag like "name,omitempty".
func tagName(tag string) string {
	if tag == "" || tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// --- Scalar helpers ---

// parseBool parses YAML boolean values including yes/no/on/off variants.
func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "false", "no", "off":
		return false, nil
	case "on", "true", "yes":
		return true, nil
	}
	return false, fmt.Errorf("invalid bool %q", s)
}

// setScalar sets a reflect.Value from a string, handling type conversion.
func setScalar(field reflect.Value, s string) error { //nolint:cyclop // natural switch on reflect.Kind
	s = unquote(s)
	switch field.Kind() { //nolint:exhaustive // only supported scalar types handled
	case reflect.Bool:
		b, err := parseBool(s)
		if err != nil {
			return fmt.Errorf("parse bool %q: %w", s, err)
		}
		field.SetBool(b)
	case reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("parse float64 %q: %w", s, err)
		}
		field.SetFloat(f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse int %q: %w", s, err)
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", s, err)
		}
		field.SetUint(n)
	case reflect.String:
		field.SetString(s)
	default:
		return fmt.Errorf("unsupported scalar type %s", field.Kind())
	}
	return nil
}

// setSliceFromStrings sets a []string or []T field from a list of string values.
func setSliceFromStrings(field reflect.Value, ft reflect.Type, items []string) error {
	if ft.Elem().Kind() == reflect.String {
		for _, item := range items {
			field.Set(reflect.Append(field, reflect.ValueOf(unquote(strings.TrimSpace(item)))))
		}
		return nil
	}
	return fmt.Errorf("unsupported flow sequence element type %s", ft.Elem().Kind())
}

// unquote removes surrounding quotes from a string value.
// Double-quoted strings have escape sequences processed; single-quoted strings do not.
func unquote(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return unescapeDoubleQuoted(s[1 : len(s)-1])
		}
		if s[0] == '\'' && s[len(s)-1] == '\'' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// unescapeDoubleQuoted processes escape sequences in a double-quoted YAML string.
func unescapeDoubleQuoted(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 >= len(s) {
			b.WriteByte(s[i])
			continue
		}
		next := s[i+1]
		if replacement, ok := escapeMap[next]; ok {
			b.WriteByte(replacement)
			i++
			continue
		}
		if next == 'u' {
			i += unescapeUnicode(&b, s, i)
			continue
		}
		b.WriteByte('\\')
		b.WriteByte(next)
		i++
	}
	return b.String()
}

// unescapeUnicode writes a \uXXXX escape to b and returns the number of extra bytes consumed.
func unescapeUnicode(b *strings.Builder, s string, i int) int {
	if i+5 < len(s) {
		hex := s[i+2 : i+6]
		code, err := strconv.ParseUint(hex, 16, 32)
		if err == nil && code <= uint64(unicode.MaxRune) {
			b.WriteRune(rune(code))
			return 5
		}
	}
	b.WriteByte('\\')
	b.WriteByte('u')
	return 1
}

// stripInlineComment removes an inline comment (` #`) from a YAML value,
// respecting single and double quoted strings.
func stripInlineComment(s string) string { //nolint:cyclop // quote-aware character parser
	inDouble := false
	inSingle := false
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '\\' && inDouble:
			i++ // skip escaped character
		case s[i] == '"' && !inSingle:
			inDouble = !inDouble
		case s[i] == '\'' && !inDouble:
			inSingle = !inSingle
		case s[i] == '#' && !inDouble && !inSingle && i > 0 && s[i-1] == ' ':
			return strings.TrimRight(s[:i-1], " ")
		}
	}
	return s
}

// splitFlow splits a flow sequence contents "a, b, c" into items.
// It is quote-aware: commas inside quoted strings are not treated as delimiters.
func splitFlow(s string) []string { //nolint:cyclop // character-by-character parser
	if s == "" {
		return nil
	}
	var result []string
	var current strings.Builder
	inDouble := false
	inSingle := false
	for _, c := range s {
		switch {
		case c == '"' && !inSingle:
			inDouble = !inDouble
			current.WriteRune(c)
		case c == '\'' && !inDouble:
			inSingle = !inSingle
			current.WriteRune(c)
		case c == ',' && !inDouble && !inSingle:
			if v := strings.TrimSpace(current.String()); v != "" {
				result = append(result, v)
			}
			current.Reset()
		default:
			current.WriteRune(c)
		}
	}
	if v := strings.TrimSpace(current.String()); v != "" {
		result = append(result, v)
	}
	return result
}

// isNullLiteral returns true if s represents a YAML null value.
func isNullLiteral(s string) bool {
	switch s {
	case "NULL", "Null", "null", "~":
		return true
	}
	return false
}

// lineIndent returns the number of leading spaces on a line.
func lineIndent(line string) int {
	for i := range len(line) {
		if line[i] != ' ' {
			return i
		}
	}
	return len(line)
}

// --- Marshal internals ---

// marshalField writes a single struct field to the builder.
func marshalField(b *strings.Builder, prefix string, name string, fv reflect.Value, omitempty bool, indent int) {
	switch fv.Kind() { //nolint:exhaustive // only supported YAML types handled
	case reflect.Bool, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		marshalScalarField(b, prefix, name, marshalScalarValue(fv))
	case reflect.Map:
		marshalMapField(b, prefix, name, fv, omitempty, indent)
	case reflect.Pointer:
		if fv.IsNil() {
			return
		}
		marshalField(b, prefix, name, fv.Elem(), omitempty, indent)
		return
	case reflect.Slice:
		marshalSliceField(b, prefix, name, fv, omitempty, indent)
	case reflect.String:
		marshalStringField(b, prefix, name, fv.String(), indent)
	case reflect.Struct:
		b.WriteString(prefix)
		b.WriteString(name)
		b.WriteString(":\n")
		marshalStruct(b, fv, indent+1)
	}
}

// marshalMapField writes a map field to the builder.
func marshalMapField(b *strings.Builder, prefix string, name string, fv reflect.Value, omitempty bool, indent int) {
	if fv.Len() == 0 {
		if !omitempty {
			b.WriteString(prefix)
			b.WriteString(name)
			b.WriteString(": {}\n")
		}
		return
	}
	b.WriteString(prefix)
	b.WriteString(name)
	b.WriteString(":\n")
	mapPrefix := strings.Repeat("  ", indent+1)
	keys := make([]string, 0, fv.Len())
	iter := fv.MapRange()
	for iter.Next() {
		keys = append(keys, iter.Key().String())
	}
	slices.Sort(keys)
	for _, k := range keys {
		b.WriteString(mapPrefix)
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(MarshalScalarString(fv.MapIndex(reflect.ValueOf(k)).String()))
		b.WriteByte('\n')
	}
}

// marshalScalarField writes a simple "key: value\n" line.
func marshalScalarField(b *strings.Builder, prefix, name, value string) {
	b.WriteString(prefix)
	b.WriteString(name)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteByte('\n')
}

func marshalScalarValue(v reflect.Value) string {
	switch v.Kind() { //nolint:exhaustive // only supported scalar types handled
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.String:
		return MarshalScalarString(v.String())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// marshalSliceField writes a slice field to the builder.
func marshalSliceField(b *strings.Builder, prefix string, name string, fv reflect.Value, omitempty bool, indent int) {
	if fv.Len() == 0 {
		if !omitempty {
			b.WriteString(prefix)
			b.WriteString(name)
			b.WriteString(": []\n")
		}
		return
	}
	b.WriteString(prefix)
	b.WriteString(name)
	b.WriteString(":\n")
	elemPrefix := strings.Repeat("  ", indent+1)
	for j := range fv.Len() {
		elem := fv.Index(j)
		if elem.Kind() == reflect.String {
			b.WriteString(elemPrefix)
			b.WriteString("- ")
			b.WriteString(MarshalScalarString(elem.String()))
			b.WriteByte('\n')
		} else if elem.Kind() == reflect.Struct {
			b.WriteString(elemPrefix)
			b.WriteString("- ")
			marshalStructInline(b, elem, indent+2)
		}
	}
}

// marshalStringField writes a string field, using block scalar for multiline values.
func marshalStringField(b *strings.Builder, prefix string, name string, s string, indent int) {
	if strings.Contains(s, "\n") {
		b.WriteString(prefix)
		b.WriteString(name)
		b.WriteString(": |\n")
		blockPrefix := strings.Repeat("  ", indent+1)
		trimmed := strings.TrimRight(s, "\n")
		for line := range strings.SplitSeq(trimmed, "\n") {
			b.WriteString(blockPrefix)
			b.WriteString(line)
			b.WriteByte('\n')
		}
	} else {
		b.WriteString(prefix)
		b.WriteString(name)
		b.WriteString(": ")
		b.WriteString(MarshalScalarString(s))
		b.WriteByte('\n')
	}
}

func marshalStruct(b *strings.Builder, v reflect.Value, indent int) {
	t := v.Type()
	prefix := strings.Repeat("  ", indent)

	for i := range t.NumField() {
		sf := t.Field(i)
		fv := v.Field(i)
		tag := sf.Tag.Get("yaml")
		if tag == "-" {
			continue
		}

		name := tagName(tag)
		if name == "" {
			name = strings.ToLower(sf.Name)
		}

		omitempty := strings.Contains(tag, "omitempty")
		if omitempty && fv.IsZero() {
			continue
		}

		marshalField(b, prefix, name, fv, omitempty, indent)
	}
}

// marshalStructInline writes a struct as the first item in a "- key: value" sequence.
func marshalStructInline(b *strings.Builder, v reflect.Value, indent int) {
	t := v.Type()
	prefix := strings.Repeat("  ", indent)
	first := true

	for i := range t.NumField() {
		sf := t.Field(i)
		fv := v.Field(i)
		tag := sf.Tag.Get("yaml")
		if tag == "-" {
			continue
		}

		name := tagName(tag)
		if name == "" {
			name = strings.ToLower(sf.Name)
		}

		omitempty := strings.Contains(tag, "omitempty")
		if omitempty && fv.IsZero() {
			continue
		}

		if first {
			// First field is on the "- " line.
			b.WriteString(name)
			b.WriteString(": ")
			b.WriteString(marshalScalarValue(fv))
			b.WriteByte('\n')
			first = false
		} else {
			b.WriteString(prefix)
			b.WriteString(name)
			b.WriteString(": ")
			b.WriteString(marshalScalarValue(fv))
			b.WriteByte('\n')
		}
	}
}

// MarshalScalarString quotes a string if it contains special YAML characters.
func MarshalScalarString(s string) string {
	if s == "" {
		return `""`
	}
	// Quote if it looks like a bool, number, or contains special chars.
	if needsQuoting(s) {
		return `"` + escapeScalar(s) + `"`
	}
	return s
}

// escapeScalar escapes special characters for a double-quoted YAML scalar.
func escapeScalar(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := range len(s) {
		if esc, ok := reverseEscapeMap[s[i]]; ok {
			b.WriteString(esc)
		} else if s[i] < 0x20 || s[i] == 0x7F {
			fmt.Fprintf(&b, `\u%04X`, s[i])
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	if s[0] == ' ' || s[len(s)-1] == ' ' {
		return true
	}
	if isYAMLKeyword(s) || isNumericLiteral(s) {
		return true
	}
	return containsSpecialChar(s)
}

// isYAMLKeyword returns true if s matches a YAML reserved keyword.
func isYAMLKeyword(s string) bool {
	switch strings.ToLower(s) {
	case "-.inf", ".inf", ".nan", "false", "no", "null", "off", "on", "true", "yes", "~":
		return true
	}
	return false
}

// isNumericLiteral returns true if s could be parsed as a number or alternate numeric literal.
func isNumericLiteral(s string) bool {
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return isAlternateNumericLiteral(s)
}

// isAlternateNumericLiteral checks for YAML 1.1/1.2 hex, octal, and binary prefixes.
func isAlternateNumericLiteral(s string) bool {
	if len(s) <= 2 || s[0] != '0' {
		return false
	}
	switch s[1] {
	case 'B', 'O', 'X', 'b', 'o', 'x':
		return true
	}
	return false
}

// containsSpecialChar returns true if s contains control characters or YAML special characters.
func containsSpecialChar(s string) bool {
	// Leading dash or question mark can be ambiguous.
	if s[0] == '-' || s[0] == '?' {
		return true
	}
	for _, c := range s {
		if c < ' ' || c == 0x7F {
			return true
		}
		if c < 256 && yamlSpecialChars[c] {
			return true
		}
	}
	return false
}
