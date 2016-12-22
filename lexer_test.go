package rdb

import (
	"strings"
	"testing"
)

func TestIsWhitespace(t *testing.T) {
	if !isWhitespace(rune(' ')) {
		t.Errorf("Space failed whitespace test.")
	}
	if !isWhitespace(rune('\n')) {
		t.Errorf("Newline failed whitespace test.")
	}
	if !isWhitespace(rune('\t')) {
		t.Errorf("Tab failed whitespace test.")
	}

	if isWhitespace(rune('z')) {
		t.Errorf("Non-whitespace character passed whitespace test.")
	}
}

func TestIsLetter(t *testing.T) {
	if !isLetter(rune('a')) {
		t.Errorf("'a' failed is letter test")
	}
	if !isLetter(rune('z')) {
		t.Errorf("'z' failed is letter test")
	}
	if !isLetter(rune('A')) {
		t.Errorf("'A' failed is letter test")
	}
	if !isLetter(rune('Z')) {
		t.Errorf("'Z' failed is letter test")
	}
	if isLetter(rune('&')) {
		t.Errorf("'&' passed is letter test")
	}
}

func TestIsNumeric(t *testing.T) {
	for i := '0'; i <= '9'; i++ {
		if !isNumeric(i) {
			t.Errorf("'%s' failed isNumeric test", string(i))
		}
	}
	if isNumeric(rune('a')) {
		t.Errorf("'a' passed isNumeric test")
	}
}

func TestIsAlphanum(t *testing.T) {
	al := "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alr := strings.NewReader(al)
	for i := 0; i < len(al); i++ {
		r, _, _ := alr.ReadRune()
		if !isAlphanum(r) {
			t.Errorf("'%s' failed alphanumeric test", string(r))
		}
	}
	nal := "'\"`~!@#$%^&*()+=-[]\\|}{:;?/.,<>}"
	nalr := strings.NewReader(nal)
	for i := 0; i < len(nal); i++ {
		r, _, _ := nalr.ReadRune()
		if isAlphanum(r) {
			t.Errorf("'%s' failed alphanumeric test", string(r))
		}
	}
}

func TestNewScanner(t *testing.T) {
	str := "I am a happy new string reader"
	s := lex(str)
	for i := 0; i < len(str); i++ {
		ch := s.read()
		if string(ch) != str[i:i+1] {
			t.Errorf("Expected: %s Got: %s", str[i:i+1], string(ch))
		}
	}
}

func TestRead(t *testing.T) {
	str := "I am a happy new string reader"
	s := lex(str)
	for i := 0; i < len(str); i++ {
		ch := s.read()
		if string(ch) != str[i:i+1] {
			t.Errorf("Expected: %s Got: %s", str[i:i+1], string(ch))
		}
	}

	ch := s.read()
	if ch != eof {
		t.Errorf("Expected EOF: Got: %+v", ch)
	}
}

func TestUnread(t *testing.T) {
	str := "abcd"
	s := lex(str)
	var ch rune
	ch = s.read()
	if ch != 'a' {
		t.Errorf("Expected: a Got: %s", string(ch))
	}
	ch = s.read()
	if ch != 'b' {
		t.Errorf("Expected: b Got: %s", string(ch))
	}
	s.unread()
	ch = s.read()
	if ch != 'b' {
		t.Errorf("Expected: b Got: %s", string(ch))
	}
	ch = s.read()
	if ch != 'c' {
		t.Errorf("Expected: c Got: %s", string(ch))
	}
	ch = s.read()
	if ch != 'd' {
		t.Errorf("Expected: d Got: %s", string(ch))
	}
}

func TestScanWhitespace(t *testing.T) {
	str := "   \n \t\t   \n \t"
	s := lex(str)
	item := s.scan()
	if item.Token != WS {
		t.Errorf("Failed to get expected whitespace token, got IOTA %d instead", int(item.Token))
	}
	if item.Value != str {
		t.Errorf("Expected: %q Got: %q", str, item.Value)
	}
}

