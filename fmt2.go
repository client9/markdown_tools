package mdtool

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"

	bf "gopkg.in/russross/blackfriday.v2"
)

type stack []*bytes.Buffer

// add item
func (s *stack) Push(v *bytes.Buffer) {
	*s = append(*s, v)
}

// get current item
func (s *stack) Peek() *bytes.Buffer {
	return (*s)[len(*s)-1]
}

// remove last item
func (s *stack) Pop() *bytes.Buffer {
	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return res
}

type fmtRenderer struct {
	debug     *log.Logger
	olCount   map[*bf.Node]int
	bufs      stack
	listDepth int
}

func newFmtRenderer() *fmtRenderer {
	return &fmtRenderer{
		debug:   log.New(os.Stderr, "debug ", 0),
		olCount: make(map[*bf.Node]int),
		bufs:    make(stack, 0, 16),
	}

}

func (f *fmtRenderer) Writer() io.Writer {
	return f.bufs.Peek()
}

// Render does a generic walk
func (f *fmtRenderer) Render(ast *bf.Node) []byte {
	buf := new(bytes.Buffer)
	f.bufs.Push(buf)
	ast.Walk(func(node *bf.Node, entering bool) bf.WalkStatus {
		return f.RenderNode(buf, node, entering)
	})
	return f.bufs.Pop().Bytes()
}

// RenderNode renders a node to a write
func (f *fmtRenderer) RenderNode(_ io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	// case bf.BlockQuote
	case bf.Paragraph:
		if entering {
			f.bufs.Push(new(bytes.Buffer))
		} else {
			ptext := f.bufs.Pop().Bytes()
			out := f.Writer()
			out.Write(ptext)
		}
	case bf.Document:
		break
	case bf.Text:
		out := f.Writer()
		out.Write(node.Literal)
	case bf.Code:
		out := f.Writer()
		out.Write([]byte{'`'})
		out.Write(node.Literal)
		out.Write([]byte{'`'})
	case bf.CodeBlock:
		// codeblocks can be inside a list or blockquote
		// so need to get writer
		out := f.Writer()
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
		out := f.Writer()
		out.Write([]byte{'~', '~'})
	case bf.Emph:
		out := f.Writer()
		out.Write([]byte{'*'})
	case bf.Strong:
		out := f.Writer()
		out.Write([]byte{'*', '*'})
	case bf.Heading:
		out := f.Writer()
		if !entering {
			out.Write([]byte{'\n', '\n'})
			break
		}
		out.Write(bytes.Repeat([]byte{'#'}, node.HeadingData.Level))
		out.Write([]byte{' '})
	case bf.HorizontalRule:
		out := f.Writer()
		out.Write([]byte{'\n'})
		out.Write([]byte{'-', '-', '-'})
		out.Write([]byte{'\n'})
	case bf.Image:
		if entering {
			f.bufs.Push(new(bytes.Buffer))
		} else {
			imgalt := f.bufs.Pop().Bytes()

			// image can be in a link!
			// [![alt](url)](text)
			out := f.Writer()
			out.Write([]byte{'!', '['})
			out.Write(imgalt)
			out.Write([]byte{']'})
			// todo
			out.Write([]byte{'('})
			out.Write(node.LinkData.Destination)
			out.Write([]byte{')'})
		}
	case bf.Link:
		if entering {
			f.bufs.Push(new(bytes.Buffer))
		} else {
			linktext := f.bufs.Pop().Bytes()
			// TODO add link title info
			out := f.Writer()
			out.Write([]byte{'['})
			out.Write(linktext)
			out.Write([]byte{']'})
			// TBD add space
			out.Write([]byte{'('})
			out.Write(node.LinkData.Destination)
			out.Write([]byte{')'})
		}
	case bf.List:
		if entering {
			f.bufs.Push(new(bytes.Buffer))
			f.listDepth++
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				f.olCount[node] = 0
			}
		} else {
			listtext := f.bufs.Pop().Bytes()
			out := f.Writer()
			out.Write(listtext)

			f.listDepth--
			if f.listDepth < 0 {
				panic("underflow listdepth")
			}
			delete(f.olCount, node)
		}
	case bf.Item:
		if entering {
			out := f.Writer()
			out.Write([]byte{'\n'})
			out.Write(bytes.Repeat([]byte{' ', ' ', ' ', ' '}, f.listDepth-1))
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				f.olCount[node.Parent]++
				out.Write([]byte(strconv.Itoa(f.olCount[node.Parent])))
				out.Write([]byte{'.'})
			} else {
				out.Write([]byte{'-'})
			}
			out.Write([]byte{' '})
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
