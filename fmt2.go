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
	inimg        bool
	inpara       bool
	inlinkBuffer *bytes.Buffer
	inimgBuffer  *bytes.Buffer
	paraBuffer   *bytes.Buffer
	listDepth    int
}

func newFmtRenderer() *fmtRenderer {
	return &fmtRenderer{
		debug:        log.New(os.Stderr, "debug ", 0),
		olCount:      make(map[*bf.Node]int),
		inlinkBuffer: new(bytes.Buffer),
		inimgBuffer: new(bytes.Buffer),
		paraBuffer: new(bytes.Buffer),
	}

}

func (f *fmtRenderer) Writer(w io.Writer) io.Writer {
	// might need to be a stack
	// but for now, and img can be inside a link
	// so img comes first.
	if f.inimg {
		return f.inimgBuffer
	}
	if f.inlink {
		return f.inlinkBuffer
	}
	if f.inpara {
		return f.paraBuffer
	}
	return w
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
		if entering {
			f.inparaBuffer.Reset()
			f.inpara = true
		} else {
			f.inpara = false
			out := f.Writer(w)
			out.Write(f.inparaBuffer.Bytes())
		}
	case bf.Document:
		break
	case bf.Text:
		out := f.Writer(w)
		out.Write(node.Literal)
	case bf.Code:
		out := f.Writer(w)
		out.Write([]byte{'`'})
		out.Write(node.Literal)
		out.Write([]byte{'`'})
	case bf.CodeBlock:
		// codeblocks can be inside a list or blockquote
		// so need to get writer
		out := f.Writer(w)
		// TBD node.CodeBlockData.IsFenced
		// TBD parent is list item or not?
		// TBD parent is blockquote or not?
		out.Write([]byte{'`', '`', '`'})
		if len(node.CodeBlockData.Info) != 0 {
			out.Write(node.CodeBlockData.Info)
		}
		out.Write([]byte{'\n'})
		out.Write(node.Literal)
		out.Write([]byte{'`', '`', '`', '\n'})
	case bf.Del:
		out := f.Writer(w)
		out.Write([]byte{'~', '~'})
	case bf.Emph:
		out := f.Writer(w)
		out.Write([]byte{'*'})
	case bf.Strong:
		out := f.Writer(w)
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
	case bf.Image:
		if entering {
			f.inimgBuffer.Reset()
			f.inimg = true
		} else {
			imgalt := f.inimgBuffer.Bytes()
			f.inimg = false

			// image can be in a link!
			// [![alt](url)](text)
			out := f.Writer(w)
			out.Write([]byte{'!','['})
			out.Write(imgalt)
			out.Write([]byte{']'})
			// todo
			out.Write([]byte{'('})
			out.Write(node.LinkData.Destination)
			out.Write([]byte{')'})
		}
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
