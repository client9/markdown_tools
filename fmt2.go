package mdtool

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"

	bf "gopkg.in/russross/blackfriday.v2"
)

type fmtRenderer struct {
	debug        *log.Logger
	olCount      map[*bf.Node]int
	inlink       bool
	inlinkBuffer *bytes.Buffer
	listDepth    int
}

func newFmtRenderer() *fmtRenderer {
	return &fmtRenderer{
		debug:        log.New(os.Stderr, "debug ", 0),
		olCount:      make(map[*bf.Node]int),
		inlinkBuffer: new(bytes.Buffer),
	}

}

// Render does a generic walk
func (f *fmtRenderer) Render(ast *bf.Node) []byte {
	var buf bytes.Buffer
	ast.Walk(func(node *bf.Node, entering bool) bf.WalkStatus {
		return f.RenderNode(&buf, node, entering)
	})
	return buf.Bytes()
}

// RenderNode renders a node to a write
func (f *fmtRenderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	// case bf.BlockQuote
	case bf.Paragraph:
		break
	case bf.Document:
		break
	case bf.Text:
		out := w
		if f.inlink {
			out = f.inlinkBuffer
		}
		out.Write(node.Literal)
	case bf.Code:
		out := w
		if f.inlink {
			out = f.inlinkBuffer
		}
		out.Write([]byte{'`'})
		out.Write(node.Literal)
		out.Write([]byte{'`'})
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
		out := w
		if f.inlink {
			out = f.inlinkBuffer
		}
		out.Write([]byte{'*'})
	case bf.Strong:
		out := w
		if f.inlink {
			out = f.inlinkBuffer
		}
		out.Write([]byte{'*', '*'})
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
	case bf.Link:
		if entering {
			f.inlinkBuffer.Reset()
			f.inlink = true
		} else {
			f.inlink = false
			// TODO add link title info

			w.Write([]byte{'['})
			w.Write(f.inlinkBuffer.Bytes())
			w.Write([]byte{']'})
			// TBD add space
			w.Write([]byte{'('})
			w.Write(node.LinkData.Destination)
			w.Write([]byte{')'})
		}
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
		}

	default:
		f.debug.Printf("RENDER NODE: [%v] %+v", entering, *node)
	}
	return bf.GoToNext
}

// Fmt2 reformats Markdown using BlackFriday v2
func Fmt2(input []byte) []byte {
	r := newFmtRenderer()
	return bf.Markdown(input, bf.WithRenderer(r))
}
