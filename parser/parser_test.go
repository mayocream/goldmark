package parser_test

import (
	"testing"

	"github.com/yuin/goldmark/parser"
)

func Test_parser_Any(t *testing.T) {
	t.Run("state", func(t *testing.T) {
		t.Logf("if requre paragragh 1: %v", parser.HasChildren&parser.RequireParagraph != 0)
		t.Logf("if requre paragragh 2: %v", (parser.HasChildren|parser.RequireParagraph)&parser.RequireParagraph != 0)
	})
}
