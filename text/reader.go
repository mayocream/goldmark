package text

import (
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/yuin/goldmark/util"
)

const invalidValue = -1

// EOF indicates the end of file.
const EOF = byte(0xff)

// A Reader interface provides abstracted method for reading text.
type Reader interface {
	io.RuneReader

	// Source returns a source of the reader.
	Source() []byte

	// ResetPosition resets positions.
	ResetPosition()

	// Peek returns a byte at current position without advancing the internal pointer.
	Peek() byte

	// PeekLine returns the current line without advancing the internal pointer.
	PeekLine() ([]byte, Segment)

	// PrecendingCharacter returns a character just before current internal pointer.
	PrecendingCharacter() rune

	// Value returns a value of the given segment.
	Value(Segment) []byte

	// LineOffset returns a distance from the line head to current position.
	LineOffset() int

	// Position returns current line number and position.
	Position() (int, Segment)

	// SetPosition sets current line number and position.
	SetPosition(int, Segment)

	// SetPadding sets padding to the reader.
	SetPadding(int)

	// Advance advances the internal pointer.
	Advance(int)

	// AdvanceAndSetPadding advances the internal pointer and add padding to the
	// reader.
	AdvanceAndSetPadding(int, int)

	// AdvanceLine advances the internal pointer to the next line head.
	AdvanceLine()

	// SkipSpaces skips space characters and returns a non-blank line.
	// If it reaches EOF, returns false.
	SkipSpaces() (Segment, int, bool)

	// SkipSpaces skips blank lines and returns a non-blank line.
	// If it reaches EOF, returns false.
	SkipBlankLines() (Segment, int, bool)

	// Match performs regular expression matching to current line.
	Match(reg *regexp.Regexp) bool

	// Match performs regular expression searching to current line.
	FindSubMatch(reg *regexp.Regexp) [][]byte
}

type reader struct {
	source       []byte
	sourceLength int
	line         int
	peekedLine   []byte
	pos          Segment
	head         int
	lineOffset   int
}

// NewReader return a new Reader that can read UTF-8 bytes .
func NewReader(source []byte) Reader {
	r := &reader{
		source:       source,
		sourceLength: len(source),
	}
	r.ResetPosition()
	return r
}

// 清空标志位
func (r *reader) ResetPosition() {
	// 记录读取的行数, 因为是直接读取 bytes, 需要手动分割记录行数
	r.line = -1
	r.head = 0
	r.lineOffset = -1
	// 初始化的时候就会读取一整行
	r.AdvanceLine()
}

func (r *reader) Source() []byte {
	return r.source
}

// 输入 col 位置标记 , 返回值
func (r *reader) Value(seg Segment) []byte {
	return seg.Value(r.source)
}

func (r *reader) Peek() byte {
	// 判断有效的 pos 位置
	if r.pos.Start >= 0 && r.pos.Start < r.sourceLength {
		if r.pos.Padding != 0 {
			// 返回空格符
			return space[0]
		}
		// 返回有效字节
		return r.source[r.pos.Start]
	}
	// 读取完成
	return EOF
}

// 读取一行
func (r *reader) PeekLine() ([]byte, Segment) {
	if r.pos.Start >= 0 && r.pos.Start < r.sourceLength {
		// 没有已经 Peek 的
		if r.peekedLine == nil {
			r.peekedLine = r.pos.Value(r.Source())
		}
		return r.peekedLine, r.pos
	}
	return nil, r.pos
}

// io.RuneReader interface
func (r *reader) ReadRune() (rune, int, error) {
	return readRuneReader(r)
}

func (r *reader) LineOffset() int {
	if r.lineOffset < 0 {
		v := 0
		for i := r.head; i < r.pos.Start; i++ {
			if r.source[i] == '\t' {
				v += util.TabWidth(v)
			} else {
				v++
			}
		}
		r.lineOffset = v - r.pos.Padding
	}
	return r.lineOffset
}

func (r *reader) PrecendingCharacter() rune {
	if r.pos.Start <= 0 {
		if r.pos.Padding != 0 {
			return rune(' ')
		}
		return rune('\n')
	}
	i := r.pos.Start - 1
	for ; i >= 0; i-- {
		if utf8.RuneStart(r.source[i]) {
			break
		}
	}
	rn, _ := utf8.DecodeRune(r.source[i:])
	return rn
}

// 前进 n 个长度
func (r *reader) Advance(n int) {
	r.lineOffset = -1
	// 如果在一行之内, 直接移动游标
	if n < len(r.peekedLine) && r.pos.Padding == 0 {
		// 前进 n 个 byte
		r.pos.Start += n
		r.peekedLine = nil // 这里不太理解为什么要清空当前行
		return
	}
	// 清空当前行
	r.peekedLine = nil
	l := r.sourceLength
	// 将游标右移 n 位，若途中碰到换行符则读取下一行并将游标移动
	for ; n > 0 && r.pos.Start < l; n-- {
		if r.pos.Padding != 0 {
			r.pos.Padding--
			continue
		}
		// 如果是换行符就下一行
		if r.source[r.pos.Start] == '\n' {
			r.AdvanceLine()
			continue
		}
		// 游标右移
		r.pos.Start++
	}
}

