// tbonsai
// Ansipixels port of cbonsai

package main

import (
	"flag"
	"os"
	"runtime/pprof"
	"strings"

	"fortio.org/cli"
	"fortio.org/log"
	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func main() {
	os.Exit(Main())
}

type State struct {
	ap *ansipixels.AnsiPixels
}

func Main() int {
	truecolorDefault := ansipixels.DetectColorMode().TrueColor
	fTrueColor := flag.Bool("truecolor", truecolorDefault,
		"Use true color (24-bit RGB) instead of 8-bit ANSI colors (default is true if COLORTERM is set)")
	fCpuprofile := flag.String("profile-cpu", "", "write cpu profile to `file`")
	fMemprofile := flag.String("profile-mem", "", "write memory profile to `file`")
	fFPS := flag.Float64("fps", 60, "Frames per second (ansipixels rendering)")
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
	ap := ansipixels.NewAnsiPixels(*fFPS)
	st := &State{
		ap: ap,
	}
	ap.TrueColor = *fTrueColor
	if err := ap.Open(); err != nil {
		return 1 // error already logged
	}
	defer ap.Restore()
	ap.SyncBackgroundColor()
	ap.AutoSync = false // keeps cursor blinking.
	ap.OnResize = func() error {
		ap.ClearScreen()
		ap.StartSyncMode()
		// Redraw/resize/do something here:
		st.Pot()
		ap.WriteBoxed(ap.H/2-3, "Welcome to tbonsai!\n%dx%d\nQ to quit.", ap.W, ap.H)
		// ...
		ap.EndSyncMode()
		return nil
	}
	_ = ap.OnResize() // initial draw.
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
	if len(st.ap.Data) == 0 {
		return true
	}
	c := st.ap.Data[0]
	switch c {
	case 'q', 'Q', 3: // Ctrl-C
		log.Infof("Exiting on %q", c)
		return false
	default:
		// Do something
	}
	return true
}

func (st *State) Pot() {
	w := st.ap.W
	h := st.ap.H
	cx := (w - 1) / 2
	// pot base ~2/3rd of the width
	radius := w / 3
	// Feet
	st.ap.WriteAtStr(cx-radius-1, h-3, "╲")
	st.ap.WriteAtStr(cx+radius+1, h-3, "╱")
	gray := tcolor.DarkGray.Foreground()
	st.ap.WriteAtStr(cx-radius, h-2, "╲"+gray+strings.Repeat("▁", 2*radius-1)+tcolor.Reset+"╱")
	st.ap.WriteString(gray)
	st.ap.WriteAtStr(cx-radius+6, h-1, "●")
	st.ap.WriteAtStr(cx+radius-6, h-1, "●") // or ⚪ at -7
	st.ap.WriteAtStr(cx-radius-1, h-4, tcolor.Green.Foreground()+strings.Repeat("▁", 2*radius+3)+tcolor.Reset)
}
