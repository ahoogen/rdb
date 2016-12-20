package rdb

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

var eof rune

// item represents a token and a literal value
type item struct {
	Token token
	Value string
}

type stateFn func(*lexer) stateFn

// lexer represents a query to be lexed and provides a buffered Reader for lexing.
type lexer struct {
	name   string
	input  string
	output string
	start  int
	pos    int
	width  int
	items  chan item
}

func lexSelect(l *lexer) stateFn {
	l.pos += len(selectStmt)
	l.emit(selectToken)

}

func lexStatement(l *lexer) stateFn {
	if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), selectStmt) {
		return lexSelect
	}
	if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), insertStmt) {
		return lexInsert
	}
	if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), fromStmt) {
		return lexFrom
	}
	l.dump()
	return nil
}

// NewLexer creates and returns a new lexer scanner from the provided Reader.
// Reader can be supplied by strings.NewReader(myString)
func lex(name, query string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: query,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := lexStatement; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t token) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) dump() {
	l.items <- item{unparsed, l.input[l.start:]}
	l.start = len(l.input)
	l.items <- item{Token: EOF}
}

// read is used to fetch the next rune in sequence from the buffered Reader.
func (l *lexer) read() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	var ch rune
	ch, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return ch
}

// unread places a previously read rune back into the buffer to be passed to the
// next call to Scan.
func (l *lexer) unread() {
	l.pos -= l.width
	_, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
}

// scan is the entrypoint into lexical scanning. scan will pass whitespace and
// alphanumeric sequences to whitespace and identifier scanners, or will return
// the token type of an individual special token.
func (l *lexer) scan() item {
	ch := l.read()
	if isWhitespace(ch) {
		l.unread()
		return l.scanWhitespace()
	} else if isNumeric(ch) || ch == '-' || ch == '+' {
		l.unread()
		return l.scanNumber()
	} else if isAlphanum(ch) || ch == '\'' || ch == '"' || ch == '`' {
		l.unread()
		return l.scanIdent()
	}

	switch ch {
	case eof:
		return item{EOF, ""}
	case '*':
		return item{astrisk, string(ch)}
	case ',':
		return item{comma, string(ch)}
	case '.':
		return item{period, string(ch)}
	case '(':
		return item{lParen, string(ch)}
	case ')':
		return item{rParen, string(ch)}
	case '{':
		return item{lBrace, string(ch)}
	case '}':
		return item{rBrace, string(ch)}
	case '=':
		return item{equals, string(ch)}
	}

	return item{illegal, string(ch)}
}

func (l *lexer) scanNumber() item {
	var isNegative, isDecimal, isScientific bool
	var buf bytes.Buffer

	for {
		ch := l.read()
		if ch == eof {
			break
		} else if ch == '+' {
			buf.WriteRune(ch)
		} else if ch == '-' {
			buf.WriteRune(ch)
			if len(buf.String()) == 1 {
				isNegative = true
			}
		} else if ch == '.' {
			buf.WriteRune(ch)
			isDecimal = true
		} else if ch == 'E' || ch == 'e' {
			buf.WriteRune(ch)
			isScientific = true
		} else if isNumeric(ch) {
			buf.WriteRune(ch)
		} else if isLetter(ch) || isWhitespace(ch) {
			l.unread()
			break
		}
	}

	var t = floatingPointNumber
	if !isDecimal && !isScientific && !isNegative {
		t = naturalNumber
	} else if !isDecimal && !isScientific {
		t = integer
	} else if !isScientific {
		t = fixedNumber
	}

	return item{t, buf.String()}
}

// scanWhitespace returns a whitespace token WS and a contiguous sequence of
// whitespace characters
func (l *lexer) scanWhitespace() item {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := l.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			l.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return item{WS, buf.String()}
}

