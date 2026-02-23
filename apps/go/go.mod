module github.com/johnkerl/pgpg/apps/go

go 1.25

require (
	github.com/johnkerl/pgpg/generated v0.0.0
	github.com/johnkerl/pgpg/lib v0.0.0
)

replace github.com/johnkerl/pgpg/lib => ../../lib/go

replace github.com/johnkerl/pgpg/generated => ./generated
