module go-frame-p

go 1.17

require (
	ccache v0.0.0
	github.com/mattn/go-sqlite3 v1.14.14
	sorm v0.0.0-00010101000000-000000000000
	surpc v0.0.0
)

require google.golang.org/protobuf v1.28.0 // indirect

replace ccache => ./CCache

replace sorm => ./SORM

replace surpc => ./SuRPC
