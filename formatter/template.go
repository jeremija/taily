package formatter

import (
	"bytes"

	"github.com/jeremija/taily/types"
	"github.com/pkg/errors"
)

// Template formats a message using a predefined format string.
type Template struct {
	openTag  rune
	closeTag rune

	openQuote  rune
	closeQuote rune

	constants  []string
	fieldNames []string
}

// TemplateOpt represents an option for the constructor.
type TemplateOpt func(m *Template)

// WithQuotes sets custom quote chars.
func WithQuotes(openQuote, closeQuote rune) TemplateOpt {
	return func(m *Template) {
		m.openQuote = openQuote
		m.closeQuote = closeQuote
	}
}

// WithTags sets custom open and close tags. To disable set to 0.
func WithTags(openTag, closeTag rune) TemplateOpt {
	return func(m *Template) {
		m.openTag = openTag
		m.closeTag = closeTag
	}
}

// NewMustache  creates a new instance of Template.
func NewTemplate(format string, opts ...TemplateOpt) (*Template, error) {
	t := &Template{
		openTag:  '{',
		closeTag: '}',

		openQuote:  '`',
		closeQuote: '`',
	}

	for _, opt := range opts {
		opt(t)
	}

	var (
		open     bool
		openPos  int
		closePos int = -1
	)

	for i, str := range format {
		if str == t.openTag {
			if open {
				return nil, errors.Errorf("multiple open tags at position: %d", i)
			}

			constant := format[closePos+1 : i]
			t.constants = append(t.constants, constant)

			open = true
			openPos = i + 1
			continue
		}

		if str == t.closeTag {
			if !open {
				return nil, errors.Errorf("close tag without open tag at position: %d", i)
			}

			fieldName := format[openPos:i]

			t.fieldNames = append(t.fieldNames, fieldName)

			open = false
			closePos = i
		}
	}

	if open {
		return nil, errors.Errorf("unclosed template at position: %d", openPos-1)
	}

	if constant := format[closePos+1:]; constant != "" {
		t.constants = append(t.constants, constant)
	}

	return t, nil
}

var customFormats = map[string]func(message types.Message) string{
	"$timestamp": func(message types.Message) string {
		return message.Timestamp.Format("2006-01-02T15:04:05.999Z")
	},
}

func (t *Template) Format(buf *bytes.Buffer, message types.Message) error {
	for i, constant := range t.constants {
		buf.WriteString(constant)

		if i < len(t.fieldNames) {
			fieldName := t.fieldNames[i]

			fieldValue, ok := message.Fields[fieldName]

			if !ok {
				if customFormat, ok := customFormats[fieldName]; ok {
					fieldValue = customFormat(message)
				}
			}

			if t.openQuote > 0 {
				buf.WriteRune(t.openQuote)
			}

			buf.WriteString(fieldValue)

			if t.closeQuote > 0 {
				buf.WriteRune(t.closeQuote)
			}
		}
	}

	return nil
}
