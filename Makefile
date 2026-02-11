build:
	make -C manual
	make -C generator
	make -C generated
	make -C runners

clean:
	make -C manual    clean
	make -C generator clean
	make -C generated clean
	make -C runners   clean

fmt:
	make -C manual    fmt
	make -C generator fmt
	make -C generated fmt
	make -C runners   fmt

.PHONY: build