func TestScanWhitespaceStops(t *testing.T) {
	str := "   \n \t\tiamnotwhitespace   \n \t"
	s := lex(str)
	item := s.scan()
	if item.Token != WS {
		t.Errorf("Failed to get expected whitespace token, got IOTA %d instead", int(item.Token))
	}
	if item.Value != str[0:len(item.Value)] {
		t.Errorf("Expected: %q Got: %q", str[0:len(item.Value)], item.Value)
	}
}

func TestScanGetsEOF(t *testing.T) {
	str := ""
	s := lex(str)
	item := s.scan()
	if item.Token != EOF {
		t.Errorf("Expected %d, got %s", EOF, item.Value)
	}
}

func TestScanGetsAstrisk(t *testing.T) {
	str := "*"
	s := lex(str)
	item := s.scan()
	if item.Token != astrisk {
		t.Errorf("Expected %d, got %d", astrisk, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsComma(t *testing.T) {
	str := ","
	s := lex(str)
	item := s.scan()
	if item.Token != comma {
		t.Errorf("Expected %d, got %d", comma, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsPeriod(t *testing.T) {
	str := "."
	s := lex(str)
	item := s.scan()
	if item.Token != period {
		t.Errorf("Expected %d, got %d", period, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsLParen(t *testing.T) {
	str := "("
	s := lex(str)
	item := s.scan()
	if item.Token != lParen {
		t.Errorf("Expected %d, got %d", lParen, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsRParen(t *testing.T) {
	str := ")"
	s := lex(str)
	item := s.scan()
	if item.Token != rParen {
		t.Errorf("Expected %d, got %d", rParen, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsLBrace(t *testing.T) {
	str := "{"
	s := lex(str)
	item := s.scan()
	if item.Token != lBrace {
		t.Errorf("Expected %d, got %d", lBrace, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsRBrace(t *testing.T) {
	str := "}"
	s := lex(str)
	item := s.scan()
	if item.Token != rBrace {
		t.Errorf("Expected RBRACE, got %d", item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsIllegal(t *testing.T) {
	str := "@"
	s := lex(str)
	item := s.scan()
	if item.Token != illegal {
		t.Errorf("Expected %d, got %d", illegal, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected %s, got %s", str, item.Value)
	}
}

func TestScanGetsIdentifier(t *testing.T) {
	str := "iamnotakeyword"
	s := lex(str)
	item := s.scan()
	if item.Token != identifier {
		t.Errorf("Expected %d, got %d", identifier, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanGetsQuotedIdentifier(t *testing.T) {
	str := "`iamnotakeyword`"
	s := lex(str)
	item := s.scan()
	if item.Token != identifier {
		t.Errorf("Expected %d, got %d", identifier, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanGetsIllegalMismatchedQuote(t *testing.T) {
	str := "`iamnotakeyword'"
	s := lex(str)
	item := s.scan()
	if item.Token != illegal {
		t.Errorf("Expected %d, got %d", illegal, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanEndsAtFirstMatchedQuote(t *testing.T) {
	str := "`iamnota`keyword'"
	s := lex(str)
	item := s.scan()
	if item.Token != identifier {
		t.Errorf("Expected %d, got %d", identifier, item.Token)
	}
	if item.Value != str[0:len(item.Value)] {
		t.Errorf("Expected: %s Got: %s", str[0:len(item.Value)], item.Value)
	}
}

func TestScanHandlesEscapedSQuoteSquote(t *testing.T) {
	str := "'iamnota''keyword'"
	s := lex(str)
	item := s.scan()
	if item.Token != quotedString {
		t.Errorf("Expected %d, got %d", quotedString, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanHandlesEscapedBslashSquote(t *testing.T) {
	str := "'iamnota\\'keyword'"
	s := lex(str)
	item := s.scan()
	if item.Token != quotedString {
		t.Errorf("Expected %d, got %d", quotedString, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanHandlesEscapedBslashDquote(t *testing.T) {
	str := `"iamnota\\"keyword"`
	s := lex(str)
	item := s.scan()
	if item.Token != quotedString {
		t.Errorf("Expected %d, got %d", quotedString, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanHandlesEscapedDquoteDquote(t *testing.T) {
	str := `"iamnota""keyword"`
	s := lex(str)
	item := s.scan()
	if item.Token != quotedString {
		t.Errorf("Expected %d, got %d", quotedString, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanHandlesQuotedQuoteChars(t *testing.T) {
	str := "'iam`nota``keyw''\"ord'"
	s := lex(str)
	item := s.scan()
	if item.Token != quotedString {
		t.Errorf("Expected %d, got %d", quotedString, item.Token)
	}
	if item.Value != str {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
}

func TestScanEndsAtNonAlphanum(t *testing.T) {
	// While this would result in an SQL syntax error,
	// RDB is NOT an SQL syntax checker, and doesn't guarantee that it should
	// be able to handle your shitty code.
	// This test produces (as far as RDB is concerned) valid token IDENTIFIER iam
	// And the token ILLEGAL mismatch quoted string `nota keyword'
	str := "iam`nota keyword'"
	s := lex(str)
	item := s.scan()
	if item.Token != identifier {
		t.Errorf("Expected %d, got %s", identifier, item.Value)
	}
	if item.Value != "iam" {
		t.Errorf("Expected: %s Got: %s", str[0:len(item.Value)], item.Value)
	}
	item = s.scan()
	if item.Token != illegal {
		t.Errorf("Expected: %d (%s) Got: %d (%s)", illegal, "`nota keyword'", item.Token, item.Value)
	}
}

func TestScanNumberStopsAtWhitespace(t *testing.T) {
	str := "3.14159 2.22e4"
	s := lex(str)
	item := s.scan()
	if item.Token != fixedNumber {
		t.Errorf("Expected %d, got %d", fixedNumber, item.Token)
	}
	if item.Value != "3.14159" {
		t.Errorf("Expected: %s Got: %s", str, item.Value)
	}
	item = s.scan()
	if item.Token != WS {
		t.Errorf("Expected whitespace token!")
	}
}

func TestKeywordTokens(t *testing.T) {
	strs := []string{
		// Special tokens ILLEGAL, EOF, WS
		"@", "", " ",

		// Numbers
		"+12", "-12", "3.14", "3.14e0",

		// IDENTIFIER token
		"`foofoo`",

		// QUOTED_STRING
		`'I am a ''quoted\' string'`,

		// ASTRISK, COMMA, PERIOD, LPAREN, RPAREN, LBRACE, RBRACE
		"*", ",", ".", "(", ")", "{", "}", "=",

		// Keywords
		"SELECT", "INSERT", "FROM", "PARTITION", "AS", "STRAIGHT_JOIN", "CROSS JOIN",
		"INNER JOIN", "OJ", "NATURAL JOIN", "NATURAL LEFT JOIN", "NATURAL LEFT OUTER JOIN",
		"NATURAL RIGHT JOIN", "NATURAL RIGHT OUTER JOIN", "LEFT JOIN", "LEFT OUTER JOIN",
		"RIGHT JOIN", "RIGHT OUTER JOIN", "USE INDEX", "USE KEY", "IGNORE INDEX", "IGNORE KEY",
		"FORCE INDEX", "FORCE KEY", "FOR JOIN", "FOR ORDER BY", "FOR GROUP BY", "WHERE",
		"VALUES", "SET", "DEFAULT", "ALL", "DISTINCT", "HIGH_PRIORITY", "LOW_PRIORITY",
		"DELAYED", "MAX_STATEMENT_TIME", "SQL_SMALL_RESULT", "SQL_BIG_RESULT",
		"SQL_BUFFER_RESULT", "SQL_CACHE", "SQL_NO_CACHE", "SQL_CALC_FOUND_ROWS",
		"ON", "USING", "ORDER BY", "GROUP BY",
	}
	for i, str := range strs {
		s := lex(str)
		item := s.scan()
		if item.Token != token(i) {
			t.Errorf("Expected token %d (%s) but got token %d (%s)", i, strs[i], item.Token, item.Value)
		}
		if item.Value != str {
			t.Errorf("Expected toekn string %s but got token string %s", str, item.Value)
		}
	}
}
