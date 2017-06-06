package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/client9/markdown_tools"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "unreleased"
)

var (
	versionCommand     = kingpin.Command("version", "show version and exit")
	vetCommand         = kingpin.Command("vet", "vet markdown structure")
	fmtCommand         = kingpin.Command("fmt", "reformat markdown")
	fmtLineLength      = fmtCommand.Flag("linelength", "line length, -1=unlimited").Default("70").Int()
	fmtHrLength        = fmtCommand.Flag("hrlength", "HR length").Default("3").Int()
	fmtHrChar          = fmtCommand.Flag("hrchar", "HR char").Default("-").String()
	fmtListIndent      = fmtCommand.Flag("listindent", "list indent").Default("  ").String()
	fmtListBulletChar  = fmtCommand.Flag("listbullet", "list bullet").Default("-").String()
	fmtListBulletSpace = fmtCommand.Flag("listbulletspace", "list bullet space").Default(" ").String()

	renderCommand = kingpin.Command("render", "render markdown to another format")
	renderType    = renderCommand.Arg("type", "render type").Default("html").String()
)

func main() {
	switch kingpin.Parse() {
	case "version":
		fmt.Println(version)
		os.Exit(2)
	case "fmt":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		bulletSpace := strings.Replace(*fmtListBulletSpace, "\\t", "\t", -1)
		bulletIndent := strings.Replace(*fmtListIndent, "\\t", "\t", -1)

		opt := mdtool.FmtOptions{
			LineLength:      *fmtLineLength,
			HrChar:          *fmtHrChar,
			HrLength:        *fmtHrLength,
			ListBulletChar:  *fmtListBulletChar,
			ListIndent:      bulletIndent,
			ListBulletSpace: bulletSpace,
		}
		out := mdtool.Fmt(rawin, &opt)
		fmt.Println(string(out))
	case "vet":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		faults := mdtool.Vet(rawin)
		for _, f := range faults {
			fmt.Printf("%d:%d offset=%d reason=%s %q\n", f.Row, f.Column, f.Offset, f.Reason, f.Line)
		}
		if len(faults) > 0 {
			os.Exit(2)
		}
	case "render":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		var rawout []byte
		switch *renderType {
		case "html":
			rawout = mdtool.RenderHTML(rawin)
		case "github":
			rawout = mdtool.RenderGitHub(rawin)
		default:
			log.Fatalf("Unknown render type %q", *renderType)
		}
		os.Stdout.Write(rawout)
	}
}
