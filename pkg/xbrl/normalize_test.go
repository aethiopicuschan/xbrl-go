package xbrl_test

import (
	"testing"

	"github.com/aethiopicuschan/xbrl-go/pkg/xbrl"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeSpace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty string returns empty",
			in:   "",
			want: "",
		},
		{
			name: "string with only converted spaces returns empty",
			in:   "\u00A0\u3000\t",
			want: "",
		},
		{
			name: "string without extra spaces is unchanged",
			in:   "foo bar",
			want: "foo bar",
		},
		{
			name: "collapse and trim ascii whitespace",
			in:   "  foo   bar\tbaz\n",
			want: "foo bar baz",
		},
		{
			name: "convert NBSP and full-width spaces then collapse",
			in:   "\u00A0foo\u3000bar\u00A0baz",
			want: "foo bar baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := xbrl.NormalizeSpace(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
