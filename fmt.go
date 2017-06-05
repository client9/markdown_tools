package mdtool

import (
	"bytes"
	"go/format"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/russross/blackfriday"
)

type markdownRenderer struct {
	normalTextMarker   map[*bytes.Buffer]int
	orderedListCounter map[int]int
	paragraph          map[int]bool // Used to keep track of whether a given list item uses a paragraph for large spacing.
	listDepth          int
	lastNormalText     string

	// TODO: Clean these up.
	headers      []string
	columnAligns []int
	columnWidths []int
	cells        []string

	lineLength      int
	listBulletChar  string
	listBulletSpace string
	listIndent      string
	headingStyle    string
	hrText          string
	opt             Options

	// stringWidth is used internally to calculate visual width of a string.
	stringWidth func(s string) (width int)
}

func formatCode(lang string, text []byte) (formattedCode []byte, ok bool) {
	switch lang {
	case "Go", "go":
		gofmt, err := format.Source(text)
		if err != nil {
			return nil, false
		}
		return gofmt, true
	default:
		return nil, false
	}
}

// Block-level callbacks.
func (mr *markdownRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	doubleSpace(out)

	// Parse out the language name.
	count := 0
	for _, elt := range strings.Fields(lang) {
		if elt[0] == '.' {
			elt = elt[1:]
		}
		if len(elt) == 0 {
			continue
		}
		out.WriteString("```")
		out.WriteString(elt)
		count++
		break
	}

	if count == 0 {
		out.WriteString("```")
	}
	out.WriteString("\n")

	if formattedCode, ok := formatCode(lang, text); ok {
		out.Write(formattedCode)
	} else {
		out.Write(text)
	}

	out.WriteString("```\n")
}
func (mr *markdownRenderer) BlockQuote(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	lines := bytes.Split(text, []byte("\n"))
	for i, line := range lines {
		if i == len(lines)-1 {
			continue
		}
		out.WriteString(">")
		if len(line) != 0 {
			out.WriteString(" ")
			out.Write(line)
		}
		out.WriteString("\n")
	}
}

// ??
func (mr *markdownRenderer) BlockHtml(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.Write(text)
	out.WriteByte('\n')
}

// ??
func (mr *markdownRenderer) TitleBlock(out *bytes.Buffer, text []byte) {
}

// Header formats h1,h2,h3, etc block
func (mr *markdownRenderer) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	marker := out.Len()
	doubleSpace(out)

	if mr.headingStyle == "atx" || level >= 3 {
		out.WriteString(strings.Repeat("#", level))
		out.WriteByte(' ')
	}

	textMarker := out.Len()
	if !text() {
		out.Truncate(marker)
		return
	}

	if mr.headingStyle != "atx" {
		switch level {
		case 1:
			len := mr.stringWidth(out.String()[textMarker:])
			out.WriteByte('\n')
			out.WriteString(strings.Repeat("=", len))
		case 2:
			len := mr.stringWidth(out.String()[textMarker:])
			out.WriteByte('\n')
			out.WriteString(strings.Repeat("-", len))
		}
	}
	out.WriteString("\n")
}

func (mr *markdownRenderer) HRule(out *bytes.Buffer) {
	doubleSpace(out)
	out.WriteString(mr.hrText)
	out.WriteByte('\n')
}

func (mr *markdownRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	doubleSpace(out)

	mr.listDepth++
	defer func() { mr.listDepth-- }()
	if flags&blackfriday.LIST_TYPE_ORDERED != 0 {
		mr.orderedListCounter[mr.listDepth] = 1
	}
	// this is the entire list, e.g. <ul> to </ul>
	if !text() {
		out.Truncate(marker)
		return
	}
}

