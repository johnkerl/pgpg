```
make

./generator/parsegen-tables \
  -cpuprofile cpu.pprof \
  -memprofile mem.pprof \
  -trace trace.out \
  -o miller-temp.json \
  generated/bnfs/miller-temp.bnf

go tool pprof -http=:8082 cpu.pprof &
```
