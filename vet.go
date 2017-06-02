package mdtool

import (
	"bytes"
	"sort"
)

var (
	newlines = []byte{'\n', '\n'}
)

type FaultType int

const (
	FaultZero = FaultType(0)
	FaultRunawayCodeFence = FaultType(1)
	FaultRunawayLinkText = FaultType(2)
	FaultRunawayLinkURL = FaultType(3)
	FaultLinkTextWhitespace = FaultType(4)
	FaultLinkURLWhitespace = FaultType(5)
)

func (s FaultType) String() string {
	switch s {
		case FaultRunawayCodeFence: return "Runaway Code Fence"
		case FaultRunawayLinkText: return "Runaway Lint Text"
		case FaultRunawayLinkURL: return "Runaway Link URL"
		case FaultLinkTextWhitespace: return "Link Text with Whitespace"
		case FaultLinkURLWhitespace: return "Link URL with Whitespace"
	}
	return "FAIL"
}

type Fault struct {
	Offset int
	Reason FaultType
	Row    int
	Column int
	Line   string
}

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
	for i := 0; i < len(raw); i++ {
		i := bytes.IndexByte(raw[i:], '[')
		if i == -1 {
			break
		}
		start := i
		i++
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
		for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t') {
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
		i = j
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

type VetFunc func([]byte, []Fault) []Fault

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
