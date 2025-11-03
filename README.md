[![GoDoc](https://godoc.org/fortio.org/tbonsai?status.svg)](https://pkg.go.dev/fortio.org/tbonsai)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tbonsai)](https://goreportcard.com/report/fortio.org/tbonsai)
[![GitHub Release](https://img.shields.io/github/release/fortio/tbonsai.svg?style=flat)](https://github.com/fortio/tbonsai/releases/)
[![CI Checks](https://github.com/fortio/tbonsai/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tbonsai/actions/workflows/include.yml)
[![codecov](https://codecov.io/github/fortio/tbonsai/graph/badge.svg?token=Yx6QaeQr1b)](https://codecov.io/github/fortio/tbonsai)

# tbonsai

Ansipixels based reinterpretation of cbonsai (without looking at that program's source)

Current state (with optional `-pot` and `-lines` only and random colors):

![Terminal Screenshot](example.png)

You can also save high resolution images, for instance
```
tbonsai -color F3A005 -trunk-width 80 -depth 7 -seed 42 -trunk-height 35 -save out80_7_35_mono.png
```

Produces

![High Res Image](out80_7_35_mono.png)

## Install
You can get the binary from [releases](https://github.com/fortio/tbonsai/releases)

Or just run
```
CGO_ENABLED=0 go install fortio.org/tbonsai@latest  # to install (in ~/go/bin typically) or just
CGO_ENABLED=0 go run fortio.org/tbonsai@latest  # to run without install
```

or
```
brew install fortio/tap/tbonsai
```

or
```
docker run -ti fortio/tbonsai
```


## Usage

Use `-pot` to show the pot for instance,

Use `-auto 1s` for a new tree every 1s without needing to press "T"


See help for other flags/options

```
tbonsai help
```
