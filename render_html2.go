package mdtool

import (
	bf "gopkg.in/russross/blackfriday.v2"
)

// RenderHTML2 renders markdown to HTML using Black Friday v2
//
func RenderHTML2(src []byte) []byte {
	return bf.Markdown(src)
}
