package grammar

import (
	"fmt"
)

type bnfParser struct {
	lex *bnfLexer
	cur bnfToken
}

func newBNFParser(input []byte) (*bnfParser, error) {
	lex := newBNFLexer(input)
	first, err := lex.next()
	if err != nil {
		return nil, err
	}
	return &bnfParser{lex: lex, cur: first}, nil
}

func (p *bnfParser) parseGrammar() (*Grammar, error) {
	var rules []Rule
	var start Symbol
	for {
		p.skipEOL()
		if p.cur.kind == bnfTokEOF {
			break
		}
		lhs, rhsList, err := p.parseRule()
		if err != nil {
			return nil, err
		}
		if len(rules) == 0 {
			start = lhs
		}
		for _, rhs := range rhsList {
			rules = append(rules, Rule{LHS: lhs, RHS: rhs})
		}
	}
	if len(rules) == 0 {
		return nil, p.errorf("no rules found")
	}
	return &Grammar{Start: start, Rules: rules}, nil
}

func (p *bnfParser) parseRule() (Symbol, [][]Symbol, error) {
	lhsTok, err := p.expect(bnfTokIdent)
	if err != nil {
		return Symbol{}, nil, err
	}
	lhs := Symbol{Name: lhsTok.lexeme, Kind: Nonterminal}
	if _, err := p.expect(bnfTokAssign); err != nil {
		return Symbol{}, nil, err
	}

	var rhsList [][]Symbol
	for {
		rhs, err := p.parseSequence()
		if err != nil {
			return Symbol{}, nil, err
		}
		rhsList = append(rhsList, rhs)
		if p.cur.kind != bnfTokPipe {
			break
		}
		if err := p.advance(); err != nil {
			return Symbol{}, nil, err
		}
	}

	if err := p.consumeRuleTerminator(); err != nil {
		return Symbol{}, nil, err
	}
	return lhs, rhsList, nil
}

func (p *bnfParser) parseSequence() ([]Symbol, error) {
	var symbols []Symbol
	for {
		switch p.cur.kind {
		case bnfTokIdent:
			symbols = append(symbols, Symbol{Name: p.cur.lexeme, Kind: Nonterminal})
			if err := p.advance(); err != nil {
				return nil, err
			}
		case bnfTokString:
			symbols = append(symbols, Symbol{Name: p.cur.lexeme, Kind: Terminal})
			if err := p.advance(); err != nil {
				return nil, err
			}
		default:
			return symbols, nil
		}
	}
}

func (p *bnfParser) consumeRuleTerminator() error {
	switch p.cur.kind {
	case bnfTokSemi, bnfTokEOL:
		if err := p.advance(); err != nil {
			return err
		}
		p.skipEOL()
		return nil
	case bnfTokEOF:
		return nil
	default:
		return p.errorf("expected rule terminator")
	}
}

func (p *bnfParser) skipEOL() {
	for p.cur.kind == bnfTokEOL || p.cur.kind == bnfTokSemi {
		_ = p.advance()
	}
}

func (p *bnfParser) expect(kind bnfTokenKind) (bnfToken, error) {
	if p.cur.kind != kind {
		return bnfToken{}, p.errorf("expected %s", tokenKindName(kind))
	}
	tok := p.cur
	return tok, p.advance()
}

func (p *bnfParser) advance() error {
	next, err := p.lex.next()
	if err != nil {
		return err
	}
	p.cur = next
	return nil
}

func (p *bnfParser) errorf(format string, args ...any) error {
	return fmt.Errorf("bnf parse %d:%d: %s", p.cur.line, p.cur.col, fmt.Sprintf(format, args...))
}

func tokenKindName(kind bnfTokenKind) string {
	switch kind {
	case bnfTokEOF:
		return "EOF"
	case bnfTokIdent:
		return "identifier"
	case bnfTokString:
		return "string"
	case bnfTokAssign:
		return "::="
	case bnfTokPipe:
		return "|"
	case bnfTokSemi:
		return ";"
	case bnfTokEOL:
		return "end-of-line"
	default:
		return "unknown"
	}
}
