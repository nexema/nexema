package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Tokenizer struct {
	reader *bufio.Reader
	pos    Position // current reader position
	ch     rune     // current character
	lit    string   // string representation of ch for debugging
}

type Position struct {
	fileName string // the name of the file, if applies
	offset   int    // the offset (character count)
	line     int    // the line
}

const (
	eof rune = -1
)

// String returns the string representation of a Position
func (p Position) String() string {
	if p.fileName != "" {

		return fmt.Sprintf("%s %d:%d", p.fileName, p.line, p.offset)
	}

	return fmt.Sprintf("%d:%d", p.line, p.offset)
}

func NewTokenizer(r *bufio.Reader) *Tokenizer {
	return &Tokenizer{
		reader: r,
		pos:    Position{offset: -1, line: 1},
	}
}

// Scan scans the next token and returns it, with its literal representation
func (t *Tokenizer) Scan() (pos Position, tok Token, lit string, err error) {
	// advance one position
	t.scan()

	// scan skipping whitespaces
	t.skipWhitespace()
	pos = t.pos

	switch ch := t.ch; {
	case isLetter(ch):
		lit, err = t.scanIdentifier()

		// check if its a keyword or a literal
		tok = GetKeyword(lit)

	// decimal or .8 float syntax
	case isDecimal(ch) || (ch == '.' && isDecimal(t.peek())):
		tok, lit, err = t.scanNumber()

	default:
		t.scan() // scan
		switch ch {
		case eof:
			tok = Token_EOF

		case '\n':
			tok, lit = Token_Whitespace, "\n"

		case '"':
			t.unscan()
			tok = Token_String
			lit, err = t.scanString()

		case '=':
			tok, lit = Token_Assign, string(ch)

		case '?':
			tok, lit = Token_Nullable, string(ch)

		case '(':
			tok, lit = Token_Lparen, string(ch)

		case ')':
			tok, lit = Token_Rparen, string(ch)

		case '[':
			tok, lit = Token_Lbrack, string(ch)

		case ']':
			tok, lit = Token_Rbrack, string(ch)

		case '{':
			tok, lit = Token_Lbrace, string(ch)

		case '}':
			tok, lit = Token_Rbrace, string(ch)

		case ':':
			tok, lit = Token_Colon, string(ch)

		case ',':
			tok, lit = Token_Comma, string(ch)

		case '.':
			tok, lit = Token_Period, string(ch)

		case '@':
			tok, lit = Token_At, string(ch)

		case '/':
			if t.ch == '/' || t.ch == '*' { // the next read char (// and /*)
				t.unscan()
				lit, err = t.scanComment()
				if err != nil {
					return
				}

				tok = Token_Comment
			}

		default:
			tok, lit = Token_Illegal, ""
		}
	}

	return
}

// scan moves the reader one position and sets the corresponding
// into t.ch
func (t *Tokenizer) scan() error {
	ch, _, err := t.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			t.ch = eof
		}

		return err
	}

	t.pos.offset++
	switch ch {
	case '\n':
		t.pos.line++
		t.pos.offset = 0

	case 0:
		return t.err("non valid character NULL terminator")
	}

	t.ch = ch
	t.lit = string(ch)
	return nil
}

// goes back 1 position in the reader
func (t *Tokenizer) unscan() {
	t.reader.UnreadRune()
	t.pos.offset--
}

// peek returns the next rune after t.pos.offset without advancing
// the scanner
func (t *Tokenizer) peek() rune {
	buf, err := t.reader.Peek(1)
	if err != nil {
		return eof
	}

	return rune(buf[0])
}

// consume advances the scanner n positions
func (t *Tokenizer) consume(n int) {
	t.reader.Discard(n)
}

// err reports an error in the current line and column
func (t *Tokenizer) err(txt string) error {
	return fmt.Errorf("%s -> %s", t.pos.String(), txt)
}

// scanIdentifier scans a valid string identifier. It must be called when t.ch
// is a valid letter
func (t *Tokenizer) scanIdentifier() (string, error) {
	buf := new(bytes.Buffer)

	// push the current ch to the buf
	buf.WriteRune(t.ch)

	// read until we have an invalid letter
	for {
		err := t.scan()
		if err != nil {
			if t.ch == eof {
				break
			}

			return "", err
		}

		// if its a letter or a digit, append to buffer
		if isLetter(t.ch) || isDigit(t.ch) {
			buf.WriteRune(t.ch)
		} else {
			break
		}
	}

	return buf.String(), nil
}

// scanNumber scans a valid number identifier. It must be called when t.ch is a valid digit
func (t *Tokenizer) scanNumber() (tok Token, lit string, err error) {
	buf := new(bytes.Buffer)
	tok = Token_Int

	// push the current ch into the buffer
	buf.WriteRune(t.ch)

	if t.ch == '.' {
		tok = Token_Float
	}

	for {
		err := t.scan()
		if err != nil {
			if t.ch == eof {
				break
			}

			return Token_Illegal, "", err
		}

		if isDecimal(t.ch) {
			buf.WriteRune(t.ch)
			continue
		} else if t.ch == '.' && tok != Token_Float { // avoid to read multiple periods
			tok = Token_Float
			buf.WriteRune('.')
			continue
		} else {
			break
		}

	}

	return tok, buf.String(), nil
}

// skipWhitespace scan tokens skipping any whitespace character
func (t *Tokenizer) skipWhitespace() error {
	for t.ch == ' ' || t.ch == '\t' || t.ch == '\n' || t.ch == '\r' {
		err := t.scan()
		if err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

// scanComment scans a //-style comment or a /*-style comment. It must be called when t.ch is /.
// The final literal is trimmed on both sides
func (t *Tokenizer) scanComment() (string, error) {
	buf := new(bytes.Buffer)

	// get next rune because initial / is already read
	t.scan()

	if t.ch == '/' {
		// read until new line is found
		for {
			t.scan()
			if t.ch == '\n' || t.ch == eof {
				goto exit
			}

			buf.WriteRune(t.ch)
		}
	}

	if t.ch == '*' {
		// read until */ is found
		for {
			t.scan()
			if t.ch == eof {
				break
			}

			if t.ch == '*' && t.peek() == '/' {
				// consume last / and break
				t.consume(1)
				goto exit // */ is not part of the comment literal
			}

			buf.WriteRune(t.ch)
		}
	}

	return "", t.err("comment not terminated")

exit:
	return strings.TrimSpace(buf.String()), nil
}

// scanString reads a string in the form "text". Expects first " is read
func (t *Tokenizer) scanString() (string, error) {
	buf := new(bytes.Buffer)
	for {
		t.scan()

		ch := t.ch
		if ch == '\n' || ch == eof {
			return "", t.err(`string literal expects to be closed with the " character`)
		}

		if ch == '"' {
			break
		}

		// check if its escaping the " character
		if ch == '\\' && t.peek() == '"' {
			buf.WriteRune('"')
			t.scan()
		} else {
			buf.WriteRune(ch)
		}
	}

	return buf.String(), nil
}

// isLetter returns true when ch is a valid UTF-8 a-zA-Z_
func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

// isDigit returns true when ch is a valid UTF-8 0-9
func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

// lower returns ch as lowercase
func lower(ch rune) rune {
	return ('a' - 'A') | ch
}

// isDecimal returns true if ch is between 0 and 9
func isDecimal(ch rune) bool {
	return '0' <= ch && ch <= '9'
}