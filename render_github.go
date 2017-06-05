package mdtool

import (
	gfm "github.com/shurcooL/github_flavored_markdown"
)

// RenderGitHub renders markdown to a similar style to GitHub
// as determined by https://github.com/shurcooL/github_flavored_markdown
//
func RenderGitHub(src []byte) []byte {
	return gfm.Markdown(src)
}
