```
make

justtime ./generators/go/parsegen-tables -o miller-temp.json generated/bnfs/miller-temp.bnf

justtime ./generators/go/parsegen-tables -nosort -o miller-temp.json generated/bnfs/miller-temp.bnf

./generators/go/parsegen-tables \
  -cpuprofile cpu.pprof \
  -memprofile mem.pprof \
  -trace trace.out \
  -o miller-temp.json \
  generated/bnfs/miller-temp.bnf

./generators/go/parsegen-tables \
  -nosort \
  -cpuprofile cpu.pprof \
  -memprofile mem.pprof \
  -trace trace.out \
  -o miller-temp.json \
  generated/bnfs/miller-temp.bnf

go tool pprof -http=:8082 cpu.pprof
```
