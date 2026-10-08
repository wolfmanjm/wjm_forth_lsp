// forthdoc extracts newly defined words from Forth source code, together with
// the comment that documents each one.
//
// Usage:
//
//	forthdoc [flags] [file.fs ...]      (reads stdin if no files are given)
//
// Which comment documents a word?
//
//  1. A block of \ or ( ... ) comments directly above the definition
//     (no blank line in between, unless -loose is given).
//  2. A \ comment on the same line as the definition header, e.g.
//     ": square ( n -- n*n ) dup * ;  \ square a number"
//     or "42 constant answer \ the answer".
//
// A stack-effect comment directly after the name is reported separately.
//
// Recognised defining words: : CODE LABEL CREATE VARIABLE 2VARIABLE FVARIABLE
// CVARIABLE CONSTANT 2CONSTANT FCONSTANT VALUE 2VALUE FVALUE DEFER BUFFER:
// USER MARKER VOCABULARY SYNONYM.  Colon definitions that use CREATE are
// detected as user-defined defining words, so later uses such as
// "10 const ten" are found too.  More can be added with -defining.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"unicode"
)

// Word is one definition found in the source.
type Word struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Line      int    `json:"line"`
	Stack     string `json:"stack,omitempty"`
	Doc       string `json:"doc,omitempty"`
	Immediate bool   `json:"immediate,omitempty"`
	Defining  bool   `json:"defining,omitempty"` // a word that itself defines words
}

var builtinDefining = []string{
	":", "code", "label", "create", "variable", "2variable", "fvariable",
	"cvariable", "constant", "2constant", "fconstant", "value", "2value",
	"fvalue", "defer", "buffer:", "user", "marker", "vocabulary", "synonym",
}

// Words whose following token is data, not a new definition or syntax.
var skipNext = map[string]bool{
	"postpone": true, "[compile]": true, "[char]": true, "char": true,
	"'": true, "[']": true,
}

// ---------------------------------------------------------------- lexer ---

type token struct {
	text string
	line int
}

type lexer struct {
	src  string
	pos  int
	line int
}

func newLexer(src string) *lexer { return &lexer{src: src, line: 1} }

// next returns the next whitespace-delimited token.
func (l *lexer) next() (token, bool) {
	for l.pos < len(l.src) && l.src[l.pos] <= ' ' {
		if l.src[l.pos] == '\n' {
			l.line++
		}
		l.pos++
	}
	if l.pos >= len(l.src) {
		return token{}, false
	}
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] > ' ' {
		l.pos++
	}
	return token{l.src[start:l.pos], l.line}, true
}

// peek returns the next token without consuming it.
func (l *lexer) peek() (token, bool) {
	pos, line := l.pos, l.line
	t, ok := l.next()
	l.pos, l.line = pos, line
	return t, ok
}

// readUntil consumes the single delimiting blank after a parsing word, then
// everything up to and including delim (e.g. ')' or '"'). Returns the text.
func (l *lexer) readUntil(delim byte) string {
	if l.pos < len(l.src) && l.src[l.pos] <= ' ' {
		if l.src[l.pos] == '\n' {
			l.line++
		}
		l.pos++
	}
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != delim {
		if l.src[l.pos] == '\n' {
			l.line++
		}
		l.pos++
	}
	text := l.src[start:l.pos]
	if l.pos < len(l.src) {
		l.pos++
	}
	return text
}

// readLine consumes the rest of the current line (for \ comments).
func (l *lexer) readLine() string {
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
	text := l.src[start:l.pos]
	if l.pos < len(l.src) {
		l.pos++
		l.line++
	}
	return text
}

// readEscString consumes an escaped string such as s\" a \" b" .
func (l *lexer) readEscString() {
	if l.pos < len(l.src) && l.src[l.pos] <= ' ' {
		if l.src[l.pos] == '\n' {
			l.line++
		}
		l.pos++
	}
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if c == '\\' && l.pos+1 < len(l.src) {
			if l.src[l.pos+1] == '\n' {
				l.line++
			}
			l.pos += 2
			continue
		}
		if c == '\n' {
			l.line++
		}
		l.pos++
		if c == '"' {
			return
		}
	}
}

// --------------------------------------------------------------- parser ---

type parser struct {
	lx       *lexer
	loose    bool
	defining map[string]bool
	words    []*Word

	// Comment block waiting to be attached to the next definition.
	pending    []string
	pendingEnd int // last line of the pending comment block
	codeLine   int // line of the first code token after the block (0 = none yet)

	lastCodeLine int // line of the most recent non-comment token
	last         *Word

	inDef     bool
	cur       *Word // current colon definition (nil for :NONAME)
	curCreate bool  // saw CREATE inside the current definition
	skip      bool  // next token is data (after CHAR, ' ...)
}

