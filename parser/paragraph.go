package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type paragraphParser struct {
}

var defaultParagraphParser = &paragraphParser{}

// NewParagraphParser returns a new BlockParser that
// parses paragraphs.
func NewParagraphParser() BlockParser {
	return defaultParagraphParser
}

func (b *paragraphParser) Trigger() []byte {
	return nil
}

// paragragh 打开新的 block
func (b *paragraphParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	_, segment := reader.PeekLine()
	segment = segment.TrimLeftSpace(reader.Source())
	if segment.IsEmpty() {
		// 如果是空行, 不添加任何元素
		return nil, NoChildren
	}
	// 创建 paragragh
	node := ast.NewParagraph()
	// 添加行元素, 之后再解析行元素
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return node, NoChildren
}

// 继续解析
func (b *paragraphParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	_, segment := reader.PeekLine()
	segment = segment.TrimLeftSpace(reader.Source())
	// 上一行有 inline 内容, 本行为空行
	if segment.IsEmpty() {
		return Close
	}
	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)
	return Continue | NoChildren
}

func (b *paragraphParser) Close(node ast.Node, reader text.Reader, pc Context) {
	parent := node.Parent()
	if parent == nil {
		// paragraph has been transformed
		return
	}
	lines := node.Lines()
	if lines.Len() != 0 {
		// trim trailing spaces
		length := lines.Len()
		lastLine := node.Lines().At(length - 1)
		node.Lines().Set(length-1, lastLine.TrimRightSpace(reader.Source()))
	}
	if lines.Len() == 0 {
		node.Parent().RemoveChild(node.Parent(), node)
		return
	}
}

func (b *paragraphParser) CanInterruptParagraph() bool {
	return false
}

func (b *paragraphParser) CanAcceptIndentedLine() bool {
	return false
}
