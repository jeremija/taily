package formatter

import (
	"bytes"

	"github.com/jeremija/taily/types"
)

// Title formats a message as Title.
type Title struct{}

// NewTitle creates a new instance of FormatterTitle.
func NewTitle(format string) Title {
	return Title{}
}

func Compile(format string, quote string) types.FormatterFunc {
	var (
		open     bool
		openPos  int
		closePos int = -1
	)

	var constants []string
	var fieldNames []string

	for i, str := range format {
		if str == '{' {
			constant := format[closePos+1 : i]
			constants = append(constants, constant)

			open = true
			openPos = i + 1
			continue
		}

		if str == '}' && open {
			fieldName := format[openPos:i]

			fieldNames = append(fieldNames, fieldName)

			open = false
			closePos = i
		}
	}

	if constant := format[closePos+1:]; constant != "" {
		constants = append(constants, constant)
	}

	return func(buf *bytes.Buffer, message types.Message) error {
		for i, constant := range constants {
			buf.WriteString(constant)

			if i < len(fieldNames) {
				fieldName := fieldNames[i]
				fieldValue := message.Fields[fieldName]

				buf.WriteString(quote)
				buf.WriteString(fieldValue)
				buf.WriteString(quote)
			}
		}

		return nil
	}
}
