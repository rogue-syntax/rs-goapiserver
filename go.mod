module github.com/rogue-syntax/rs-goapiserver

go 1.24.0

toolchain go1.24.4

replace github.com/rogue-syntax/rs_zerolog v1.0.0 => /home/fremont0/rs_zerolog/rs_zerolog

require (
	github.com/go-sql-driver/mysql v1.7.1
	golang.org/x/crypto v0.46.0
)

require github.com/google/uuid v1.6.0

require (
	github.com/gorilla/websocket v1.5.1
	github.com/rogue-syntax/goqb-rs v0.0.0-20230223155230-5901ce3f23e6
	github.com/rogue-syntax/rs_zerolog v1.0.0

)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-chi/chi/v5 v5.0.8 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/lib/pq v1.10.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-sqlite3 v1.14.8 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

require (
	github.com/gabriel-vasile/mimetype v1.4.3
	github.com/jmoiron/sqlx v1.3.5
	github.com/mailgun/mailgun-go/v4 v4.12.0
	github.com/minio/minio-go/v7 v7.0.66
	github.com/nbutton23/zxcvbn-go v0.0.0-20210217022336-fa2cb2858354
	github.com/pkg/errors v0.9.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
