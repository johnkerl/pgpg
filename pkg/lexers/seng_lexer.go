package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
)

const (
	SENGLexerTypeNoun                       = 1
	SENGLexerTypeAdjective                  = 2
	SENGLexerTypeArticle                    = 3
	SENGLexerTypeTransitiveVerb             = 4
	SENGLexerTypeIntransitiveVerb           = 5
	SENGLexerTypeTransitiveImperativeVerb   = 6
	SENGLexerTypeIntransitiveImperativeVerb = 7
	SENGLexerTypeAdverb                     = 8
	SENGLexerTypePreposition                = 9
)

var sengLexicon = map[string]tokens.TokenType{
	"dog":   SENGLexerTypeNoun,
	"cat":   SENGLexerTypeNoun,
	"fox":   SENGLexerTypeNoun,
	"mouse": SENGLexerTypeNoun,
	"food":  SENGLexerTypeNoun,
	"book":  SENGLexerTypeNoun,

	"red":   SENGLexerTypeAdjective,
	"green": SENGLexerTypeAdjective,
	"brown": SENGLexerTypeAdjective,
	"quick": SENGLexerTypeAdjective,
	"lazy":  SENGLexerTypeAdjective,

	"the": SENGLexerTypeArticle,
	"a":   SENGLexerTypeArticle,

	"goes":  SENGLexerTypeTransitiveVerb,
	"puts":  SENGLexerTypeTransitiveVerb,
	"reads": SENGLexerTypeTransitiveVerb,
	"eats":  SENGLexerTypeTransitiveVerb,

	"walks":  SENGLexerTypeIntransitiveVerb,
	"runs":   SENGLexerTypeIntransitiveVerb,
	"sleeps": SENGLexerTypeIntransitiveVerb,
	"jumps":  SENGLexerTypeIntransitiveVerb,

	"go":   SENGLexerTypeIntransitiveImperativeVerb,
	"jump": SENGLexerTypeIntransitiveImperativeVerb,

	"put":  SENGLexerTypeTransitiveImperativeVerb,
	"read": SENGLexerTypeTransitiveImperativeVerb,
	"eat":  SENGLexerTypeTransitiveImperativeVerb,

	"quickly": SENGLexerTypeAdverb,
	"slowly":  SENGLexerTypeAdverb,

	"under": SENGLexerTypePreposition,
	"over":  SENGLexerTypePreposition,
}

// SENGLexer is for the SENG grammar. It delegated to the WordLexer, but then augments this
// by adding token-type (i.e. part-of-speech) information.
type SENGLexer struct {
	wordLexer AbstractLexer
}

func NewSENGLexer(inputText string) AbstractLexer {
	return &SENGLexer{
		wordLexer: NewWordLexer(inputText),
	}
}

func (lexer *SENGLexer) Scan() (token *tokens.Token) {

	token = lexer.wordLexer.Scan()
	if token.IsEOF() || token.IsError() {
		return token
	}

	stringLexeme := string(token.Lexeme)
	tokenType, ok := sengLexicon[stringLexeme]
	if !ok {
		return tokens.NewErrorToken(
			fmt.Sprintf("SENG lexer: unrecognized token %q", stringLexeme),
			&token.Location,
		)
	}

	token.Type = tokenType
	return token
}

func (lexer *SENGLexer) DecodeType(tokenType tokens.TokenType) (string, error) {
	switch tokenType {
	case tokens.TokenTypeEOF:
		return "EOF", nil
	case tokens.TokenTypeError:
		return "error", nil

	case SENGLexerTypeNoun:
		return "noun", nil
	case SENGLexerTypeAdjective:
		return "adjective", nil
	case SENGLexerTypeArticle:
		return "article", nil
	case SENGLexerTypeTransitiveVerb:
		return "transitive verb", nil
	case SENGLexerTypeIntransitiveVerb:
		return "intransitive verb", nil
	case SENGLexerTypeTransitiveImperativeVerb:
		return "transitive imperative verb", nil
	case SENGLexerTypeIntransitiveImperativeVerb:
		return "intransitive imperative verb", nil
	case SENGLexerTypeAdverb:
		return "adverb", nil
	case SENGLexerTypePreposition:
		return "preposition", nil

	default:
		return "", fmt.Errorf("unrecognized token type %d", int(tokenType))
	}
}
