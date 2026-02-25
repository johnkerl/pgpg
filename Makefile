build:
	make -C go
	make -C apps/go/generated
	make -C apps/go

test:
	make -C go                test
	make -C apps/go           test

fmt:
	make -C go                fmt
	make -C apps/go/generated fmt
	make -C apps/go           fmt

clean:
	make -C go                clean
	make -C apps/go/generated clean
	make -C apps/go           clean

.PHONY: build test fmt clean
