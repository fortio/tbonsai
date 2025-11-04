// tbonsai
// Ansipixels port of cbonsai

package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
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
	ap             *ansipixels.AnsiPixels
	pot            bool
	tree           bool
	auto           time.Duration
	last           time.Time
	monoColor      tcolor.RGBColor
	rand           rand.Rand
	lines          bool
	depth          int
	trunkWidth     float64
	trunkHeightPct float64
	spread         float64
	kitty          bool
	width          int
	height         int
}

func SavePNG(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// KittyImage sends an image using the Kitty graphics protocol with auto-fit
// https://sw.kovidgoyal.net/kitty/graphics-protocol/
func KittyImage(w io.Writer, img image.Image, termWidth, termHeight int) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	data := buf.Bytes()
	chunkSize := 4096 // 4KB chunks
	// First, delete all previous images (a=d action=delete)
	fmt.Fprint(w, "\x1b_Ga=d;\x1b\\")
	// Setup
	// First chunk with terminal size for auto-fit (preserves aspect ratio)
	// c=columns, r=rows specify the display area
	// C=1: do not move cursor after displaying image
	fmt.Fprintf(w, "\x1b_Ga=T,f=100,q=1,C=1,z=-1,c=%d,r=%d", termWidth, termHeight)
	i := 0
	for len(data) > chunkSize {
		chunk := data[:chunkSize]
		data = data[chunkSize:]
		// q=1: suppress OK response but keep errors
		// a=T: transmit image data
		// f=100: PNG format
		// m=1: more chunks follow
		if i == 0 {
			// First chunk already has c= and r= for auto-fit
			fmt.Fprint(w, ",m=1;")
		} else {
			fmt.Fprint(w, "\x1b_Ga=T,f=100,q=1,m=1;")
		}
		fmt.Fprint(w, base64.StdEncoding.EncodeToString(chunk))
		fmt.Fprint(w, "\x1b\\")
		i++
	}
	// Last chunk (m=0 is default)
	if i == 0 {
		fmt.Fprint(w, ";")
	} else {
		fmt.Fprint(w, "\x1b_Ga=T,f=100,q=1;")
	}
	fmt.Fprint(w, base64.StdEncoding.EncodeToString(data))
	fmt.Fprint(w, "\x1b\\")
	return nil
}

func PNGMode(st *State, filename string, width, height int) int {
	// Save a single generated tree as a PNG image and exit
	c := ptree.NewCanvasWithOptions(st.rand, width, height, st.depth, st.trunkWidth, st.trunkHeightPct, st.spread)
	c.MonoColor = st.monoColor
	var img draw.Image
	if st.lines {
		img = image.NewNRGBA(image.Rect(0, 0, width, height))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, width, height))
	}
	ptree.DrawTree(img, c, st.lines)
	if err := SavePNG(filename, img); err != nil {
		return log.FErrf("failed to save PNG: %v", err)
	}
	return 0
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
	fSave := flag.String("save", "", "If set to a `file name`, saves one generated tree as a PNG image to that file and exits")
	fKitty := flag.Bool("kitty", false, "Use Kitty graphics protocol for high-res images (resizable, regeneratable)")
	fWidth := flag.Int("width", 1280, "Width of the generated tree image when using Kitty mode or saving to PNG")
	fHeight := flag.Int("height", 720, "Height of the generated tree image when using Kitty mode or saving to PNG")
	fDepth := flag.Int("depth", 4, "Tree depth (number of branch levels)")
	fTrunkWidth := flag.Float64("trunk-width", 8.0, "Starting width of the trunk")
	fTrunkHeight := flag.Float64("trunk-height", 40.0, "Trunk height as `percentage` of available height")
	fSpread := flag.Float64("spread", 1.0, "Branch angle spread multiplier (< 1.0 narrower, > 1.0 wider)")
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
		ap:             ap,
		pot:            *fPot,
		auto:           *fAuto,
		rand:           rnd,
		lines:          *fLines,
		depth:          *fDepth,
		trunkWidth:     *fTrunkWidth,
		trunkHeightPct: *fTrunkHeight,
		spread:         *fSpread,
		kitty:          *fKitty,
		width:          *fWidth,
		height:         *fHeight,
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
	if *fSave != "" {
		return PNGMode(st, *fSave, *fWidth, *fHeight)
	}
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
	var width, height int
	var dy int
	if st.pot {
		dy = 3
	}
	if st.kitty {
		// Use fixed dimensions for Kitty mode
		width = st.width
		height = st.height
	} else {
		// Use terminal dimensions for ansipixels mode
		width = st.ap.W
		height = 2 * (st.ap.H - dy)
	}

	c := ptree.NewCanvasWithOptions(st.rand, width, height, st.depth, st.trunkWidth, st.trunkHeightPct, st.spread)
	c.MonoColor = st.monoColor

	var img draw.Image
	if st.lines {
		img = image.NewNRGBA(image.Rect(0, 0, width, height))
	} else {
		img = image.NewRGBA(image.Rect(0, 0, width, height))
	}
	ptree.DrawTree(img, c, st.lines)

	st.ap.StartSyncMode()
	st.ap.ClearScreen()
	st.Pot()
	if st.kitty {
		st.ap.MoveCursor(0, 0)
		_ = KittyImage(st.ap.Out, img, st.ap.W, st.ap.H-dy)
	} else {
		// Convert NRGBA to RGBA if needed
		var showImg *image.RGBA
		if st.lines {
			showImg = image.NewRGBA(img.Bounds())
			draw.Draw(showImg, img.Bounds(), img, image.Point{}, draw.Src)
		} else {
			showImg = img.(*image.RGBA)
		}
		_ = st.ap.ShowScaledImage(showImg)
	}
	st.ap.EndSyncMode()
	st.last = time.Now()
}
