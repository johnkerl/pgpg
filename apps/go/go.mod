module github.com/johnkerl/pgpg/apps/go

go 1.25

require (
	github.com/johnkerl/pgpg/generated v0.0.0
	github.com/johnkerl/pgpg/manual v0.0.0
)

require github.com/johnkerl/goffl v0.1.0 // indirect

replace github.com/johnkerl/pgpg/manual => ../../manual

replace github.com/johnkerl/pgpg/generated => ../../generated