func (mr *markdownRenderer) ListItem(out *bytes.Buffer, text []byte, flags int) {

	if flags&blackfriday.LIST_TYPE_ORDERED != 0 {
		spaces := strings.Repeat(mr.listIndent, mr.listDepth-1)
		digits := strconv.Itoa(mr.orderedListCounter[mr.listDepth])
		prefix := spaces + digits + "." + mr.listBulletSpace
		indent := spaces + strings.Repeat(" ", len(digits)) + " " + mr.listBulletSpace
		out.Write(WordWrap(text, mr.lineLength, prefix, indent))
		mr.orderedListCounter[mr.listDepth]++
	} else {
		spaces := strings.Repeat(mr.listIndent, mr.listDepth-1)
		prefix := spaces + mr.listBulletChar + mr.listBulletSpace
		indent := spaces + " " + mr.listBulletSpace
		out.Write(WordWrap(text, mr.lineLength, prefix, indent))
	}
	out.WriteByte('\n')
	if mr.paragraph[mr.listDepth] {
		if flags&blackfriday.LIST_ITEM_END_OF_LIST == 0 {
			out.WriteByte('\n')
		}
		mr.paragraph[mr.listDepth] = false
	}
}
func (mr *markdownRenderer) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()
	doubleSpace(out)

	mr.paragraph[mr.listDepth] = true

	if !text() {
		out.Truncate(marker)
		return
	}

	// make COPY of the part that was written
	// since we are going to modify the original
	// slice.  There are probably tricks to skip this
	// however this is safe.
	tmp := make([]byte, out.Len()-marker)
	copy(tmp, out.Bytes()[marker:])

	out.Truncate(marker)
	out.WriteByte('\n') // <<- without copy, this overwrites tmp
	out.Write(WordWrap(tmp, mr.lineLength, "", ""))
	out.WriteByte('\n')
}

func (mr *markdownRenderer) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	doubleSpace(out)
	for column, cell := range mr.headers {
		out.WriteByte('|')
		out.WriteByte(' ')
		out.WriteString(cell)
		for i := mr.stringWidth(cell); i < mr.columnWidths[column]; i++ {
			out.WriteByte(' ')
		}
		out.WriteByte(' ')
	}
	out.WriteString("|\n")
	for column, width := range mr.columnWidths {
		out.WriteByte('|')
		if mr.columnAligns[column]&blackfriday.TABLE_ALIGNMENT_LEFT != 0 {
			out.WriteByte(':')
		} else {
			out.WriteByte('-')
		}
		for ; width > 0; width-- {
			out.WriteByte('-')
		}
		if mr.columnAligns[column]&blackfriday.TABLE_ALIGNMENT_RIGHT != 0 {
			out.WriteByte(':')
		} else {
			out.WriteByte('-')
		}
	}
	out.WriteString("|\n")
	for i := 0; i < len(mr.cells); {
		for column := range mr.headers {
			cell := []byte(mr.cells[i])
			i++
			out.WriteByte('|')
			out.WriteByte(' ')
			switch mr.columnAligns[column] {
			default:
				fallthrough
			case blackfriday.TABLE_ALIGNMENT_LEFT:
				out.Write(cell)
				for i := mr.stringWidth(string(cell)); i < mr.columnWidths[column]; i++ {
					out.WriteByte(' ')
				}
			case blackfriday.TABLE_ALIGNMENT_CENTER:
				spaces := mr.columnWidths[column] - mr.stringWidth(string(cell))
				for i := 0; i < spaces/2; i++ {
					out.WriteByte(' ')
				}
				out.Write(cell)
				for i := 0; i < spaces-(spaces/2); i++ {
					out.WriteByte(' ')
				}
			case blackfriday.TABLE_ALIGNMENT_RIGHT:
				for i := mr.stringWidth(string(cell)); i < mr.columnWidths[column]; i++ {
					out.WriteByte(' ')
				}
				out.Write(cell)
			}
			out.WriteByte(' ')
		}
		out.WriteString("|\n")
	}

	mr.headers = nil
	mr.columnAligns = nil
	mr.columnWidths = nil
	mr.cells = nil
}

// TableRow does nothing since the markup is handled by Cell
func (mr *markdownRenderer) TableRow(out *bytes.Buffer, text []byte) {
}

// TableHeaderCell renders a <th> tag in markdown
func (mr *markdownRenderer) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	mr.columnAligns = append(mr.columnAligns, align)
	columnWidth := mr.stringWidth(string(text))
	mr.columnWidths = append(mr.columnWidths, columnWidth)
	mr.headers = append(mr.headers, string(text))
}

