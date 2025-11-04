[![GoDoc](https://godoc.org/fortio.org/tbonsai?status.svg)](https://pkg.go.dev/fortio.org/tbonsai)
[![Go Report Card](https://goreportcard.com/badge/fortio.org/tbonsai)](https://goreportcard.com/report/fortio.org/tbonsai)
[![GitHub Release](https://img.shields.io/github/release/fortio/tbonsai.svg?style=flat)](https://github.com/fortio/tbonsai/releases/)
[![CI Checks](https://github.com/fortio/tbonsai/actions/workflows/include.yml/badge.svg)](https://github.com/fortio/tbonsai/actions/workflows/include.yml)
[![codecov](https://codecov.io/github/fortio/tbonsai/graph/badge.svg?token=Yx6QaeQr1b)](https://codecov.io/github/fortio/tbonsai)

# tbonsai

Ansipixels based reinterpretation of cbonsai (without looking at that program's source)

Using `-leaves` and regular terminal half block resolution

![Terminal Leaves](exampleLeaves.png)

Same with `-kitty` using the Kitty Image protocol for high resolution image directly in the terminal background:

![Terminal Leaves Kitty](examplesLeavesKitty.png)

With optional `-pot` and `-lines` only and random colors:

![Terminal Screenshot](exampleLines.png)

You can also save high resolution images, for instance
```
tbonsai -trunk-width 6.25 -depth 7 -seed 42 -trunk-height 35 -save out80_7_35_mono.png
```

Produces

![High Res Image](out80_7_35_mono.png)

Use `-kitty` to see the high resolution image in your terminal that supports the Kitty Image protocol (kitty, ghostty, etc...)

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

Use `-leaves` to show leaves (winter tree otherwise)

Use `-pot` to show the pot for instance,

Use `-auto 1s` for a new tree every 1s without needing to press "T"

Use `-rainbow` for rainbow colors.

Etc,... See help for other flags/options

```
tbonsai help

tbonsai 1.0.0 usage:
        tbonsai [flags]
or 1 of the special arguments
        tbonsai {help|envhelp|version|buildinfo}
flags:
  -auto interval
        If >0, automatically redraw a new tree at this interval and no user input is needed
  -color hex color
        Trunk base color as hex color (default with leaves: #654321 dark brown, branches gradually lighten with depth).
  -depth int
        Tree depth (number of branch levels) (default 6)
  -fps float
        Frames per second (ansipixels rendering) (default 60)
  -height int
        Height of the generated tree image when using Kitty mode or saving to PNG (default 720)
  -kitty
        Use Kitty graphics protocol for high-res images (resizable, regeneratable)
  -leaf-size float
        Leaf size multiplier (default 1)
  -leaves
        Draw leaves at branch endpoints
  -lines
        Use simple line drawing instead of polygon mode (default is polygon)
  -pot
        Draw the pot
  -rainbow
        Use random colors for each branch instead of depth-based brown gradient
  -save file name
        If set to a file name, saves one generated tree as a PNG image to that file and exits
  -seed uint
        Seed for random number generation. 0 means different random each run
  -spread float
        Branch angle spread multiplier (< 1.0 narrower, > 1.0 wider) (default 1)
  -truecolor
        Use true color (24-bit RGB) instead of 8-bit ANSI colors (default is true if COLORTERM is set)
  -trunk-height percentage
        Trunk height as percentage of available height (default 35)
  -trunk-width percentage
        Starting width of the trunk as percentage of image width (default 7)
  -width int
        Width of the generated tree image when using Kitty mode or saving to PNG (default 1280)
```
