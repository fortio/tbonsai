// tbonsai
// Ansipixels port of cbonsai

package main

import (
	"flag"
	"image"
	"image/draw"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"fortio.org/cli"
	"fortio.org/duration"
	"fortio.org/log"
	"fortio.org/rand"
	"fortio.org/tbonsai/ptree"
	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func main() {
	os.Exit(Main())
}

type State struct {
	ap        *ansipixels.AnsiPixels
	pot       bool
	tree      bool
	auto      time.Duration
	last      time.Time
	monoColor tcolor.RGBColor
	rand      rand.Rand
	lines     bool
}

func Main() int {
	truecolorDefault := ansipixels.DetectColorMode().TrueColor
	fTrueColor := flag.Bool("truecolor", truecolorDefault,
		"Use true color (24-bit RGB) instead of 8-bit ANSI colors (default is true if COLORTERM is set)")
	fCpuprofile := flag.String("profile-cpu", "", "write cpu profile to `file`")
	fMemprofile := flag.String("profile-mem", "", "write memory profile to `file`")
	fPot := flag.Bool("pot", false, "Draw the pot")
	fFPS := flag.Float64("fps", 60, "Frames per second (ansipixels rendering)")
	fMonoColor := flag.String("color", "",
		"If set to a `hex color` like FD9103, use that single color for the tree instead of random colors")
	fAuto := duration.Flag("auto", 0, "If >0, automatically redraw a new tree at this `interval` and no user input is needed")
	fSeed := flag.Uint64("seed", 0, "Seed for random number generation. 0 means different random each run")
	fLines := flag.Bool("lines", false, "Use simple line drawing instead of polygon mode (default is polygon)")
	cli.Main()
	if *fCpuprofile != "" {
		f, err := os.Create(*fCpuprofile)
		if err != nil {
			return log.FErrf("can't open file for cpu profile: %v", err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			return log.FErrf("can't start cpu profile: %v", err)
		}
		log.Infof("Writing cpu profile to %s", *fCpuprofile)
		defer pprof.StopCPUProfile()
	}
	rnd := rand.New(*fSeed)
	ap := ansipixels.NewAnsiPixels(*fFPS)
	st := &State{
		ap:    ap,
		pot:   *fPot,
		auto:  *fAuto,
		rand:  rnd,
		lines: *fLines,
	}
	if *fMonoColor != "" {
		c, err := tcolor.FromString(*fMonoColor)
		if err != nil {
			return log.FErrf("can't parse mono-color %q: %v", *fMonoColor, err)
		}
		ct, data := c.Decode()
		st.monoColor = tcolor.ToRGB(ct, data)
	}
	ap.TrueColor = *fTrueColor
	if err := ap.Open(); err != nil {
		return 1 // error already logged
	}
	defer ap.Restore()
	if st.auto > 0 {
		st.tree = true
		ap.HideCursor()
	}
	ap.SyncBackgroundColor()
	ap.OnResize = func() error {
		ap.ClearScreen()
		ap.StartSyncMode()
		if st.tree {
			// In tree mode, redraw a new tree at the new size
			st.DrawTree()
		} else {
			// Initial screen being resized
			st.Pot()
			ap.WriteBoxed(ap.H/2-3, "Welcome to tbonsai!\n%dx%d\nQ to quit,\nT for a tree.", ap.W, ap.H)
		}
		ap.EndSyncMode()
		return nil
	}
	_ = ap.OnResize()   // initial draw.
	ap.AutoSync = false // keeps cursor blinking.
	err := ap.FPSTicks(st.Tick)
	if *fMemprofile != "" {
		f, errMP := os.Create(*fMemprofile)
		if errMP != nil {
			return log.FErrf("can't open file for mem profile: %v", errMP)
		}
		errMP = pprof.WriteHeapProfile(f)
		if errMP != nil {
			return log.FErrf("can't write mem profile: %v", err)
		}
		log.Infof("Wrote memory profile to %s", *fMemprofile)
		_ = f.Close()
	}
	if err != nil {
		log.Infof("Exiting on %v", err)
		return 1
	}
	return 0
}

func (st *State) Tick() bool {
	if st.auto > 0 && time.Since(st.last) >= st.auto {
		st.DrawTree()
	}
	if len(st.ap.Data) == 0 {
		return true
	}
	c := st.ap.Data[0]
	switch c {
	case 'q', 'Q', 3: // Ctrl-C
		log.Infof("Exiting on %q", c)
		return false
	case 't', 'T':
		if !st.tree {
			st.ap.HideCursor()
			st.tree = true
		}
		st.DrawTree()
	default:
		// Do something
	}
	return true
}

func (st *State) Pot() {
	if !st.pot {
		return
	}
	w := st.ap.W
	h := st.ap.H
	cx := (w - 1) / 2
	// pot base 1/4th of the width
	radius := w / 8
	// Feet
	st.ap.WriteAtStr(cx-radius-1, h-3, "╲")
	st.ap.WriteAtStr(cx+radius+1, h-3, "╱")
	gray := tcolor.DarkGray.Foreground()
	st.ap.WriteAtStr(cx-radius, h-2, "╲"+gray+strings.Repeat("▁", 2*radius-1)+tcolor.Reset+"╱")
	st.ap.WriteString(gray)
	st.ap.WriteAtStr(cx-radius+5, h-1, "●")
	st.ap.WriteAtStr(cx+radius-5, h-1, "●") // or ⚪ at -7
	st.ap.WriteAtStr(cx-radius-1, h-4, tcolor.Green.Foreground()+strings.Repeat("▁", 2*radius+3)+tcolor.Reset)
	if !st.tree {
		st.TreeBase(st.rand) // alternative tree base when not drawing branches as lines/polygons but unicode blocks instead.
	}
}

func (st *State) DrawTree() {
	dy := 6
	if !st.pot {
		dy = 0
	}
	c := ptree.NewCanvas(st.rand, st.ap.W, 2*st.ap.H-dy)
	c.MonoColor = st.monoColor
	img := image.NewNRGBA(image.Rect(0, 0, st.ap.W, 2*st.ap.H-dy))
	ptree.DrawTree(img, c, st.lines)
	nimg := image.NewRGBA(img.Bounds())
	draw.Draw(nimg, img.Bounds(), img, image.Point{}, draw.Src)
	st.ap.StartSyncMode()
	st.ap.ClearScreen()
	st.Pot()
	_ = st.ap.ShowScaledImage(nimg)
	st.ap.EndSyncMode()
	st.last = time.Now()
}
