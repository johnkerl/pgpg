# PGPG Python apps. Run from repo root: make -C apps/py [target]
# Requires: make -C generated/py all (so lexers/parsers exist)

fmt:
	black .

clean:
	@:

.PHONY: test fmt clean
test:
	python3 trylex.py -e g:json '[1]'
	python3 tryparse.py -e g:json '[1]'
	python3 tryparse.py -e g:pemdas '1+2*3'
	python3 pemdas_eval.py -e '1+2*3'
	@echo "OK"
