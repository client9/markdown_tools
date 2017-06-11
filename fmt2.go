package mdtool

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	bf "gopkg.in/russross/blackfriday.v2"
)

const indent1 = "    "

type state struct {
	buf    *bytes.Buffer
	indent string
}

type stack []state

// add item
func (s *stack) Push(v *bytes.Buffer, prefix string) {
	indent := ""
	if len(*s) != 0 {
		indent = (*s)[len(*s)-1].indent
	}
	indent += prefix
	*s = append(*s, state{buf: v, indent: indent})
}

// get current item
func (s *stack) Peek() *bytes.Buffer {
	return (*s)[len(*s)-1].buf
}

func (s *stack) CurrentIndent() string {
	return (*s)[len(*s)-1].indent
}

// remove last item
func (s *stack) Pop() (*bytes.Buffer, string) {
	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return res.buf, res.indent
}

func isPrevBlock(n *bf.Node) bool {
	prev := n.Prev
	if prev == nil {
		return false
	}
	switch prev.Type {
	case bf.Paragraph, bf.BlockQuote, bf.CodeBlock, bf.List, bf.Heading:
		return true
	}
	return false
}

type fmtRenderer struct {
	debug     *log.Logger
	olCount   map[*bf.Node]int
	bufs      stack
	listDepth int

	linelen int
}

func newFmtRenderer() *fmtRenderer {
	return &fmtRenderer{
		debug:   log.New(os.Stderr, "debug ", 0),
		olCount: make(map[*bf.Node]int),
		bufs:    make(stack, 0, 16),
		linelen: 70,
	}

}

func (f *fmtRenderer) Writer() *bytes.Buffer {
	return f.bufs.Peek()
}

