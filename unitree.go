package main

import (
	"strings"

	"fortio.org/rand"
	"fortio.org/terminal/ansipixels"
)

const (
	BlockFull = "█"
	BlockBL   = "◢" // bottom left triangle
	BlockBR   = "◣" // bottom right triangle
	BlockTL   = "◥" // top left triangle
	BlockTR   = "◤" // top right triangle
)

// Example of what we want to generate procedurally.
const trunk = `
██   ◢██◤
◥█◣  ███
 ◥█◣◢██◤
 ◢████◤
 ◥████◣
 ◢█████◣
`

func (st *State) TreeBase(_ rand.Rand) {
	w := st.ap.W
	h := st.ap.H
	cx := (w - 1) / 2

	// Generate the trunk as a string
	// ... (or reuse ptree and convert to string like above)

	// Draw it line by line
	l := LineByLine{x: cx - 4, y: h - 9}
	l.DrawMulti(st.ap, trunk)
}

type LineByLine struct {
	x, y int
}

func (l *LineByLine) DrawMulti(ap *ansipixels.AnsiPixels, s string) {
	for line := range strings.SplitSeq(s, "\n") {
		if line == "" {
			continue
		}
		l.DrawLine(ap, line)
	}
}

func (l *LineByLine) DrawLine(ap *ansipixels.AnsiPixels, s string) {
	leadingSpaces := 0
	for _, c := range s {
		if c != ' ' {
			break
		}
		leadingSpaces++
	}
	s = s[leadingSpaces:]
	ap.WriteAtStr(l.x+leadingSpaces, l.y, s)
	l.y++
}
