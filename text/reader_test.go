package text

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/yuin/goldmark/util"
)

// " "  = 0100 0000 = 20
// "\n" = 0000 1010 = 0A
var testText = []byte("あ this is an example paragragh\n next line \n\n\n\n\r last line\n")

func testReader() Reader {
	return NewReader(testText)
}

func Test_reader_Any(t *testing.T) {
	r := testReader()

	t.Run("space", func(t *testing.T) {
		s := []byte("\n") // \s is also space
		is := util.IsSpace(s[0])
		t.Logf(`\n is space: %v`, is)

		t.Logf("text: \n%s", string(util.VisualizeSpaces(testText)))
	})

	t.Run("peekLine0", func(t *testing.T) {
		line, seg := r.PeekLine()
		t.Log("line: ", string(util.VisualizeSpaces(line)),
			"\nseg: ", spew.Sdump(seg),
		)
	})

	t.Run("rune", func(t *testing.T) {
		ru, size, _ := r.ReadRune()
		t.Log("rune: ", ru,
			"\nstring: ", string(ru),
			"\nsize: ", size,
		)
	})

	t.Run("peekLine1", func(t *testing.T) {
		line, seg := r.PeekLine()
		t.Log("line: ", string(util.VisualizeSpaces(line)),
			"\nseg: ", spew.Sdump(seg),
		)
	})

	t.Run("peekLine2", func(t *testing.T) {
		r.AdvanceLine()
		r.AdvanceLine()
		// 跳过了 3 个空行
		seg, lines, ok := r.SkipBlankLines()
		t.Log("seg: ", spew.Sdump(seg),
			"\nlines: ", lines, // 返回 3
			"\nok: ", ok,
		)
	})

}

func Test_reader_SkipBlankLines(t *testing.T) {
	r := testReader()
	seg, lines, ok := r.SkipBlankLines()
	t.Log("seg: ", spew.Sdump(seg),
		"\nlines: ", lines,
		"\nok: ", ok,
	)
}

func Test_reader_Peek(t *testing.T) {
	r := testReader()
	b := r.Peek()
	t.Log("byte: ", b, "\nstring: ", string(b))
}
