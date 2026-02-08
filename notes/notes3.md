go build ./cmd/pgpg

./pgpg ir-to-go examples/minimal.json --out /tmp/miniparser

Check `/tmp/miniparser/parser.go` for `action_make_s` and embedded tables
