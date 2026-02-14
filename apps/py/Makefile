# PGPG Python apps. Run from repo root: make -C apps/py [target]
# Requires: make -C generated/py all (so lexers/parsers exist)

fmt:
	black .

clean:
	@:

.PHONY: test fmt clean
test:
	python3 trylex.py g:json expr '[1]'
	python3 tryparse.py g:json expr '[1]'
	python3 tryparse.py g:pemdas expr '1+2*3'
	python3 pemdas_eval.py expr '1+2*3'
	@echo "OK"
