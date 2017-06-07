package mdtool

import (
	"bytes"
	"sort"
)

var (
	newlines = []byte{'\n', '\n'}
)

// FaultType defines various markdown structural problems
type FaultType int

const (
	// FaultZero is an unknown type
	FaultZero = FaultType(0)
	// FaultRunawayCodeFence is an un-ended code block
	FaultRunawayCodeFence = FaultType(1)
	// FaultRunawayLinkText is a broken URL
	FaultRunawayLinkText = FaultType(2)
	// FaultRunawayLinkURL is a broken URL with un-ended ']'
	FaultRunawayLinkURL = FaultType(3)
	// FaultLinkTextWhitespace is a broken URL with text have newlines
	FaultLinkTextWhitespace = FaultType(4)
	// FaultLinkURLWhitespace is a broken URL with URL having newlines
	FaultLinkURLWhitespace = FaultType(5)
	// FaultLinkSpaceBetweenTextAndLink is whitespace between ']' and '('
	FaultLinkSpaceBetweenTextAndLink = FaultType(6)
)

func (s FaultType) String() string {
	switch s {
	case FaultRunawayCodeFence:
		return "Runaway Code Fence"
	case FaultRunawayLinkText:
		return "Runaway Lint Text"
	case FaultRunawayLinkURL:
		return "Runaway Link URL"
	case FaultLinkTextWhitespace:
		return "Link Text with Whitespace"
	case FaultLinkURLWhitespace:
		return "Link URL with Whitespace"
	case FaultLinkSpaceBetweenTextAndLink:
		return "Whitespace between Link Text and Link URL"
	}
	return "FAIL"
}

// Fault defined the type and location of markdown problem
type Fault struct {
	Offset int
	Reason FaultType
	Row    int
	Column int
	Line   string
}

// GetLine converts an offset into line with row, col info
func GetLine(raw []byte, offset int) (row, col int, line string) {
	for {
		row++
		idx := bytes.IndexByte(raw, '\n')
		if idx == -1 {
			break
		}
		if idx >= offset {
			return row, offset, string(raw[:idx])
		}
		idx++
		offset -= idx
		raw = raw[idx:]
	}
	return row, offset, string(raw)
}

func verifyURL(raw []byte, faults []Fault) []Fault {
	for idx := 0; idx < len(raw); idx++ {
		i := bytes.IndexByte(raw[idx:], '[')
		if i == -1 {
			break
		}
		start := idx + i
		i += idx + 1
		j := bytes.IndexByte(raw[i:], ']')
		if j == -1 {
			// runaway!
			faults = append(faults, Fault{
				Offset: start,
				Reason: FaultRunawayLinkText,
			})
			break
		}
		desc := raw[i : i+j]
		i += j + 1

		// skip space and tabs
		spaceCount := 0
		for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t') {
			spaceCount++
			i++
		}
		// if ran out, then assume [whatever] is just ok
		if i == len(raw) {
			break
		}
		// if next is a '(', then assume [whatever] is ok
		if raw[i] != '(' {
			continue
		}
		// if we had spaces between ']' and '(' then
		// something is wrong.
		if spaceCount > 0 {
			faults = append(faults, Fault{
				Offset: start,
				Reason: FaultLinkSpaceBetweenTextAndLink,
			})
		}

		i++
		j = bytes.IndexByte(raw[i:], ')')
		if j == -1 {
			faults = append(faults, Fault{
				Offset: start,
				Reason: FaultRunawayLinkURL,
			})
			break
		}

		aurl := raw[i : i+j]
		// we have description and url
		// verify they don't have '\n\n' in them
		if bytes.Contains(desc, newlines) {
			faults = append(faults, Fault{
				Offset: start,
				Reason: FaultLinkTextWhitespace,
			})
		}
		if bytes.IndexByte(aurl, '\n') != -1 {
			faults = append(faults, Fault{
				Offset: start,
				Reason: FaultLinkURLWhitespace,
			})
		}

		// TBD is link valid in form?

		// otherwise ok!
		idx = i + j
	}
	return faults
}

// runawayCodeFence looks for un-ended code fences
//
// returns -1 is code blocks seem ok
// returns idx of runaway code fence
//
func runawayCodeFence(raw []byte, faults []Fault) []Fault {
	codeFenceMarker := []byte{'`', '`', '`'}
	count := 0
	idx := 0
	last := 0
	for idx < len(raw) {
		i := bytes.Index(raw[idx:], codeFenceMarker)
		if i == -1 {
			break
		}
		if idx == 0 || raw[idx+i-1] == '\n' {
			count++
			last = idx + i
		}
		idx += i + len(codeFenceMarker) + 1
	}
	if count%2 == 1 {
		faults = append(faults, Fault{
			Offset: last,
			Reason: FaultRunawayCodeFence,
		})
	}
	return faults
}

// VetFunc defines a markdown vet function
type VetFunc func([]byte, []Fault) []Fault

// Vet is the main function to find structural problems with Markdown
func Vet(raw []byte) []Fault {
	faults := []Fault{}
	vetfuncs := []VetFunc{
		verifyURL,
		runawayCodeFence,
	}

	for _, fn := range vetfuncs {
		faults = fn(raw, faults)
	}

	if len(faults) == 0 {
		return nil
	}

	// sort by location first
	sort.Slice(faults, func(i, j int) bool { return faults[i].Offset < faults[j].Offset })

	// convert offsets to line numbers
	// very bad linear rescan!
	for i := range faults {
		row, col, line := GetLine(raw, faults[i].Offset)
		faults[i].Row = row
		faults[i].Column = col
		faults[i].Line = line
	}

	return faults
}
