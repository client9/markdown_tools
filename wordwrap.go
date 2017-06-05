package mdtool

import (
	"bytes"
)

// WordWrap formats text word wrap, including initial prefix and ongoing
// indentation
func WordWrap(src []byte, length int, prefix string, indent string) []byte {
	out := bytes.Buffer{}
	// TODO: do something about counting tab size
	prefixLen := len(prefix)
	indentLen := len(indent)
	out.WriteString(prefix)
	if length <= 1 {
		length = 1 << 23
	}
	col := prefixLen
	words := bytes.Fields(src)
	for _, word := range words {
		// if word is so big, just put it on it's own line
		if len(word) >= length-indentLen {
			out.WriteByte('\n')
			out.WriteString(indent)
			out.Write(word)
			out.WriteByte('\n')
			out.WriteString(indent)
			col = indentLen
			continue
		}
		if len(word)+col+1 < length {
			if col != indentLen {
				out.WriteByte(' ')
				col++
			}
			out.Write(word)
			col += len(word)
			continue
		}
		// word overflows
		out.WriteByte('\n')
		out.WriteString(indent)
		out.Write(word)
		col = indentLen + len(word)
	}
	return out.Bytes()
}
