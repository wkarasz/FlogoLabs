package field

import (
	"fmt"
	"strings"
	"text/scanner"
	"unicode"
)

type MappingField struct {
	fields []string
	ref    string
	s      *scanner.Scanner
}

func NewMappingField(fields []string) *MappingField {
	return &MappingField{fields: fields}
}

func ParseMappingField(mRef string) (*MappingField, error) {
	//Remove any . at beginning
	if strings.HasPrefix(mRef, ".") {
		mRef = mRef[1:]
	}
	g := &MappingField{ref: mRef}

	err := g.Start(mRef)
	if err != nil {
		return nil, fmt.Errorf("parse mapping [%s] failed, due to %s", mRef, err.Error())
	}
	return g, nil
}

func (m *MappingField) GetRef() string {
	return m.ref
}

func (m *MappingField) Getfields() []string {
	return m.fields
}

func (m *MappingField) paserName() error {
	fieldName := ""
	switch ch := m.s.Scan(); ch {
	case '.':
		return m.Parser()
	case '[':
		//Done
		if fieldName != "" {
			m.fields = append(m.fields, fieldName)
		}
		m.s.Mode = scanner.ScanInts
		nextAfterBracket := m.s.Scan()
		if nextAfterBracket == '"' || nextAfterBracket == '\'' {
			//Special charactors
			m.s.Mode = scanner.ScanIdents
			return m.handleSpecialField(nextAfterBracket)
		} else {
			//HandleArray
			if m.fields == nil || len(m.fields) <= 0 {
				m.fields = append(m.fields, "["+m.s.TokenText()+"]")
			} else {
				m.fields[len(m.fields)-1] = m.fields[len(m.fields)-1] + "[" + m.s.TokenText() + "]"
			}
			ch := m.s.Scan()
			if ch != ']' {
				return fmt.Errorf("Inliad array format")
			}
			m.s.Mode = scanner.ScanIdents
			return m.Parser()
		}
	case scanner.EOF:
		if fieldName != "" {
			m.fields = append(m.fields, fieldName)
		}
	default:
		fieldName = fieldName + m.s.TokenText()
		if fieldName != "" {
			m.fields = append(m.fields, fieldName)
		}
		return m.Parser()
	}

	return nil
}

func (m *MappingField) handleSpecialField(startQutoes int32) error {
	fieldName := ""
	run := true

	for run {
		switch ch := m.s.Scan(); ch {
		case startQutoes:
			//Check if end with startQutoes]
			nextAfterQuotes := m.s.Scan()
			if nextAfterQuotes == ']' {
				//end specialfield, startover
				m.fields = append(m.fields, fieldName)
				run = false
				return m.Parser()
			} else {
				fieldName = fieldName + string(startQutoes)
				fieldName = fieldName + m.s.TokenText()
			}
		default:
			fieldName = fieldName + m.s.TokenText()
		}
	}
	return nil
}

func (m *MappingField) Parser() error {
	switch ch := m.s.Scan(); ch {
	case '.':
		return m.paserName()
	case '[':
		m.s.Mode = scanner.ScanInts
		nextAfterBracket := m.s.Scan()
		if nextAfterBracket == '"' || nextAfterBracket == '\'' {
			//Special charactors
			m.s.Mode = scanner.ScanIdents
			return m.handleSpecialField(nextAfterBracket)
		} else {
			//HandleArray
			if m.fields == nil || len(m.fields) <= 0 {
				m.fields = append(m.fields, "["+m.s.TokenText()+"]")
			} else {
				m.fields[len(m.fields)-1] = m.fields[len(m.fields)-1] + "[" + m.s.TokenText() + "]"
			}
			//m.handleArray()
			ch := m.s.Scan()
			if ch != ']' {
				return fmt.Errorf("Inliad array format")
			}
			m.s.Mode = scanner.ScanIdents
			return m.Parser()
		}
	case scanner.EOF:
		//Done
		return nil
	default:
		m.fields = append(m.fields, m.s.TokenText())
		return m.paserName()
	}
	return nil
}

func (m *MappingField) Start(jsonPath string) error {
	m.s = new(scanner.Scanner)
	m.s.IsIdentRune = IsIdentRune
	m.s.Init(strings.NewReader(jsonPath))
	m.s.Mode = scanner.ScanIdents
	//Donot skip space and new line
	m.s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<'\r' | 1<<' '
	return m.Parser()
}

func IsIdentRune(ch rune, i int) bool {
	return ch == '$' || ch == '-' || ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
}
