package ptree

import (
	"image"
	"image/color"
	"math/rand/v2"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func DrawTree(img *image.NRGBA, c *Canvas) {
	for _, b := range c.Branches {
		drawBranch(img, b)
	}
}

func drawBranch(img *image.NRGBA, b *Branch) {
	c := tcolor.Oklchf(.7, .7, rand.Float64()) //nolint:gosec // not crypto.
	ct, data := c.Decode()
	rgbg := tcolor.ToRGB(ct, data)
	ansipixels.DrawAALine(img, b.Start.X, b.Start.Y, b.End.X, b.End.Y, color.NRGBA{R: rgbg.R, G: rgbg.G, B: rgbg.B, A: 255})
}
