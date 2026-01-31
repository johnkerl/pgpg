go build ./cmd/pgpg

./pgpg ir-to-go examples/minimal.json --out /tmp/miniparser

Check `/tmp/miniparser/parser.go` for `action_make_s` and embedded tables

If you want, I can next:
* Add a small runner example that feeds tokens to the generated parser.
* Implement a minimal LR(0) table generator to populate parser.actions/gotos automatically.