func (r *reader) AdvanceAndSetPadding(n, padding int) {
	r.Advance(n)
	if padding > r.pos.Padding {
		r.SetPadding(padding)
	}
}

// 读取一整行并将 Pos 标志前进
func (r *reader) AdvanceLine() {
	r.lineOffset = -1
	r.peekedLine = nil
	r.pos.Start = r.pos.Stop
	r.head = r.pos.Start
	// 默认 stop = 0
	if r.pos.Start < 0 {
		return
	}
	// 遍历文件
	r.pos.Stop = r.sourceLength
	for i := r.pos.Start; i < r.sourceLength; i++ {
		// 获取字符
		c := r.source[i]
		// 获取一整行, 以换行符结尾
		// 例如: "paragraph\n"
		if c == '\n' {
			r.pos.Stop = i + 1
			break
		}
	}
	r.line++
	r.pos.Padding = 0
}

func (r *reader) Position() (int, Segment) {
	return r.line, r.pos
}

func (r *reader) SetPosition(line int, pos Segment) {
	r.lineOffset = -1
	r.line = line
	r.pos = pos
}

func (r *reader) SetPadding(v int) {
	r.pos.Padding = v
}

func (r *reader) SkipSpaces() (Segment, int, bool) {
	return skipSpacesReader(r)
}

func (r *reader) SkipBlankLines() (Segment, int, bool) {
	return skipBlankLinesReader(r)
}

func (r *reader) Match(reg *regexp.Regexp) bool {
	return matchReader(r, reg)
}

func (r *reader) FindSubMatch(reg *regexp.Regexp) [][]byte {
	return findSubMatchReader(r, reg)
}

// A BlockReader interface is a reader that is optimized for Blocks.
type BlockReader interface {
	Reader
	// Reset resets current state and sets new segments to the reader.
	Reset(segment *Segments)
}

type blockReader struct {
	source         []byte
	segments       *Segments
	segmentsLength int
	line           int
	pos            Segment
	head           int
	last           int
	lineOffset     int
}

// NewBlockReader returns a new BlockReader.
func NewBlockReader(source []byte, segments *Segments) BlockReader {
	r := &blockReader{
		source: source,
	}
	if segments != nil {
		r.Reset(segments)
	}
	return r
}

func (r *blockReader) ResetPosition() {
	r.line = -1
	r.head = 0
	r.last = 0
	r.lineOffset = -1
	r.pos.Start = -1
	r.pos.Stop = -1
	r.pos.Padding = 0
	if r.segmentsLength > 0 {
		last := r.segments.At(r.segmentsLength - 1)
		r.last = last.Stop
	}
	r.AdvanceLine()
}

func (r *blockReader) Reset(segments *Segments) {
	r.segments = segments
	r.segmentsLength = segments.Len()
	r.ResetPosition()
}

func (r *blockReader) Source() []byte {
	return r.source
}

func (r *blockReader) Value(seg Segment) []byte {
	line := r.segmentsLength - 1
	ret := make([]byte, 0, seg.Stop-seg.Start+1)
	for ; line >= 0; line-- {
		if seg.Start >= r.segments.At(line).Start {
			break
		}
	}
	i := seg.Start
	for ; line < r.segmentsLength; line++ {
		s := r.segments.At(line)
		if i < 0 {
			i = s.Start
		}
		ret = s.ConcatPadding(ret)
		for ; i < seg.Stop && i < s.Stop; i++ {
			ret = append(ret, r.source[i])
		}
		i = -1
		if s.Stop > seg.Stop {
			break
		}
	}
	return ret
}

// io.RuneReader interface
func (r *blockReader) ReadRune() (rune, int, error) {
	return readRuneReader(r)
}

func (r *blockReader) PrecendingCharacter() rune {
	if r.pos.Padding != 0 {
		return rune(' ')
	}
	if r.segments.Len() < 1 {
		return rune('\n')
	}
	firstSegment := r.segments.At(0)
	if r.line == 0 && r.pos.Start <= firstSegment.Start {
		return rune('\n')
	}
	l := len(r.source)
	i := r.pos.Start - 1
	for ; i < l && i >= 0; i-- {
		if utf8.RuneStart(r.source[i]) {
			break
		}
	}
	if i < 0 || i >= l {
		return rune('\n')
	}
	rn, _ := utf8.DecodeRune(r.source[i:])
	return rn
}

func (r *blockReader) LineOffset() int {
	if r.lineOffset < 0 {
		v := 0
		for i := r.head; i < r.pos.Start; i++ {
			if r.source[i] == '\t' {
				v += util.TabWidth(v)
			} else {
				v++
			}
		}
		r.lineOffset = v - r.pos.Padding
	}
	return r.lineOffset
}

