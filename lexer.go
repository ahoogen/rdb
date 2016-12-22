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

	return lexFrom
}

func lexInsert(l *lexer) stateFn {
	l.pos += len(insertStmt)
	l.emit(insertToken)

	return lexValues
}

func lexFrom(l *lexer) stateFn {
	l.pos += len(fromStmt)
	l.emit(fromToken)

	return lexDump
}

func lexValues(l *lexer) stateFn {

	return lexDump
}

func lexDump(l *lexer) stateFn {
	l.dump()
	return nil
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
func lex(query string) *lexer {
	l := &lexer{
		input: query,
		//items: make(chan item),
	}
	// go l.run()
	return l //, l.items
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
	} else if isLetter(ch) {
		l.unread()
		return l.scanKeyword()
	} else if ch == '\'' || ch == '"' || ch == '`' {
		l.unread()
		return l.scanQuoted()
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
	buf.WriteRune(l.read())

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
func (l *lexer) scanQuoted() item {
	var buf bytes.Buffer
	var first, last rune
	first = l.read()
	buf.WriteRune(first)

	for {
		// Stop at EOF
		if ch := l.read(); ch == eof {
			break

			// If we're quoted atop at first matching unescaped quote
		} else if ch == first {
			buf.WriteRune(ch)
			check := last
			last = ch
			next := l.read()
			l.unread()

			// If Peek is EOF we're already at the end of the string and past
			// quote escaping checking
			if next == eof {
				break
			}

			// If we're at a quote
			if ch == '`' || ((check != '\\' && check != ch) && next != ch) {
				break
			}

			// We're still inside the token, keep adding to buffer
		} else {
			last = ch
			buf.WriteRune(ch)
		}
	}

	if first != last {
		return item{illegal, buf.String()}
	}

	var tok = identifier
	if first != '`' {
		tok = quotedString
	}

	return item{tok, buf.String()}
}

func (l *lexer) scanKeyword() item {
	if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), selectStmt) {
		l.pos += len(selectStmt)
		return item{selectToken, selectStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), insertStmt) {
		return item{insertToken, insertStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), fromStmt) {
		return item{fromToken, fromStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), partitionStmt) {
		return item{partitionToken, partitionStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), asStmt) {
		return item{asToken, asStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), straightJoinStmt) {
		return item{straightJoinToken, straightJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), crossJoinStmt) {
		return item{crossJoinToken, crossJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), innerJoinStmt) {
		return item{innerJoinToken, innerJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), ojStmt) {
		return item{ojToken, ojStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), naturalJoinStmt) {
		return item{naturalJoinToken, naturalJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), naturalLeftJoinStmt) {
		return item{naturalLeftJoinToken, naturalLeftJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), naturalLeftOuterJoinStmt) {
		return item{naturalLeftOuterJoinToken, naturalLeftOuterJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), naturalRightJoinStmt) {
		return item{naturalRightJoinToken, naturalRightJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), naturalRightOuterJoinStmt) {
		return item{naturalRightOuterJoinToken, naturalRightOuterJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), leftJoinStmt) {
		return item{leftJoinToken, leftJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), leftOuterJoinStmt) {
		return item{leftOuterJoinToken, leftOuterJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), rightJoinStmt) {
		return item{rightJoinToken, rightJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), rightOuterJoinStmt) {
		return item{rightOuterJoinToken, rightOuterJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), useIndexStmt) {
		return item{useIndexToken, useIndexStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), useKeyStmt) {
		return item{useKeyToken, useKeyStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), ignoreIndexStmt) {
		return item{ignoreIndexToken, ignoreIndexStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), ignoreKeyStmt) {
		return item{ignoreKeyToken, ignoreKeyStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), forceIndexStmt) {
		return item{forceIndexToken, forceIndexStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), forceKeyStmt) {
		return item{forceKeyToken, forceKeyStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), forJoinStmt) {
		return item{forJoinToken, forJoinStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), forOrderByStmt) {
		return item{forOrderByToken, forOrderByStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), forGroupByStmt) {
		return item{forGroupByToken, forGroupByStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), whereStmt) {
		return item{whereToken, whereStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), valuesStmt) {
		return item{valuesToken, valuesStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), setStmt) {
		return item{setToken, setStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), defaultStmt) {
		return item{defaultToken, defaultStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), allStmt) {
		return item{allToken, allStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), distinctStmt) {
		return item{distinctToken, distinctStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), highPriorityStmt) {
		return item{highPriorityToken, highPriorityStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), lowPriorityStmt) {
		return item{lowPriorityToken, lowPriorityStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), delayedStmt) {
		return item{delayedToken, delayedStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), maxStatementTimeStmt) {
		return item{maxStatementTimeToken, maxStatementTimeStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), sqlSmallResultStmt) {
		return item{sqlSmallResultToken, sqlSmallResultStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), sqlBigResultStmt) {
		return item{sqlBigResultToken, sqlBigResultStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), sqlBufferResultStmt) {
		return item{sqlBufferResultToken, sqlBufferResultStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), sqlCacheStmt) {
		return item{sqlCacheToken, sqlCacheStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), sqlNoCacheStmt) {
		return item{sqlNoCacheToken, sqlNoCacheStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), sqlCalcFoundRowsStmt) {
		return item{sqlCalcFoundRowsToken, sqlCalcFoundRowsStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), onStmt) {
		return item{onToken, onStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), usingStmt) {
		return item{usingToken, usingStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), orderByStmt) {
		return item{orderByToken, orderByStmt}
	} else if strings.HasPrefix(strings.ToUpper(l.input[l.pos:]), groupByStmt) {
		return item{groupByToken, groupByStmt}
	}

	var buf bytes.Buffer
	for {
		ch := l.read()
		if !isAlphanum(ch) {
			l.unread()
			break
		}
		buf.WriteRune(ch)
	}

	return item{identifier, buf.String()}
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
