package mdtool

import (
	"github.com/russross/blackfriday"
)

// RenderHTML renders markdown to HTML in similar style to Hugo
//
func RenderHTML(src []byte) []byte {

	commonExtensions := 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS

	// Extra Blackfriday extensions that Hugo enables by default
	flags := commonExtensions |
		blackfriday.EXTENSION_AUTO_HEADER_IDS |
		blackfriday.EXTENSION_FOOTNOTES

	// HTML output flags
	hflags := 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_FOOTNOTE_RETURN_LINKS

	// TBD
	/*
			if ctx.Config.Smartypants {
			htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
		}

		if ctx.Config.AngledQuotes {
			htmlFlags |= blackfriday.HTML_SMARTYPANTS_ANGLED_QUOTES
		}

		if ctx.Config.Fractions {
			htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
		}

		if ctx.Config.HrefTargetBlank {
			htmlFlags |= blackfriday.HTML_HREF_TARGET_BLANK
		}

		if ctx.Config.SmartDashes {
			htmlFlags |= blackfriday.HTML_SMARTYPANTS_DASHES
		}

		if ctx.Config.LatexDashes {
			htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
		}
	*/
	renderParameters := blackfriday.HtmlRendererParameters{}
	renderer := blackfriday.HtmlRendererWithParameters(hflags, "", "", renderParameters)
	return blackfriday.Markdown(src, renderer, flags)
}