func (r *blockReader) Peek() byte {
	if r.line < r.segmentsLength && r.pos.Start >= 0 && r.pos.Start < r.last {
		if r.pos.Padding != 0 {
			return space[0]
		}
		return r.source[r.pos.Start]
	}
	return EOF
}

func (r *blockReader) PeekLine() ([]byte, Segment) {
	if r.line < r.segmentsLength && r.pos.Start >= 0 && r.pos.Start < r.last {
		return r.pos.Value(r.source), r.pos
	}
	return nil, r.pos
}

func (r *blockReader) Advance(n int) {
	r.lineOffset = -1

	if n < r.pos.Stop-r.pos.Start && r.pos.Padding == 0 {
		r.pos.Start += n
		return
	}

	for ; n > 0; n-- {
		if r.pos.Padding != 0 {
			r.pos.Padding--
			continue
		}
		if r.pos.Start >= r.pos.Stop-1 && r.pos.Stop < r.last {
			r.AdvanceLine()
			continue
		}
		r.pos.Start++
	}
}

func (r *blockReader) AdvanceAndSetPadding(n, padding int) {
	r.Advance(n)
	if padding > r.pos.Padding {
		r.SetPadding(padding)
	}
}

func (r *blockReader) AdvanceLine() {
	r.SetPosition(r.line+1, NewSegment(invalidValue, invalidValue))
	r.head = r.pos.Start
}

func (r *blockReader) Position() (int, Segment) {
	return r.line, r.pos
}

func (r *blockReader) SetPosition(line int, pos Segment) {
	r.lineOffset = -1
	r.line = line
	if pos.Start == invalidValue {
		if r.line < r.segmentsLength {
			s := r.segments.At(line)
			r.head = s.Start
			r.pos = s
		}
	} else {
		r.pos = pos
		if r.line < r.segmentsLength {
			s := r.segments.At(line)
			r.head = s.Start
		}
	}
}

func (r *blockReader) SetPadding(v int) {
	r.lineOffset = -1
	r.pos.Padding = v
}

func (r *blockReader) SkipSpaces() (Segment, int, bool) {
	return skipSpacesReader(r)
}

func (r *blockReader) SkipBlankLines() (Segment, int, bool) {
	return skipBlankLinesReader(r)
}

func (r *blockReader) Match(reg *regexp.Regexp) bool {
	return matchReader(r, reg)
}

func (r *blockReader) FindSubMatch(reg *regexp.Regexp) [][]byte {
	return findSubMatchReader(r, reg)
}

func skipBlankLinesReader(r Reader) (Segment, int, bool) {
	lines := 0
	for {
		// 读取当前的一行
		line, seg := r.PeekLine()
		if line == nil {
			// 已经到内容结尾了，最后一行
			return seg, lines, false
		}
		// 判断一行是否都是空字符串, 包含 \n \s 和空格
		if util.IsBlank(line) {
			lines++
			r.AdvanceLine()
		} else {
			return seg, lines, true
		}
	}
}

func skipSpacesReader(r Reader) (Segment, int, bool) {
	// 标记空格数量
	chars := 0
	for {
		// 读取当前的一行
		line, segment := r.PeekLine()
		if line == nil {
			// 已经到内容结尾了，最后一行
			return segment, chars, false
		}
		// 遍历每个 byte
		for i, c := range line {
			// 是否是空格
			if util.IsSpace(c) {
				// 计数增加
				chars++
				// 指针前进
				r.Advance(1)
				continue
			}
			return segment.WithStart(segment.Start + i + 1), chars, true
		}
	}
}

func matchReader(r Reader, reg *regexp.Regexp) bool {
	oldline, oldseg := r.Position()
	match := reg.FindReaderSubmatchIndex(r)
	r.SetPosition(oldline, oldseg)
	if match == nil {
		return false
	}
	r.Advance(match[1] - match[0])
	return true
}

func findSubMatchReader(r Reader, reg *regexp.Regexp) [][]byte {
	oldline, oldseg := r.Position()
	match := reg.FindReaderSubmatchIndex(r)
	r.SetPosition(oldline, oldseg)
	if match == nil {
		return nil
	}
	runes := make([]rune, 0, match[1]-match[0])
	for i := 0; i < match[1]; {
		r, size, _ := readRuneReader(r)
		i += size
		runes = append(runes, r)
	}
	result := [][]byte{}
	for i := 0; i < len(match); i += 2 {
		result = append(result, []byte(string(runes[match[i]:match[i+1]])))
	}

	r.SetPosition(oldline, oldseg)
	r.Advance(match[1] - match[0])
	return result
}

func readRuneReader(r Reader) (rune, int, error) {
	// 读取一行
	line, _ := r.PeekLine()
	if line == nil {
		return 0, 0, io.EOF
	}
	// 解析 utf8
	rn, size := utf8.DecodeRune(line)
	if rn == utf8.RuneError {
		return 0, 0, io.EOF
	}
	// 指针前进
	r.Advance(size)
	return rn, size, nil
}
