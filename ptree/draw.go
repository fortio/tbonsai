package ptree

import (
	"image"
	"image/color"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
)

func DrawTree(img *image.NRGBA, c *Canvas) {
	var notset tcolor.RGBColor
	for _, b := range c.Branches {
		rgb := c.MonoColor
		if rgb == notset {
			c := tcolor.Oklchf(.7, .7, c.Rand.Float64())
			ct, data := c.Decode()
			rgb = tcolor.ToRGB(ct, data)
		}
		drawBranch(img, b, rgb)
	}
}

func drawBranch(img *image.NRGBA, b *Branch, rgb tcolor.RGBColor) {
	ansipixels.DrawAALine(img, b.Start.X, b.Start.Y, b.End.X, b.End.Y, color.NRGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255})
}