// TableCell renders a <td> tag in markdown
func (mr *markdownRenderer) TableCell(out *bytes.Buffer, text []byte, align int) {
	columnWidth := mr.stringWidth(string(text))
	column := len(mr.cells) % len(mr.headers)
	if columnWidth > mr.columnWidths[column] {
		mr.columnWidths[column] = columnWidth
	}
	mr.cells = append(mr.cells, string(text))
}

// ??
func (mr *markdownRenderer) Footnotes(out *bytes.Buffer, text func() bool) {
	out.WriteString("<Footnotes: Not implemented.>") // TODO
}

// ??
func (mr *markdownRenderer) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	out.WriteString("<FootnoteItem: Not implemented.>") // TODO
}

// AutoLink handles raw or naked URLs in text.
func (mr *markdownRenderer) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	text := escape(link)
	out.Write(text)
}

// CodeSpan is an inline code span
func (mr *markdownRenderer) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteByte('`')
	out.Write(text)
	out.WriteByte('`')
}

func (mr *markdownRenderer) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("**")
	out.Write(text)
	out.WriteString("**")
}

func (mr *markdownRenderer) Emphasis(out *bytes.Buffer, text []byte) {
	if len(text) == 0 {
		return
	}
	out.WriteByte('*')
	out.Write(text)
	out.WriteByte('*')
}

// Image renders a span image element
func (mr *markdownRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	out.WriteString("![")
	out.Write(alt)
	out.WriteString("](")
	out.Write(escape(link))
	if len(title) != 0 {
		out.WriteString(` "`)
		out.Write(title)
		out.WriteString(`"`)
	}
	out.WriteString(")")
}

// ??
func (mr *markdownRenderer) LineBreak(out *bytes.Buffer) {
	out.WriteString("  \n")
}

// Link renders markdown representation of a HTML <a>
func (mr *markdownRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteByte('[')
	out.Write(content)
	out.WriteByte(']')
	out.WriteByte('(')
	out.Write(escape(link))
	if len(title) != 0 {
		out.WriteString(` "`)
		out.Write(title)
		out.WriteByte('"')
	}
	out.WriteByte(')')
}

func (mr *markdownRenderer) RawHtmlTag(out *bytes.Buffer, tag []byte) {
	out.Write(tag)
}

func (mr *markdownRenderer) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("***")
	out.Write(text)
	out.WriteString("***")
}

func (mr *markdownRenderer) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.WriteString("~~")
	out.Write(text)
	out.WriteString("~~")
}

func (mr *markdownRenderer) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	out.WriteString("<FootnoteRef: Not implemented.>") // TODO
}

