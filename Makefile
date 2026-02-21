build:
	make -C manual
	make -C generators/go
	make -C apps/generated
	make -C apps/go

test:
	make -C manual        test
	make -C generators/go test
	make -C apps/go       test
	make -C generators/py test
	make -C apps/generated/py test
	make -C generators/js test
	make -C apps/generated/js test
	make -C apps/py       test

fmt:
	make -C manual        fmt
	make -C generators/go fmt
	make -C apps/generated fmt
	make -C apps/go       fmt
	make -C generators/py fmt
	make -C apps/generated/py fmt
	make -C apps/py       fmt

clean:
	make -C manual        clean
	make -C generators/go clean
	make -C apps/generated clean
	make -C apps/go       clean
	make -C generators/py clean
	make -C apps/generated/py clean
	make -C generators/js clean
	make -C apps/generated/js clean
	make -C apps/py       clean

.PHONY: build test fmt clean
