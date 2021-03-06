package mdtool

import (
	"testing"
)

type testcase struct {
	input  string
	faults []Fault
}

var cases = []testcase{
	// nothing
	{
		input:  "",
		faults: []Fault{},
	},

	// code
	{
		input:  "```\ncode\n```\n",
		faults: []Fault{},
	},

	// runaway code
	{
		input: "```\ncode\n",
		faults: []Fault{
			{
				Reason: FaultRunawayCodeFence,
			},
		},
	},
	// runaway code
	{
		input: "something\n```bash\ncode\n```\nsomething\n```\ncode\n",
		faults: []Fault{
			{
				Reason: FaultRunawayCodeFence,
			},
		},
	},
	// normal url
	{
		input:  "[text](http://golang.org/)",
		faults: []Fault{},
	},
	// url runaway
	{
		input: "[text](http://golang.org/",
		faults: []Fault{
			{
				Reason: FaultRunawayLinkURL,
			},
		},
	},
	// url runaway
	{
		input: "[text\n\n",
		faults: []Fault{
			{
				Reason: FaultRunawayLinkText,
			},
		},
	},
	// url runaway
	{
		input: "[line1\n\nline2](http://golang.org/)",
		faults: []Fault{
			{
				Reason: FaultLinkTextWhitespace,
			},
		},
	},
	// url runaway
	{
		input: "[line1](http://golang.\n\norg/)",
		faults: []Fault{
			{
				Reason: FaultLinkURLWhitespace,
			},
		},
	},
	// trailing whitespace after code fense
	{
		input: "```\ncode\n``` \nsomething\n",
		faults: []Fault{
			{
				Reason: FaultCodeFenceTrailingWhitespace,
			},
		},
	},
	// trailing whitespace  after code fence plus tag
	{
		input:  "``` go\ncode\n```\nsomething\n",
		faults: []Fault{},
	},
}

func TestVet(t *testing.T) {
	for i, tt := range cases {
		faults := Vet([]byte(tt.input))
		if len(faults) != len(tt.faults) {
			t.Errorf("%d: %q want %d faults got %d",
				i, tt.input, len(tt.faults), len(faults))
		}
	}
}