// escape replaces instances of backslash with escaped backslash in text.
func escape(text []byte) []byte {
	return bytes.Replace(text, []byte(`\`), []byte(`\\`), -1)
}

func isNumber(data []byte) bool {
	for _, b := range data {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func needsEscaping(text []byte, lastNormalText string) bool {
	switch string(text) {
	case `\`,
		"`",
		"*",
		"_",
		"{", "}",
		"[", "]",
		"(", ")",
		"#",
		"+",
		"-":
		return true
	case "!":
		return false
	case ".":
		// Return true if number, because a period after a number must be escaped to not get parsed as an ordered list.
		return isNumber([]byte(lastNormalText))
	case "<", ">":
		return true
	default:
		return false
	}
}

// Low-level callbacks.
func (mr *markdownRenderer) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (mr *markdownRenderer) NormalText(out *bytes.Buffer, text []byte) {
	normalText := string(text)
	if needsEscaping(text, mr.lastNormalText) {
		text = append([]byte("\\"), text...)
	}
	mr.lastNormalText = normalText
	if mr.listDepth > 0 && string(text) == "\n" { // TODO: See if this can be cleaned up... It's needed for lists.
		return
	}
	cleanString := cleanWithoutTrim(string(text))
	if cleanString == "" {
		return
	}
	if mr.skipSpaceIfNeededNormalText(out, cleanString) { // Skip first space if last character is already a space (i.e., no need for a 2nd space in a row).
		cleanString = cleanString[1:]
	}
	if true {
		tmp := WordWrap([]byte(cleanString), mr.lineLength, "", "")
		if len(cleanString) > 0 && cleanString[0] == ' ' {
			out.Write([]byte{' '})
		}
		out.Write(tmp)
		if len(cleanString) >= 1 && cleanString[len(cleanString)-1] == ' ' {
			out.Write([]byte{' '})
			mr.normalTextMarker[out] = out.Len()
		}
	} else {
		out.WriteString(cleanString)
		if len(cleanString) >= 1 && cleanString[len(cleanString)-1] == ' ' { // If it ends with a space, make note of that.
			mr.normalTextMarker[out] = out.Len()
		}
	}
}

// Header and footer.
func (mr *markdownRenderer) DocumentHeader(out *bytes.Buffer) {}
func (mr *markdownRenderer) DocumentFooter(out *bytes.Buffer) {}

func (mr *markdownRenderer) GetFlags() int { return 0 }

func (mr *markdownRenderer) skipSpaceIfNeededNormalText(out *bytes.Buffer, cleanString string) bool {
	if cleanString[0] != ' ' {
		return false
	}
	if _, ok := mr.normalTextMarker[out]; !ok {
		mr.normalTextMarker[out] = -1
	}
	return mr.normalTextMarker[out] == out.Len()
}

// cleanWithoutTrim is like clean, but doesn't trim blanks.
func cleanWithoutTrim(s string) string {
	var b []byte
	var p byte
	for i := 0; i < len(s); i++ {
		q := s[i]
		if q == '\n' || q == '\r' || q == '\t' {
			q = ' '
		}
		if q != ' ' || p != ' ' {
			b = append(b, q)
			p = q
		}
	}
	return string(b)
}

func doubleSpace(out *bytes.Buffer) {
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
}

// NewRenderer returns a Markdown renderer.
// If opt is nil the defaults are used.
func NewRenderer(opt *Options) blackfriday.Renderer {
	if opt == nil {
		opt = &Options{
			LineLength:      78,
			HrChar:          "-",
			HrLength:        3,
			ListBulletChar:  '-',
			ListIndent:      "  ",
			ListBulletSpace: " ", // "\t",
			HeadingStyle:    "atx",
		}
	}
	switch opt.ListBulletChar {
	case '-', '+', '*':
		break
	default:
		panic("unknown bullet type")
	}

	switch opt.HrChar {
	case "-":
		break
	default:
		panic("unknown HR char")
	}

	return &markdownRenderer{
		normalTextMarker:   make(map[*bytes.Buffer]int),
		orderedListCounter: make(map[int]int),
		paragraph:          make(map[int]bool),

		stringWidth:     runewidth.StringWidth,
		lineLength:      opt.LineLength,
		hrText:          strings.Repeat(opt.HrChar, opt.HrLength),
		listBulletChar:  string(opt.ListBulletChar),
		listIndent:      opt.ListIndent,
		listBulletSpace: opt.ListBulletSpace,
		headingStyle:    opt.HeadingStyle,
	}
}

// Options specifies options for formatting.
type Options struct {
	// LineLength wraps lines at N characters, or -1 for run-on
	LineLength int

	// ListIndent is the identation string to use
	ListIndent string

	// ListBulletChar is the bullet for unordered lists, either '-' or '*'
	ListBulletChar byte

	// ListBulletSpace is whitespace between bullet and text
	ListBulletSpace string

	// HeadingStyle controls the markdown style for headlines
	//  so far it's only hash, or nothash
	HeadingStyle string

	// HrChar is the character to use for horizontal rules
	HrChar string

	// HrLength is the length of the horizontal rule in chars
	HrLength int
}

// Fmt formats Markdown.
func Fmt(text []byte, opt *Options) []byte {

	// extensions for GitHub Flavored Markdown-like parsing.
	const extensions = blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK |
		blackfriday.EXTENSION_DEFINITION_LISTS

	output := blackfriday.Markdown(text, NewRenderer(opt), extensions)
	return output
}
