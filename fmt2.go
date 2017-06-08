package mdtool

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"

	bf "gopkg.in/russross/blackfriday.v2"
)

type FmtRenderer struct {
	debug     *log.Logger
	olCount   map[*bf.Node]int
	listDepth int
}

func NewFmtRenderer() *FmtRenderer {
	return &FmtRenderer{
		debug:   log.New(os.Stderr, "debug ", 0),
		olCount: make(map[*bf.Node]int),
	}

}

// Render does a generic walk
func (f *FmtRenderer) Render(ast *bf.Node) []byte {
	var buf bytes.Buffer
	ast.Walk(func(node *bf.Node, entering bool) bf.WalkStatus {
		return f.RenderNode(&buf, node, entering)
	})
	return buf.Bytes()
}

func (f *FmtRenderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	// case bf.BlockQuote
	case bf.Paragraph:
		break
	case bf.Document:
		break
	case bf.Text:
		w.Write(node.Literal)
	case bf.Code:
		w.Write([]byte{'`'})
		w.Write(node.Literal)
		w.Write([]byte{'`'})
	case bf.CodeBlock:
		// TBD node.CodeBlockData.IsFenced
		// TBD parent is list item or not?
		// TBD parent is blockquote or not?
		w.Write([]byte{'`', '`', '`'})
		if len(node.CodeBlockData.Info) != 0 {
			w.Write(node.CodeBlockData.Info)
		}
		w.Write([]byte{'\n'})
		w.Write(node.Literal)
		w.Write([]byte{'`', '`', '`', '\n'})
	case bf.Emph:
		w.Write([]byte{'*'})
	case bf.Strong:
		w.Write([]byte{'*', '*'})
	case bf.Heading:
		if !entering {
			w.Write([]byte{'\n', '\n'})
			break
		}
		w.Write(bytes.Repeat([]byte{'#'}, node.HeadingData.Level))
		w.Write([]byte{' '})
	case bf.HorizontalRule:
		w.Write([]byte{'\n'})
		w.Write([]byte{'-', '-', '-'})
		w.Write([]byte{'\n'})
	case bf.List:
		if entering {
			f.listDepth++
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				f.olCount[node] = 0
			}
		} else {
			f.listDepth--
			if f.listDepth < 0 {
				panic("underflow listdepth")
			}
			delete(f.olCount, node)
			//w.Write([]byte{'\n'})
		}
	case bf.Item:
		f.debug.Printf("RENDER NODE: [%v] %+v", entering, *node)
		if entering {
			w.Write([]byte{'\n'})
			w.Write(bytes.Repeat([]byte{' ', ' ', ' ', ' '}, f.listDepth-1))
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				f.olCount[node.Parent]++
				w.Write([]byte(strconv.Itoa(f.olCount[node.Parent])))
				w.Write([]byte{'.'})
			} else {
				w.Write([]byte{'-'})
			}
			w.Write([]byte{' '})
		} else {
		}

	default:
		f.debug.Printf("RENDER NODE: [%v] %+v", entering, *node)
	}
	return bf.GoToNext
}

func Fmt2(input []byte) []byte {
	r := NewFmtRenderer()
	return bf.Markdown(input, bf.WithRenderer(r))
}
