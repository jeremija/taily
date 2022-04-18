package formatter_test

import (
	"bytes"
	"testing"

	"github.com/jeremija/taily/formatter"
	"github.com/jeremija/taily/types"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	type testCase struct {
		pattern string
		opts    []formatter.TemplateOpt
		fields  types.Fields
		want    string
		wantErr string
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
		{
			pattern: "{",
			wantErr: "unclosed template at position: 0",
		},
		{
			pattern: "{{",
			wantErr: "multiple open tags at position: 1",
		},
		{
			pattern: "}",
			want:    "}",
			wantErr: "close tag without open tag at position: 0",
		},
		{
			pattern: "{}",
			want:    "``",
		},
		{
			pattern: "[]",
			opts: []formatter.TemplateOpt{
				formatter.WithTags('[', ']'),
				formatter.WithQuotes('"', '"'),
			},
			want: `""`,
		},
		{
			pattern: "[]",
			opts: []formatter.TemplateOpt{
				formatter.WithTags('[', ']'),
				formatter.WithQuotes(0, 0),
			},
			want: ``,
		},
		{
			pattern: "",
			want:    "",
		},
	}

	var b bytes.Buffer

	for _, tc := range testCases {
		b.Reset()
		cmp, err := formatter.NewTemplate(tc.pattern, tc.opts...)

		if tc.wantErr != "" {
			if assert.Error(t, err) {
				assert.Contains(t, tc.wantErr, err.Error())
			}

			assert.Nil(t, cmp)
			continue
		} else {
			assert.NoError(t, err)
		}

		err = cmp.Format(&b, types.Message{
			Fields: tc.fields,
		})
		assert.NoError(t, err)
		assert.Equal(t, tc.want, b.String())
	}
}
