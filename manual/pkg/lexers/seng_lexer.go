package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

const (
	SENGLexerTypeNoun                       tokens.TokenType = "noun"
	SENGLexerTypeAdjective                  tokens.TokenType = "adjective"
	SENGLexerTypeArticle                    tokens.TokenType = "article"
	SENGLexerTypeTransitiveVerb             tokens.TokenType = "transitive verb"
	SENGLexerTypeIntransitiveVerb           tokens.TokenType = "intransitive verb"
	SENGLexerTypeTransitiveImperativeVerb   tokens.TokenType = "transitive imperative verb"
	SENGLexerTypeIntransitiveImperativeVerb tokens.TokenType = "intransitive imperative verb"
	SENGLexerTypeAdverb                     tokens.TokenType = "adverb"
	SENGLexerTypePreposition                tokens.TokenType = "preposition"
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
