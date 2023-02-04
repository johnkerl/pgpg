package types

type AbstractLexer interface {
	Scan() (token *Token, err error)
}
