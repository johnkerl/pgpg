build:
	make -C manual
make -C generators/go _go
	make -C generated
	make -C runners

clean:
	make -C manual    clean
make -C generators/go _go clean
	make -C generated clean
	make -C runners   clean

fmt:
	make -C manual    fmt
make -C generators/go _go fmt
	make -C generated fmt
	make -C runners   fmt

.PHONY: build
