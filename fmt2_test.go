package mdtool


import (
	"strings"
	"testing"


//        bf "gopkg.in/russross/blackfriday.v2"
)

var fmtcases = [][2]string{
	{"# h1", "# h1"},
	{"#  h1", "# h1"},
	{"# h1 *emph*", "# h1 *emph*"},
	{"# h1 *emph* **strong** ***triple***", "# h1 *emph* **strong** ***triple***"},
	{" `foo` ", "`foo`"},
	{"``` foo\ncode\n```", "```foo\ncode\n```"},
	{"* 1\n* 2\n* 3\n", "- 1\n- 2\n- 3\n"},
	{`
- 1.1
- 1.2
  - 2.1
  - 2.2
- 1.3
`,
`
- 1.1
- 1.2
    - 2.1
    - 2.2
- 1.3
`},
{`
1. one
1. two
1. three
`,
`
1. one
2. two
3. three
`},
{`
1. one
1. two
  1. two.one
  1. two.two
1. three
`,
`
1. one
2. two
    1. two.one
    2. two.two
3. three
`},
}

func TestFmt2(t *testing.T) {
	for idx, tt := range fmtcases {
		src := []byte(tt[0])
		want := strings.TrimSpace(tt[1])
		got := strings.TrimSpace(string(Fmt2(src)))
		if got != want  {
			t.Errorf("Case %d)\n'%s'\n'%s'", idx, want, got)
		}
	}
}