func parse(src string, loose bool, extra []string) []*Word {
	p := &parser{lx: newLexer(src), loose: loose, defining: map[string]bool{}}
	for _, w := range builtinDefining {
		p.defining[w] = true
	}
	for _, w := range extra {
		if w = strings.ToLower(strings.TrimSpace(w)); w != "" {
			p.defining[w] = true
		}
	}
	p.run()
	return p.words
}

func (p *parser) run() {
	for {
		tok, ok := p.lx.next()
		if !ok {
			return
		}
		lower := strings.ToLower(tok.text)

		if p.skip { // e.g. the "x" in  char x
			p.skip = false
			p.noteCode(tok.line)
			continue
		}

		// Comments and strings: consume their text so nothing inside them
		// is mistaken for code.
		switch lower {
		case `\`, `\g`:
			text := p.lx.readLine()
			p.comment(tok.line, tok.line, text, tok.line == p.lastCodeLine)
			continue
		case "(":
			text := p.lx.readUntil(')')
			p.comment(tok.line, p.lx.line, text, tok.line == p.lastCodeLine)
			continue
		case ".(":
			p.lx.readUntil(')')
			p.noteCode(tok.line)
			continue
		case `."`, `s"`, `c"`, `abort"`:
			p.lx.readUntil('"')
			p.noteCode(tok.line)
			continue
		case `s\"`, `c\"`:
			p.lx.readEscString()
			p.noteCode(tok.line)
			continue
		}

		p.noteCode(tok.line)

		if p.inDef {
			switch {
			case lower == ";":
				p.endDef(tok.line)
			case lower == "create" || lower == "<builds":
				p.curCreate = true
			case skipNext[lower]:
				p.skip = true
			}
			continue
		}

		switch {
		case lower == "immediate":
			if p.last != nil {
				p.last.Immediate = true
			}
		case lower == ":noname":
			p.inDef, p.cur, p.curCreate = true, nil, false
			p.pending = nil
		case p.defining[lower]:
			p.define(tok, lower)
		case skipNext[lower]:
			p.skip = true
		}
	}
}

// noteCode is called for every non-comment token. It drops a pending comment
// block that is not directly followed by the statement being parsed.
func (p *parser) noteCode(line int) {
	if len(p.pending) > 0 {
		switch {
		case p.codeLine == 0:
			p.codeLine = line
			if !p.loose && line-p.pendingEnd > 1 {
				p.pending = nil // blank line(s) between comment and code
			}
		case line != p.codeLine:
			p.pending = nil
		}
	}
	p.lastCodeLine = line
}

// comment records a comment as a candidate doc block.
func (p *parser) comment(start, end int, text string, trailing bool) {
	if p.inDef || trailing {
		return // comments inside bodies or after code are not doc blocks
	}
	continuing := len(p.pending) > 0 && p.codeLine == 0 &&
		(p.loose || start-p.pendingEnd <= 1)
	if !continuing {
		p.pending = nil
	}
	for _, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" && !hasAlnum(ln) {
			ln = "" // ------ or ====== separator lines
		}
		if ln == "" {
			if n := len(p.pending); n > 0 && p.pending[n-1] != "" {
				p.pending = append(p.pending, "")
			}
			continue
		}
		p.pending = append(p.pending, ln)
	}
	p.pendingEnd = end
	p.codeLine = 0
}

// define handles a defining word: the next token is the new word's name.
func (p *parser) define(kw token, lower string) {
	nameTok, ok := p.lx.next()
	if !ok {
		return
	}
	w := &Word{Name: nameTok.text, Kind: strings.ToUpper(kw.text), Line: kw.line}
	w.Doc = joinDoc(p.pending)
	p.pending, p.codeLine = nil, 0

	if lower == "synonym" {
		p.lx.next() // the word being aliased
	}

	// Stack-effect comment right after the name.
	if t, ok := p.lx.peek(); ok && t.text == "(" && (t.line == p.lx.line || lower == ":") {
		p.lx.next()
		w.Stack = strings.Join(strings.Fields(p.lx.readUntil(')')), " ")
	}
	p.lastCodeLine = p.lx.line

	// Same-line \ comment after the header.
	if t, ok := p.lx.peek(); ok && isLineComment(t.text) && t.line == p.lx.line {
		p.lx.next()
		w.Doc = appendDoc(w.Doc, strings.TrimSpace(p.lx.readLine()))
	}

	p.words = append(p.words, w)
	p.last = w
	if lower == ":" {
		p.inDef, p.cur, p.curCreate = true, w, false
	}
}