// Render does a generic walk
func (f *fmtRenderer) Render(ast *bf.Node) []byte {
	buf := new(bytes.Buffer)
	f.bufs.Push(buf, "")
	ast.Walk(func(node *bf.Node, entering bool) bf.WalkStatus {
		return f.RenderNode(buf, node, entering)
	})
	f.bufs.Pop()
	if len(f.bufs) != 0 {
		panic("internal error, buffer stack underflow")
	}
	out := buf.Bytes()
	if len(out) > 0 && out[len(out)-1] != '\n' {
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// RenderNode renders a node to a write
func (f *fmtRenderer) RenderNode(_ io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	case bf.BlockQuote:
		if entering {
			out := f.Writer()
			indent := "> "
			if node.Parent.Type == bf.Item {
				indent = indent1 + indent
			}
			if isPrevBlock(node) {
				out.WriteString("\n\n")
			}
			f.bufs.Push(new(bytes.Buffer), indent)
		} else {
			ptext, _ := f.bufs.Pop()
			out := f.Writer()
			out.Write(ptext.Bytes())
			if node.Parent.Type == bf.Item {
				out.WriteByte('\n')
			}
		}
	case bf.Paragraph:
		if entering {
			indent := ""
			// handle case of
			if node.Parent.Type == bf.Item && node.Prev != nil {
				indent = indent1
			}
			f.bufs.Push(new(bytes.Buffer), indent)
		} else {
			ptext, indent := f.bufs.Pop()
			out := f.Writer()
			if node.Parent.Type == bf.Item && node.Prev != nil {
				// lists are special:  indent and bullet has already been
				// printed.
				out.WriteString("\n\n")
				out.Write(WordWrap(ptext.Bytes(), f.linelen, indent, indent))
				out.WriteByte('\n')
			} else {
				prefix := indent
				if node.Parent.Type == bf.Item {
					prefix = ""
				}

				if isPrevBlock(node) {
					out.WriteString("\n\n")
				}
				out.Write(WordWrap(ptext.Bytes(), f.linelen, prefix, indent))
			}
		}
	case bf.Document:
		break
	case bf.Text:
		out := f.Writer()
		out.Write(node.Literal)
	case bf.Code:
		out := f.Writer()
		out.WriteByte('`')
		out.Write(node.Literal)
		out.WriteByte('`')
	case bf.CodeBlock:
		out := f.Writer()

		indent := f.bufs.CurrentIndent()
		switch node.Parent.Type {
		case bf.Item:
			indent += indent1
			leftover := len(indent) % 4
			// BF Bug: must be multiple of 4
			//
			if len(indent)%4 != 0 {
				indent = indent[:len(indent)-leftover]
			}
			out.Write([]byte{'\n', '\n'})
		case bf.BlockQuote:
			break
		default:
			if isPrevBlock(node) {
				out.WriteString("\n\n")
			}
		}

		buf := bytes.Buffer{}
		buf.WriteString("```")
		if len(node.CodeBlockData.Info) != 0 {
			buf.Write(node.CodeBlockData.Info)
		}
		buf.WriteByte('\n')
		buf.Write(node.Literal)
		buf.WriteString("```")

		out.Write(writeIndent(buf.Bytes(), indent))
		if node.Parent.Type == bf.Item {
			out.WriteByte('\n')
		}
	case bf.Del:
		f.Writer().WriteString("~~")
	case bf.Emph:
		f.Writer().WriteByte('*')
	case bf.Strong:
		f.Writer().WriteString("**")
	case bf.Heading:
		if entering {
			f.bufs.Push(new(bytes.Buffer), "")
		} else {
			inner, _ := f.bufs.Pop()
			out := f.Writer()
			out.Write(bytes.Repeat([]byte{'#'}, node.HeadingData.Level))
			out.WriteByte(' ')
			out.Write(inner.Bytes())
		}
	case bf.HorizontalRule:
		out := f.Writer()
		out.Write([]byte{'\n'})
		out.Write([]byte{'-', '-', '-'})
		out.Write([]byte{'\n'})
	case bf.Image:
		if entering {
			f.bufs.Push(new(bytes.Buffer), "")
		} else {
			buf, _ := f.bufs.Pop()
			imgalt := buf.Bytes()

			// image can be in a link!
			// [![alt](url)](text)
			out := f.Writer()
			out.Write([]byte{'!', '['})
			out.Write(imgalt)
			out.Write([]byte{']'})
			// TODO: add space
			out.Write([]byte{'('})
			out.Write(node.LinkData.Destination)
			out.Write([]byte{')'})
		}
	case bf.Link:
		if entering {
			f.bufs.Push(new(bytes.Buffer), "")
		} else {
			buf, _ := f.bufs.Pop()
			linktext := buf.Bytes()
			// TODO add link title info
			out := f.Writer()
			out.WriteByte('[')
			out.Write(linktext)
			out.WriteByte(']')
			// TBD add space
			out.WriteByte('(')
			out.Write(node.LinkData.Destination)
			out.WriteByte(')')
		}
	case bf.List:
		if entering {
			indent := ""
			if f.listDepth > 0 {
				indent = indent1
			}
			f.bufs.Push(new(bytes.Buffer), indent)
			f.listDepth++
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				f.olCount[node] = 0
			}
		} else {
			buf, _ := f.bufs.Pop()
			listtext := buf.Bytes()
			out := f.Writer()
			if isPrevBlock(node) {
				out.WriteString("\n\n")
			}
			out.Write(listtext)

			f.listDepth--
			if f.listDepth < 0 {
				panic("underflow listdepth")
			}
			delete(f.olCount, node)
		}
	case bf.Item:
		if entering {
			// generate bullet
			buf := bytes.Buffer{}
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				f.olCount[node.Parent]++
				buf.WriteString(strconv.Itoa(f.olCount[node.Parent]))
				buf.WriteByte('.')
			} else {
				buf.WriteByte('-')
			}
			buf.WriteByte(' ')
			bullet := buf.Bytes()

			out := f.Writer()
			out.WriteString(strings.Repeat(indent1, f.listDepth-1))
			out.Write(bullet)
			f.bufs.Push(new(bytes.Buffer), strings.Repeat(" ", len(bullet)))
		} else {
			buf, _ := f.bufs.Pop()
			text := buf.Bytes()
			out := f.Writer()
			out.Write(text)
			out.WriteByte('\n')
		}

	default:
		log.Printf("RENDER NODE: [%v] %+v", entering, *node)
	}
	return bf.GoToNext
}

// Fmt2 reformats Markdown using BlackFriday v2
func Fmt2(input []byte) []byte {
	r := newFmtRenderer()
	return bf.Markdown(input, bf.WithRenderer(r))
}
