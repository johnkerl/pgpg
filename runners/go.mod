module github.com/johnkerl/pgpg/runners

go 1.19

require (
	github.com/johnkerl/pgpg/generated v0.0.0
	github.com/johnkerl/pgpg/manual v0.0.0
)

replace github.com/johnkerl/pgpg/manual => ../manual

replace github.com/johnkerl/pgpg/generated => ../generated
