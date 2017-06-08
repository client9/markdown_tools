package mdtool

import (
	bf "gopkg.in/russross/blackfriday.v2"
)

// ASTNode is a serializable abstract syntax tree node
type ASTNode struct {
	Type     string
	Literal  string      `json:",omitempty"`
	Attr     interface{} `json:",omitempty"`
	Children []*ASTNode  `json:",omitempty"`
}

// NewASTNode converts a BlackFriday AST node into a serializable format
func NewASTNode(node *bf.Node) *ASTNode {
	a := &ASTNode{
		Type:    node.Type.String(),
		Literal: string(node.Literal),
	}

	switch node.Type {
	case bf.Heading:
		a.Attr = &node.HeadingData
	case bf.List, bf.Item:
		a.Attr = &node.ListData
	case bf.Link, bf.Image:
		a.Attr = &node.LinkData
	case bf.CodeBlock:
		a.Attr = &node.CodeBlockData
	case bf.Table, bf.TableHead, bf.TableBody, bf.TableRow, bf.TableCell:
		a.Attr = &node.TableCellData
	}
	for child := node.FirstChild; child != nil; child = child.Next {
		a.Children = append(a.Children, NewASTNode(child))
	}
	return a
}

// Ast takes an input and returns a JSON-friendly ASTNode
func Ast(src []byte, opts ...bf.Option) *ASTNode {
	proc := bf.DefaultProcessor()
	parser := proc.NewParser(opts...)
	node := parser.Parse(src)
	return NewASTNode(node)
}
