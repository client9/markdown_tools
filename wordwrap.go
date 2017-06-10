package mdtool

import (
	"bytes"
)

func writeIndent(src []byte, indent string) []byte {
	buf := bytes.Buffer{}
	lines := bytes.Split(src, []byte{'\n'})
	for i, line := range lines {
		// we want to leave the last line without
		// a \n since we'll add it later.
		if i != 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(indent)
		buf.Write(line)
	}
	return buf.Bytes()
}

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
	for i, word := range words {
		// if word is so big, just add it an call it a day
		if len(word) >= length-indentLen {
			out.Write(word)
			out.WriteByte('\n')
			out.WriteString(indent)
			col = indentLen
			continue
		}
		if len(word)+col+1 < length {
			if i != 0 {
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
