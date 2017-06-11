package mdtool

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	//        bf "gopkg.in/russross/blackfriday.v2"
)

// from
//
func diff(orig string, b2 []byte) (data []byte, err error) {
	f2, err := ioutil.TempFile("", "markdownfmt")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	f2.Write(b2)
	f2.Close()
	data, err = exec.Command("diff", "-u", orig, f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
}

var fmtcases = [][2]string{
	{"# h1", "# h1"},
	{"# *h1*", "# *h1*"},
	{"#  h1", "# h1"},
	{"# h1 *emph*", "# h1 *emph*"},
	{"# h1 *emph* **strong** ***triple***", "# h1 *emph* **strong** ***triple***"},
	{" `foo` ", "`foo`"},
	{"[text](url)", "[text](url)"},
	{"[***triple bold***](url)", "[***triple bold***](url)"},
	{"![alt](image.jpg)", "![alt](image.jpg)"},
	{"[![alt](image.jpg)](url)", "[![alt](image.jpg)](url)"},
	{"<https://golang.org/>", "<https://golang.org/>"},
	{"a https://golang.org/ b", "a https://golang.org/ b"},
	{"``` foo\ncode\n```", "```foo\ncode\n```"},
	{"* 1\n* 2\n* 3\n", "- 1\n- 2\n- 3\n"},
}

func TestFmt2(t *testing.T) {
	for idx, tt := range fmtcases {
		src := []byte(tt[0])
		want := strings.TrimSpace(tt[1])
		got := strings.TrimSpace(string(Fmt2(src)))
		if got != want {
			t.Errorf("Case %d)\n'%s'\n'%s'", idx, want, got)
		}
	}
}

func TestFmt2Fixtures(t *testing.T) {
	files, err := filepath.Glob("fixtures/*.md")
	if err != nil {
		t.Fatalf("Unable to get testcase files: %s", err)
	}
	for _, filename := range files {
		raw, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Errorf("Unable to read %q: %s", filename, err)
		}
		got := Fmt2(raw)
		if !bytes.Equal(raw, got) {
			df, err := diff(filename, got)
			if err != nil {
				t.Errorf("Unable to diff: %s", err)
			}
			t.Errorf("File %q did not match:\n%s", filename, df)
		}
	}
}
