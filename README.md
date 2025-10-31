[![GoDoc](https://godoc.org/fortio.org/tbonsai?status.svg)](https://pkg.go.dev/fortio.org/tbonsai)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tbonsai)](https://goreportcard.com/report/fortio.org/tbonsai)
[![GitHub Release](https://img.shields.io/github/release/fortio/tbonsai.svg?style=flat)](https://github.com/fortio/tbonsai/releases/)
[![CI Checks](https://github.com/fortio/tbonsai/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tbonsai/actions/workflows/include.yml)
[![codecov](https://codecov.io/github/fortio/tbonsai/graph/badge.svg?token=Yx6QaeQr1b)](https://codecov.io/github/fortio/tbonsai)

# tbonsai

Ansipixels port of cbonsai

WIP - current state:

![Screenshot](example.png)

Which is progress from earlier:

```
            ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁
            ╲                                                     ╱
             ╲▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁╱
                  ⚪                                       ⚪
```

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

```
tbonsai help

flags:
```
