package main

import "strings"

type ParsedCmd struct {
	Args           []string
	RedirectStdout Redirect
	RedirectStderr Redirect
}

type Redirect struct {
	File string
	Type redirectOperation
}

type redirectTarget int

const (
	noRedirect redirectTarget = iota
	redirectStdout
	redirectStderr
)

type redirectOperation int

const (
	resultRedirect redirectOperation = iota
	resultAppend
)

// tokenize splits input into raw tokens, handling single quotes, double quotes, and backslash escaping. It does not interpret redirects.
func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]
		switch {
		case ch == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
		case ch == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
		case ch == '\\' && !inSingleQuote && !inDoubleQuote:
			i++
			if i < len(input) {
				current.WriteByte(input[i])
			}
		case ch == '\\' && inDoubleQuote:
			if i+1 < len(input) && (input[i+1] == '"' || input[i+1] == '\\') {
				i++
				current.WriteByte(input[i])
			} else {
				current.WriteByte(ch)
			}
		case (ch == ' ' || ch == '\t') && !inSingleQuote && !inDoubleQuote:
			flush()
		default:
			current.WriteByte(ch)
		}
	}
	flush()
	return tokens
}

// parseArgs builds a ParsedCmd by tokenizing input and then scanning tokens for redirect operators (>, >>, 2>, 2>>).
func parseArgs(input string) ParsedCmd {
	tokens := tokenize(input)
	var args []string
	parsed := ParsedCmd{}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		target, op, isRedir := parseRedirToken(tok)
		if isRedir {
			i++
			if i >= len(tokens) {
				break
			}
			r := Redirect{File: tokens[i], Type: op}
			switch target {
			case redirectStdout:
				parsed.RedirectStdout = r
			case redirectStderr:
				parsed.RedirectStderr = r
			}
			continue
		}

		args = append(args, tok)
	}

	parsed.Args = args
	return parsed
}

// parseRedirToken checks whether a token is a redirect operator and returns its target, operation type, and whether it matched at all.
func parseRedirToken(tok string) (redirectTarget, redirectOperation, bool) {
	switch tok {
	case ">", "1>":
		return redirectStdout, resultRedirect, true
	case ">>", "1>>":
		return redirectStdout, resultAppend, true
	case "2>":
		return redirectStderr, resultRedirect, true
	case "2>>":
		return redirectStderr, resultAppend, true
	}
	return noRedirect, resultRedirect, false
}

// splitPipeline splits a raw input string on unquoted '|' characters.
func splitPipeline(input string) []string {
	var segments []string
	var current strings.Builder
	inSingle, inDouble := false, false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		switch {
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
			current.WriteByte(ch)
		case ch == '"' && !inSingle:
			inDouble = !inDouble
			current.WriteByte(ch)
		case ch == '|' && !inSingle && !inDouble:
			segments = append(segments, current.String())
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}
	segments = append(segments, current.String())
	return segments
}
