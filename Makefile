build:
	make -C lib/go
	make -C generators/go
	make -C apps/go/generated
	make -C apps/go

test:
	make -C lib/go            test
	make -C generators/go     test
	make -C apps/go           test
	make -C generators/py     test
	make -C apps/py/generated test
	make -C generators/js     test
	make -C apps/js/generated test
	make -C apps/py           test

fmt:
	make -C lib/go            fmt
	make -C generators/go     fmt
	make -C apps/go/generated fmt
	make -C apps/go           fmt
	make -C generators/py     fmt
	make -C apps/py/generated fmt
	make -C apps/py           fmt

clean:
	make -C lib/go            clean
	make -C generators/go     clean
	make -C apps/go/generated clean
	make -C apps/go           clean
	make -C generators/py     clean
	make -C apps/py/generated clean
	make -C generators/js     clean
	make -C apps/js/generated clean
	make -C apps/py           clean

.PHONY: build test fmt clean
