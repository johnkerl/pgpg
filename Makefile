build:
	make -C manual
	make -C generators/go
	make -C generated
	make -C apps/go

clean:
	make -C manual        clean
	make -C generators/go clean
	make -C generated     clean
	make -C apps/go       clean
	make -C generators/py clean
	make -C generated/py  clean
	make -C generators/js clean
	make -C generated/js  clean
	make -C apps/py       clean

fmt:
	make -C manual        fmt
	make -C generators/go fmt
	make -C generated     fmt
	make -C apps/go       fmt
	make -C generators/py fmt
	make -C generated/py  fmt
	make -C apps/py       fmt

test:
	make -C manual        test
	make -C generators/go test
	make -C apps/go       test
	make -C generators/py test
	make -C generated/py  test
	make -C generators/js test
	make -C generated/js  test
	make -C apps/py       test

.PHONY: build clean fmt test
