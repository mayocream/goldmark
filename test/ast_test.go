package ast_test

import (
	"testing"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func Test_parser_Any(t *testing.T) {
	t.Run("paragragh", func(t *testing.T) {
		t.Logf("is paragragh 1: %v", ast.IsParagraph(&ast.Paragraph{}))
		t.Logf("is paragragh 2: %v", ast.IsParagraph(&ast.Heading{}))
		t.Logf("is paragragh 3: %v", ast.IsParagraph(east.NewDefinitionList(1, ast.NewParagraph())))
	})
}
