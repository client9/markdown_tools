package vet

import (
	"bytes"
	"log"
	"sort"
)

var (
	newlines = []byte{'\n', '\n'}
)

const (
	FaultZero = iota
	FaultRunawayCodeFence
	FaultRunawayLinkText
	FaultRunawayLinkURL
	FaultLinkTextWhitespace
	FaultLinkURLWhitespace
)

type Fault struct {
	Line   int
	Column int
	Offset int
	Reason int
}

func (f *Fault) ComputeLine(raw []byte) {


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
		log.Printf("GOT DESC: %s", desc)
		i += j + 1

		// skip space and tabs
		for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t') {
			i++
		}
		log.Printf("emaining: %s", raw[i:])
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

		aurl := raw[i:i+j]
		log.Printf("URL: %s", aurl)
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
		log.Printf("IDX=%d", idx)
		i := bytes.Index(raw[idx:], codeFenceMarker)
		if i == -1 {
			log.Printf("codefense done")
			break
		}
		if idx == 0 || raw[idx + i -1] == '\n' {
			count++
			last = idx + i
			log.Printf("CF: Got one at %d", last)
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
	return faults
}
