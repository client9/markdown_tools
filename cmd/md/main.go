package main

import (
	"encoding/json"
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
	astCommand         = kingpin.Command("ast", "dump JSON representation of AST")
	vetCommand         = kingpin.Command("vet", "vet markdown structure")
	vetFiles           = vetCommand.Arg("files", "file to process, if none use stdin").Strings()
	fmtCommand         = kingpin.Command("fmt", "reformat markdown")
	fmt2Command        = kingpin.Command("fmt2", "reformat markdown, take 2")
	fmtWrite           = fmtCommand.Flag("write", "write in place").Short('w').Bool()
	fmtFiles           = fmtCommand.Arg("files", "file to process, if none use stdin").Strings()
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
	case "ast":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		node := mdtool.Ast(rawin)
		rawout, err := json.MarshalIndent(node, "", "   ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(rawout))
		return
	case "fmt2":
		rawin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		out := mdtool.Fmt2(rawin)
		fmt.Println(string(out))
		return
	case "fmt":
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

		if len(*fmtFiles) == 0 {
			rawin, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			out := mdtool.Fmt(rawin, &opt)
			fmt.Println(string(out))
			return
		}
		for _, name := range *fmtFiles {
			rawin, err := ioutil.ReadFile(name)
			if err != nil {
				log.Fatalf("Can't read %q: %s", name, err)
			}
			out := mdtool.Fmt(rawin, &opt)

			if !*fmtWrite {
				fmt.Println(string(out))
				continue
			}
			ioutil.WriteFile(name, out, 0)
		}
	case "vet":
		errCount := 0
		if len(*vetFiles) == 0 {
			rawin, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			faults := mdtool.Vet(rawin)

			for _, f := range faults {
				errCount++
				fmt.Printf("%d:%d offset=%d reason=%s %q\n", f.Row, f.Column, f.Offset, f.Reason, f.Line)
			}
			if len(faults) > 0 {
				os.Exit(2)
			}
		}
		for _, name := range *vetFiles {
			rawin, err := ioutil.ReadFile(name)
			if err != nil {
				log.Fatalf("Can't read %q: %s", name, err)
			}
			faults := mdtool.Vet(rawin)
			for _, f := range faults {
				errCount++
				fmt.Printf("%s:%d:%d offset=%d reason=%s %q\n", name, f.Row, f.Column, f.Offset, f.Reason, f.Line)
			}
		}
		if errCount > 0 {
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
		case "html2":
			rawout = mdtool.RenderHTML2(rawin)
		case "github":
			rawout = mdtool.RenderGitHub(rawin)
		default:
			log.Fatalf("Unknown render type %q", *renderType)
		}
		os.Stdout.Write(rawout)
	}
}
