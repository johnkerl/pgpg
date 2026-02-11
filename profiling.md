```
make

justtime ./generator/parsegen-tables -o miller-temp.json generated/bnfs/miller-temp.bnf

justtime ./generator/parsegen-tables -nosort -o miller-temp.json generated/bnfs/miller-temp.bnf

./generator/parsegen-tables \
  -cpuprofile cpu.pprof \
  -memprofile mem.pprof \
  -trace trace.out \
  -o miller-temp.json \
  generated/bnfs/miller-temp.bnf

./generator/parsegen-tables \
  -nosort \
  -cpuprofile cpu.pprof \
  -memprofile mem.pprof \
  -trace trace.out \
  -o miller-temp.json \
  generated/bnfs/miller-temp.bnf

go tool pprof -http=:8082 cpu.pprof
```
