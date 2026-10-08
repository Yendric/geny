package islands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var errNotIsland = errors.New("not an island")

type tag struct {
	island      Island
	end         int
	selfClosing bool
}

func parseTag(src []byte, start int, reg *Registry) (tag, error) {
	s := scanner{src: src, pos: start}
	if !s.consume("<") {
		return tag{}, errNotIsland
	}
	name := s.name()
	if !reg.Has(name) {
		return tag{}, errNotIsland
	}
	t, err := s.tagBody(New(name))
	if err != nil {
		return tag{}, fmt.Errorf("island %s: %w", name, err)
	}
	return t, nil
}

type scanner struct {
	src []byte
	pos int
}

func isIdentByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
		c == '_' || c == '-' || c == '.'
}

func (s *scanner) peek() byte {
	if s.pos >= len(s.src) {
		return 0
	}
	return s.src[s.pos]
}

func (s *scanner) consume(prefix string) bool {
	if !bytes.HasPrefix(s.src[s.pos:], []byte(prefix)) {
		return false
	}
	s.pos += len(prefix)
	return true
}

func (s *scanner) skipSpace() {
	for s.pos < len(s.src) && bytes.IndexByte([]byte(" \t\r\n"), s.src[s.pos]) >= 0 {
		s.pos++
	}
}

func (s *scanner) ident() string {
	start := s.pos
	for s.pos < len(s.src) && isIdentByte(s.src[s.pos]) {
		s.pos++
	}
	return string(s.src[start:s.pos])
}

// name reads an island name such as ui/Button, leaving the slash of a trailing /> unread.
func (s *scanner) name() string {
	start := s.pos
	for s.ident() != "" && s.peek() == '/' && s.pos+1 < len(s.src) && isIdentByte(s.src[s.pos+1]) {
		s.pos++
	}
	return string(s.src[start:s.pos])
}

func (s *scanner) tagBody(island Island) (tag, error) {
	directiveSet := false
	for {
		s.skipSpace()
		switch {
		case s.pos >= len(s.src):
			return tag{}, errors.New("unterminated tag")
		case s.consume(">"):
			return tag{island: island, end: s.pos}, nil
		case s.consume("/>"):
			return tag{island: island, end: s.pos, selfClosing: true}, nil
		}

		key := s.ident()
		if key == "" {
			return tag{}, fmt.Errorf("unexpected %q in tag", s.peek())
		}

		if key == "client" && s.consume(":") {
			if directiveSet {
				return tag{}, errors.New("more than one client directive")
			}
			d, err := ParseDirective(s.ident())
			if err != nil {
				return tag{}, err
			}
			if s.peek() == '=' {
				return tag{}, errors.New("client directives take no value")
			}
			directiveSet = true
			island.Directive = d
			continue
		}

		value := json.RawMessage("true")
		if s.consume("=") {
			v, err := s.value()
			if err != nil {
				return tag{}, fmt.Errorf("prop %s: %w", key, err)
			}
			value = v
		}
		var err error
		if island, err = island.WithProp(key, value); err != nil {
			return tag{}, err
		}
	}
}

func (s *scanner) value() (json.RawMessage, error) {
	switch quote := s.peek(); quote {
	case '"', '\'':
		s.pos++
		end := bytes.IndexByte(s.src[s.pos:], quote)
		if end < 0 {
			return nil, errors.New("unterminated string")
		}
		raw, err := json.Marshal(string(s.src[s.pos : s.pos+end]))
		s.pos += end + 1
		return raw, err
	case '{':
		s.pos++
		dec := json.NewDecoder(bytes.NewReader(s.src[s.pos:]))
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		s.pos += int(dec.InputOffset())
		s.skipSpace()
		if !s.consume("}") {
			return nil, errors.New("expected }")
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, raw); err != nil {
			return nil, err
		}
		return compact.Bytes(), nil
	}
	return nil, errors.New(`value must be "string" or {json}`)
}
