module github.com/sagan/ptool

go 1.24.1

// workaround for some problem
replace github.com/hekmon/transmissionrpc/v2 => ./transmissionrpc

// workaround for some problem
replace github.com/stromland/cobra-prompt => ./cobra-prompt

require (
	github.com/KarpelesLab/reflink v1.0.2
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/Noooste/azuretls-client v1.6.4
	github.com/PuerkitoBio/goquery v1.10.2
	github.com/anacrolix/torrent v1.58.1
	github.com/ettle/strcase v0.2.0
	github.com/glebarez/sqlite v1.11.0
	github.com/gofrs/flock v0.12.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hekmon/transmissionrpc/v2 v2.0.1
	github.com/influxdata/go-prompt v0.2.8
	github.com/jpillora/go-tld v1.2.1
	github.com/mattn/go-runewidth v0.0.16
	github.com/natefinch/atomic v1.0.1
	github.com/noirbizarre/gonja v0.0.0-20200629003239-4d051fd0be61
	github.com/pelletier/go-toml/v2 v2.2.3
	github.com/pkg/errors v0.9.1
	github.com/shibumi/go-pathspec v1.3.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.19.0
	github.com/stromland/cobra-prompt v0.5.0
	github.com/zishang520/socket.io/clients/engine/v3 v3.0.0-rc.12
	github.com/zishang520/socket.io/clients/socket/v3 v3.0.0-rc.12
	github.com/zishang520/socket.io/v3 v3.0.0-rc.12
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa
	golang.org/x/net v0.50.0
	gorm.io/gorm v1.25.12
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/anacrolix/generics v0.0.3-0.20240902042256-7fb2702ef0ca // indirect
	github.com/bradfitz/iter v0.0.0-20191230175014-e8f45d346db8 // indirect
	github.com/dunglas/httpsfv v1.1.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gookit/color v1.6.0 // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.59.0 // indirect
	github.com/quic-go/webtransport-go v0.10.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/zishang520/socket.io/parsers/engine/v3 v3.0.0-rc.12 // indirect
	github.com/zishang520/socket.io/parsers/socket/v3 v3.0.0-rc.12 // indirect
	github.com/zishang520/socket.io/servers/engine/v3 v3.0.0-rc.12 // indirect
	github.com/zishang520/socket.io/servers/socket/v3 v3.0.0-rc.12 // indirect
	lukechampine.com/blake3 v1.4.0 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
	modernc.org/sqlite v1.35.0 // indirect
	resty.dev/v3 v3.0.0-beta.6 // indirect
)

require (
	github.com/Noooste/fhttp v1.0.12 // indirect
	github.com/Noooste/utls v1.3.6 // indirect
	github.com/Noooste/websocket v1.0.3 // indirect
	github.com/anacrolix/missinggo v1.3.0 // indirect
	github.com/anacrolix/missinggo/v2 v2.8.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/cloudflare/circl v1.6.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/glebarez/go-sqlite v1.22.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hekmon/cunits/v2 v2.1.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-tty v0.0.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sys v0.41.0
	golang.org/x/term v0.40.0
	golang.org/x/text v0.34.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)
