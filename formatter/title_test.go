package formatter

import (
	"bytes"
	"testing"

	"github.com/jeremija/taily/types"
	"github.com/stretchr/testify/assert"
)

func TestCompile(t *testing.T) {
	type testCase struct {
		pattern string
		fields  types.Fields
		want    string
	}

	testCases := []testCase{
		{
			pattern: "this is {a} test {b}",
			fields:  nil,
			want:    "this is `` test ``",
		},
		{
			pattern: "this is {a} test {b}",
			fields: types.Fields{
				"a": "one",
				"b": "two",
			},
			want: "this is `one` test `two`",
		},
		{
			pattern: "no custom fields",
			want:    "no custom fields",
		},
	}

	var b bytes.Buffer

	for _, tc := range testCases {
		b.Reset()
		cmp := Compile(tc.pattern, "`")
		err := cmp.Format(&b, types.Message{
			Fields: tc.fields,
		})
		assert.NoError(t, err)
		assert.Equal(t, tc.want, b.String())
	}
}
