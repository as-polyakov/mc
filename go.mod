module github.com/minio/mc

go 1.17

replace go.etcd.io/etcd => go.etcd.io/etcd/v3 v3.5.4

replace go.etcd.io/etcd/clientv3 => go.etcd.io/etcd/client/v3 v3.5.4

//replace github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt v3.2.2+incompatible
//replace github.com/minio/minio v0.0.0-20220817155032-67cf15d03665 => github.com/as-polyakov/minio v0.1.0-cvefix.2

//replace github.com/minio/minio/pkg/console => github.com/as-polyakov/minio/pkg/console v0.1.0-cvefix.2

//replace github.com/minio/minio/pkg/trie => github.com/as-polyakov/minio/pkg/trie v0.1.0-cvefix.2

//replace github.com/minio/minio/pkg/words => github.com/as-polyakov/minio/pkg/words v0.1.0-cvefix.2

//replace github.com/minio/minio/pkg/cert => github.com/as-polyakov/pkg/cert v0.1.0-cvefix.2

//exclude github.com/minio/minio v0.0.0-20220817155032-67cf15d03665

require (
	//github.com/minio/minio v0.0.0-20220817155032-67cf15d03665
	github.com/as-polyakov/minio v0.1.0-cvefix.3
	github.com/cheggaaa/pb v1.0.29
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.13.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/klauspost/compress v1.15.9
	github.com/mattn/go-ieproxy v0.0.1
	github.com/mattn/go-isatty v0.0.14
	github.com/minio/cli v1.22.0
	github.com/minio/minio-go/v7 v7.0.34
	github.com/minio/sha256-simd v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/profile v1.3.0
	github.com/pkg/xattr v0.4.5
	github.com/posener/complete v1.2.3
	github.com/rjeczalik/notify v0.9.2
	github.com/rs/xid v1.4.0
	github.com/shirou/gopsutil v3.20.11+incompatible
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa
	golang.org/x/net v0.0.0-20220812174116-3211cb980234
	golang.org/x/text v0.3.7
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
	gopkg.in/h2non/filetype.v1 v1.0.5
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/dswarbrick/smart v0.0.0-20190505152634-909a45200d6d // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.1.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/ncw/directio v1.0.5 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/secure-io/sio-go v0.3.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	go.etcd.io/etcd/api/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.4 // indirect
	go.etcd.io/etcd/client/v3 v3.5.4 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	google.golang.org/genproto v0.0.0-20220815135757-37a418bb8959 // indirect
	google.golang.org/grpc v1.48.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
)
