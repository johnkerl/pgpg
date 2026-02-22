package asts

import (
	"testing"

	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
)

func TestNewAST(t *testing.T) {
	root := NewASTNode(nil, NodeType("root"), nil)
	ast := NewAST(root)
	if ast == nil || ast.RootNode != root {
		t.Errorf("NewAST: got %+v", ast)
	}
}

func TestNewASTNode(t *testing.T) {
	// Terminal (children nil)
	n := NewASTNode(nil, NodeType("id"), nil)
	if n == nil || n.Type != NodeType("id") || n.Children != nil {
		t.Errorf("NewASTNode terminal: got %+v", n)
	}
	// With children
	child := NewASTNode(nil, NodeType("x"), nil)
	parent := NewASTNode(nil, NodeType("seq"), []*ASTNode{child})
	if parent == nil || len(parent.Children) != 1 || parent.Children[0] != child {
		t.Errorf("NewASTNode with children: got %+v", parent)
	}
}

func TestNewASTNodeTerminal(t *testing.T) {
	tok := tokens.NewToken([]rune("42"), tokens.TokenType("int"), tokens.NewTokenLocation())
	n := NewASTNodeTerminal(tok, NodeType("number"))
	if n == nil || n.Token != tok || n.Children != nil {
		t.Errorf("NewASTNodeTerminal: got %+v", n)
	}
}

func TestWithChildPrepended(t *testing.T) {
	parent := NewASTNode(nil, NodeType("p"), nil)
	child := NewASTNode(nil, NodeType("c"), nil)
	out := WithChildPrepended(parent, child)
	if out != parent || len(parent.Children) != 1 || parent.Children[0] != child {
		t.Errorf("WithChildPrepended: got %+v", parent)
	}
	// Prepend again
	child2 := NewASTNode(nil, NodeType("c2"), nil)
	WithChildPrepended(parent, child2)
	if len(parent.Children) != 2 || parent.Children[0] != child2 || parent.Children[1] != child {
		t.Errorf("WithChildPrepended second: got %+v", parent.Children)
	}
}

func TestWithChildAppended(t *testing.T) {
	parent := NewASTNode(nil, NodeType("p"), nil)
	child := NewASTNode(nil, NodeType("c"), nil)
	out := WithChildAppended(parent, child)
	if out != parent || len(parent.Children) != 1 || parent.Children[0] != child {
		t.Errorf("WithChildAppended: got %+v", parent)
	}
}

func TestWithChildrenAdopted(t *testing.T) {
	parent := NewASTNode(nil, NodeType("p"), nil)
	child := NewASTNode(nil, NodeType("c"), []*ASTNode{
		NewASTNode(nil, NodeType("a"), nil),
		NewASTNode(nil, NodeType("b"), nil),
	})
	out := WithChildrenAdopted(parent, child)
	if out != parent || len(parent.Children) != 2 || len(child.Children) != 0 {
		t.Errorf("WithChildrenAdopted: parent.Children=%v child.Children=%v", parent.Children, child.Children)
	}
}

func TestCheckArity(t *testing.T) {
	good := NewASTNode(nil, NodeType("x"), []*ASTNode{
		NewASTNode(nil, NodeType("a"), nil),
		NewASTNode(nil, NodeType("b"), nil),
	})
	if err := good.CheckArity(2); err != nil {
		t.Errorf("CheckArity(2) on 2 children: %v", err)
	}
	if err := good.CheckArity(1); err == nil {
		t.Error("CheckArity(1) on 2 children: expected error")
	}
}
