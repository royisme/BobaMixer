package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type yamlTokenKind int

const (
	tokenMap yamlTokenKind = iota
	tokenList
)

const (
	boolTrue  = "true"
	boolFalse = "false"
)

type yamlToken struct {
	key       string
	value     string
	kind      yamlTokenKind
	indent    int
	hasValue  bool
	inlineMap bool
}

func parseYAML(data []byte) (map[string]interface{}, error) {
	lines := tokenizeYAML(string(data))
	tokens, err := buildTokens(lines)
	if err != nil {
		return nil, err
	}
	result, idx, err := parseMap(tokens, 0, 0)
	if err != nil {
		return nil, err
	}
	if idx != len(tokens) {
		return nil, fmt.Errorf("unparsed tokens remaining: %d", len(tokens)-idx)
	}
	return result, nil
}

type yamlLine struct {
	indent int
	text   string
}

func tokenizeYAML(raw string) []yamlLine {
	var lines []yamlLine
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		lines = append(lines, yamlLine{indent: indent, text: strings.TrimRight(line, "\r")})
	}
	return lines
}

func buildTokens(lines []yamlLine) ([]yamlToken, error) {
	tokens := make([]yamlToken, 0, len(lines))
	for _, ln := range lines {
		text := strings.TrimSpace(ln.text)
		if strings.HasPrefix(text, "-") {
			entry := strings.TrimSpace(strings.TrimPrefix(text, "-"))
			tok := yamlToken{kind: tokenList, indent: ln.indent, value: entry}
			if entry == "" {
				tok.hasValue = false
			} else if strings.Contains(entry, ":") {
				parts := strings.SplitN(entry, ":", 2)
				tok.inlineMap = true
				tok.key = strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if val != "" {
					tok.value = val
					tok.hasValue = true
				}
			} else {
				tok.hasValue = true
				tok.value = entry
			}
			tokens = append(tokens, tok)
			continue
		}
		parts := strings.SplitN(text, ":", 2)
		if len(parts) == 1 {
			return nil, fmt.Errorf("invalid line: %s", text)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		tok := yamlToken{kind: tokenMap, indent: ln.indent, key: key}
		if val != "" {
			tok.value = val
			tok.hasValue = true
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func parseMap(tokens []yamlToken, idx int, indent int) (map[string]interface{}, int, error) {
	result := make(map[string]interface{})
	for idx < len(tokens) {
		tok := tokens[idx]
		if tok.indent < indent {
			break
		}
		if tok.indent > indent {
			return nil, idx, fmt.Errorf("unexpected indent for map entry: %d>%d", tok.indent, indent)
		}
		if tok.kind != tokenMap {
			break
		}
		idx++
		var value interface{}
		var err error
		if tok.hasValue {
			value = parseScalar(tok.value)
		} else {
			if idx >= len(tokens) {
				value = map[string]interface{}{}
			} else {
				next := tokens[idx]
				if next.indent < indent+2 {
					value = map[string]interface{}{}
				} else if next.kind == tokenList && next.indent == indent+2 {
					value, idx, err = parseList(tokens, idx, indent+2)
				} else if next.kind == tokenMap && next.indent == indent+2 {
					value, idx, err = parseMap(tokens, idx, indent+2)
				} else {
					err = fmt.Errorf("unsupported nested structure for key %s", tok.key)
				}
			}
		}
		if err != nil {
			return nil, idx, err
		}
		result[tok.key] = value
	}
	return result, idx, nil
}

func parseList(tokens []yamlToken, idx int, indent int) ([]interface{}, int, error) {
	var items []interface{}
	for idx < len(tokens) {
		tok := tokens[idx]
		if tok.indent < indent {
			break
		}
		if tok.indent > indent {
			return nil, idx, fmt.Errorf("unexpected indent for list: %d>%d", tok.indent, indent)
		}
		if tok.kind != tokenList {
			break
		}
		idx++
		if tok.inlineMap {
			m := make(map[string]interface{})
			if tok.hasValue {
				m[tok.key] = parseScalar(tok.value)
			} else {
				var err error
				if idx < len(tokens) {
					next := tokens[idx]
					if next.indent == indent+2 && next.kind == tokenMap {
						var nested map[string]interface{}
						nested, idx, err = parseMap(tokens, idx, indent+2)
						if err != nil {
							return nil, idx, err
						}
						m[tok.key] = nested
						items = append(items, m)
						continue
					} else if next.indent == indent+2 && next.kind == tokenList {
						var nested []interface{}
						nested, idx, err = parseList(tokens, idx, indent+2)
						if err != nil {
							return nil, idx, err
						}
						m[tok.key] = nested
						items = append(items, m)
						continue
					}
				}
				m[tok.key] = map[string]interface{}{}
			}
			// absorb additional map entries belonging to this list item
			if idx < len(tokens) && tokens[idx].indent == indent+2 && tokens[idx].kind == tokenMap {
				more, nextIdx, err := parseMap(tokens, idx, indent+2)
				if err != nil {
					return nil, idx, err
				}
				for k, v := range more {
					m[k] = v
				}
				idx = nextIdx
			}
			items = append(items, m)
			continue
		}
		if tok.hasValue {
			items = append(items, parseScalar(tok.value))
			continue
		}
		if idx >= len(tokens) {
			items = append(items, nil)
			continue
		}
		next := tokens[idx]
		var (
			val interface{}
			err error
		)
		if next.indent == indent+2 && next.kind == tokenList {
			val, idx, err = parseList(tokens, idx, indent+2)
		} else if next.indent == indent+2 && next.kind == tokenMap {
			val, idx, err = parseMap(tokens, idx, indent+2)
		} else {
			err = errors.New("invalid nested structure in list")
		}
		if err != nil {
			return nil, idx, err
		}
		items = append(items, val)
	}
	return items, idx, nil
}

func parseScalar(raw string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "\"") && strings.HasSuffix(raw, "\"") && len(raw) >= 2 {
		return strings.Trim(raw, "\"")
	}
	if raw == boolTrue || raw == boolFalse {
		return raw == boolTrue
	}
	if strings.Contains(raw, ".") {
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return f
		}
	} else if n, err := strconv.Atoi(raw); err == nil {
		return n
	}
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		trimmed := strings.Trim(raw[1:len(raw)-1], " ")
		if trimmed == "" {
			return []interface{}{}
		}
		parts := splitCSV(trimmed)
		arr := make([]interface{}, 0, len(parts))
		for _, part := range parts {
			arr = append(arr, parseScalar(strings.TrimSpace(part)))
		}
		return arr
	}
	return strings.Trim(raw, "\"")
}

func splitCSV(s string) []string {
	var parts []string
	current := strings.Builder{}
	inQuotes := false
	for _, r := range s {
		switch r {
		case ',':
			if inQuotes {
				current.WriteRune(r)
				continue
			}
			parts = append(parts, current.String())
			current.Reset()
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())
	return parts
}
