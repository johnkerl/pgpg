build:
	make -C manual
make -C generators/go _go
	make -C generated
	make -C apps/go

clean:
	make -C manual    clean
make -C generators/go _go clean
	make -C generated clean
	make -C apps/go   clean

fmt:
	make -C manual    fmt
make -C generators/go _go fmt
	make -C generated fmt
	make -C apps/go   fmt

.PHONY: build
