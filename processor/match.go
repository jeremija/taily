package processor

import (
	"strings"
	"time"

	"github.com/jeremija/taily/types"
)

type match struct {
	time     time.Time
	messages []types.Message
}

func groupKey(groupBy []string, message types.Message) string {
	var s strings.Builder

	if len(groupBy) == 0 {
		return string(message.ReaderID)
	}

	s.WriteString(string(message.ReaderID))

	for _, fieldName := range groupBy {
		v := message.Fields[fieldName]
		s.WriteString("_")
		s.WriteString(v)
	}

	return s.String()
}