func (p *parser) endDef(semiLine int) {
	p.inDef = false
	w := p.cur
	p.cur = nil
	if w == nil {
		return
	}
	if p.curCreate {
		w.Defining = true
		p.defining[strings.ToLower(w.Name)] = true
	}
	// One-line definition with a trailing comment:  : sq dup * ; \ square
	if w.Doc == "" && w.Line == semiLine {
		if t, ok := p.lx.peek(); ok && isLineComment(t.text) && t.line == semiLine {
			p.lx.next()
			w.Doc = strings.TrimSpace(p.lx.readLine())
		}
	}
}

// -------------------------------------------------------------- helpers ---

func isLineComment(s string) bool {
	s = strings.ToLower(s)
	return s == `\` || s == `\g`
}

func hasAlnum(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func joinDoc(lines []string) string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func appendDoc(doc, extra string) string {
	switch {
	case extra == "":
		return doc
	case doc == "":
		return extra
	}
	return doc + "\n" + extra
}

func kindLabel(w *Word) string {
	var s string
	if w.Kind == ":" {
		s = "colon definition"
	} else {
		s = strings.ToLower(w.Kind)
	}
	if w.Defining {
		s += ", defining word"
	}
	if w.Immediate {
		s += ", immediate"
	}
	return s
}

// --------------------------------------------------------------- output ---

func codeSpan(s string) string {
	if strings.Contains(s, "`") {
		return "`` " + s + " ``"
	}
	return "`" + s + "`"
}

func printMarkdown(out io.Writer, file string, words []*Word) {
	fmt.Fprintf(out, "# %s\n\n", file)
	for _, w := range words {
		fmt.Fprintf(out, "### %s\n\n", codeSpan(w.Name))
		meta := kindLabel(w) + fmt.Sprintf(" (line %d)", w.Line)
		if w.Stack != "" {
			meta = codeSpan("( "+w.Stack+" )") + " — " + meta
		}
		fmt.Fprintf(out, "%s\n\n", meta)
		if w.Doc != "" {
			fmt.Fprintf(out, "%s\n\n", w.Doc)
		}
	}
}

func printText(out io.Writer, file string, words []*Word) {
	fmt.Fprintf(out, "== %s ==\n", file)
	tw := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for _, w := range words {
		stack := ""
		if w.Stack != "" {
			stack = "( " + w.Stack + " )"
		}
		doc := strings.SplitN(w.Doc, "\n", 2)[0]
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", w.Name, strings.ToLower(w.Kind), stack, doc)
	}
	tw.Flush()
}

// ----------------------------------------------------------------- main ---

type fileResult struct {
	File  string  `json:"file"`
	Words []*Word `json:"words"`
}

func main() {
	format := flag.String("format", "md", "output format: md, text or json")
	loose := flag.Bool("loose", false, "allow blank lines between a comment block and its definition")
	extra := flag.String("defining", "", "comma-separated extra defining words (e.g. \"enum,field:\")")
	onlyDoc := flag.Bool("documented", false, "only list words that have a doc comment")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] [file.fs ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	type input struct{ name, src string }
	var inputs []input
	if flag.NArg() == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "forthdoc:", err)
			os.Exit(1)
		}
		inputs = append(inputs, input{"<stdin>", string(b)})
	}
	for _, name := range flag.Args() {
		b, err := os.ReadFile(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "forthdoc:", err)
			os.Exit(1)
		}
		inputs = append(inputs, input{name, string(b)})
	}

	var results []fileResult
	for _, in := range inputs {
		words := parse(in.src, *loose, strings.Split(*extra, ","))
		if *onlyDoc {
			kept := words[:0]
			for _, w := range words {
				if w.Doc != "" {
					kept = append(kept, w)
				}
			}
			words = kept
		}
		results = append(results, fileResult{in.name, words})
	}

	switch *format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		var v any = results
		if len(results) == 1 {
			v = results[0]
		}
		if err := enc.Encode(v); err != nil {
			fmt.Fprintln(os.Stderr, "forthdoc:", err)
			os.Exit(1)
		}
	case "text":
		for _, r := range results {
			printText(os.Stdout, r.File, r.Words)
		}
	case "md":
		for _, r := range results {
			printMarkdown(os.Stdout, r.File, r.Words)
		}
	default:
		fmt.Fprintf(os.Stderr, "forthdoc: unknown format %q\n", *format)
		os.Exit(2)
	}
}
