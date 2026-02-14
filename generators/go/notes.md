make
./lexgen-tables -o 1.json 1.bnf
./lexgen-code   -o 1.go   1.json

change `package lexers` to `package main`

Add:
```
func main() {
	for _, arg := range os.Args[1:] {
		lexer := NewGeneratedLexer(arg)
		_ = lexer.Run()
	}
}
```

$ go run 1.go '1+2'
Line    1 column    1 type digit            token <<1>>
Line    1 column    2 type sign             token <<+>>
Line    1 column    3 type digit            token <<2>>
Line    1 column    4 type EOF              token <<>>