// scanIdent fetches the next token from a lexing stream and returns the matched
// token type, or a token type of IDENTIFIER if it is not one of the SQL Keywords
// Any quoted identifier must begin and end with the same type of quote.
func (l *lexer) scanIdent() item {
	var buf bytes.Buffer
	var first, last rune
	first = l.read()
	buf.WriteRune(first)
	var quoted bool

	switch first {
	case '\'', '"', '`':
		quoted = true
	default:
		quoted = false
	}

	for {
		// Stop at EOF
		if ch := l.read(); ch == eof {
			break

			// Stop at first non alphanumeric if toekn isn't quoted
		} else if !quoted && !isAlphanum(ch) {
			l.unread()
			break

			// If we're quoted atop at first matching unescaped quote
		} else if quoted && ch == first {
			buf.WriteRune(ch)
			check := last
			last = ch
			next, err := l.r.Peek(1)

			// If Peek is EOF we're already at the end of the string and past
			// quote escaping checking
			if err != nil {
				break
			}

			// If we're at a quote
			if ch == '`' || ((check != '\\' && check != ch) && rune(next[0]) != ch) {
				break
			}

			// We're still inside the token, keep adding to buffer
		} else {
			last = ch
			buf.WriteRune(ch)
		}
	}

	if quoted && first != last {
		return item{illegal, buf.String()}
	}

	// The following represents a subset of keywords that MySQL supports.
	// TODO: Performance would benefit by moving to a btree type structure
	// It's a small list but still 60 tokens, drop check time to O(log(n)) from O(n)
	// and should this token list grow (for instance when string and math functions
	// are added), that'd be a good thing to have set up.
	switch strings.ToUpper(buf.String()) {
	case selectStmt:
		return item{selectToken, buf.String()}
	case insertStmt:
		return item{insertToken, buf.String()}
	case fromStmt:
		return item{fromToken, buf.String()}
	case partitionStmt:
		return item{partitionToken, buf.String()}
	case asStmt:
		return item{asToken, buf.String()}
	case straightJoinStmt:
		return item{straightJoinToken, buf.String()}
	case crossJoinStmt:
		return item{cross, buf.String()}
	case innerJoinStmt:
		return item{inner, buf.String()}
	case "OUTER":
		return item{outer, buf.String()}
	case "OJ":
		return item{oj, buf.String()}
	case "NATURAL":
		return item{natural, buf.String()}
	case "LEFT":
		return item{left, buf.String()}
	case "RIGHT":
		return item{right, buf.String()}
	case "JOIN":
		return item{join, buf.String()}
	case "WHERE":
		return item{where, buf.String()}
	case "VALUES", "VALUE":
		return item{values, buf.String()}
	case "SET":
		return item{set, buf.String()}
	case "DEFAULT":
		return item{defaultValue, buf.String()}
	case "ALL":
		return item{all, buf.String()}
	case "DISTINCT", "DISTINCTROW":
		return item{distinct, buf.String()}
	case "HIGH_PRIORITY":
		return item{highPriority, buf.String()}
	case "LOW_PRIORITY":
		return item{lowPriority, buf.String()}
	case "DELAYED":
		return item{delayed, buf.String()}
	case "MAX_STATEMENT_TIME":
		return item{maxStatementTime, buf.String()}
	case "SQL_SMALL_RESULT":
		return item{sqlSmallResult, buf.String()}
	case "SQL_BIG_RESULT":
		return item{sqlBigResult, buf.String()}
	case "SQL_BUFFER_RESULT":
		return item{sqlBufferResult, buf.String()}
	case "SQL_CACHE":
		return item{sqlCache, buf.String()}
	case "SQL_NO_CACHE":
		return item{sqlNoCache, buf.String()}
	case "SQL_CALC_FOUND_ROWS":
		return item{sqlCalcFoundRows, buf.String()}
	case "ON":
		return item{on, buf.String()}
	case "USING":
		return item{using, buf.String()}
	case "USE":
		return item{use, buf.String()}
	case "IGNORE":
		return item{ignore, buf.String()}
	case "FORCE":
		return item{force, buf.String()}
	case "INDEX":
		return item{index, buf.String()}
	case "KEY":
		return item{key, buf.String()}
	case "FOR":
		return item{forStmt, buf.String()}
	case "ORDER":
		return item{order, buf.String()}
	case "GROUP":
		return item{group, buf.String()}
	case "BY":
		return item{by, buf.String()}
	}

	var tok = identifier
	if quoted && first != '`' {
		tok = quotedString
	}

	return item{tok, buf.String()}
}

// isWhitespace is a helper function to identify whitespace characters within
// a lexing stream.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\n' || ch == '\t'
}

// isLetter is a helper function to identify alphabetic characters within
// a lexing stream.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isNumeric is a helper function to identify numeric characters within a
// lexing stream.
func isNumeric(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// isAlphanum is a helper function that composes numeric and alphabetic check functions
// along with accepting an underscore for the purposes of lexing identifiers
// within a lexing stream.
func isAlphanum(ch rune) bool {
	return isLetter(ch) || isNumeric(ch) || ch == '_'
}
