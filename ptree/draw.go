package ptree

import (
	"image"
	"image/color"
	"image/draw"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
	"golang.org/x/image/vector"
)

func DrawTree(img draw.Image, c *Canvas, useLines bool) {
	var notset tcolor.RGBColor
	for _, b := range c.Branches {
		rgb := c.MonoColor
		if rgb == notset {
			c := tcolor.Oklchf(.7, .7, c.Rand.Float64())
			ct, data := c.Decode()
			rgb = tcolor.ToRGB(ct, data)
		}
		if useLines {
			drawBranchLine(img.(*image.NRGBA), b, rgb)
		} else {
			drawBranchPolygon(img.(*image.RGBA), b, rgb)
		}
	}
}

func drawBranchLine(img *image.NRGBA, b *Branch, rgb tcolor.RGBColor) {
	ansipixels.DrawAALine(img, b.Start.X, b.Start.Y, b.End.X, b.End.Y, toNRGBA(rgb))
}

func toNRGBA(rgb tcolor.RGBColor) color.NRGBA {
	return color.NRGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

func toRGBA(rgb tcolor.RGBColor) color.RGBA {
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

func drawBranchPolygon(img *image.RGBA, b *Branch, rgb tcolor.RGBColor) {
	perpX, perpY := b.Perpendicular()
	if perpX == 0 && perpY == 0 {
		return
	}

	// Calculate the 4 vertices of the trapezoid
	// Start point offsets (full width / 2)
	startHalfWidth := b.StartWidth / 2
	// For trunk, bottom should be flat (horizontal)
	var s1x, s1y, s2x, s2y float64
	if b.IsTrunk {
		// Flat bottom: perpendicular to vertical (horizontal)
		s1x = b.Start.X + startHalfWidth
		s1y = b.Start.Y
		s2x = b.Start.X - startHalfWidth
		s2y = b.Start.Y
	} else {
		s1x = b.Start.X + perpX*startHalfWidth
		s1y = b.Start.Y + perpY*startHalfWidth
		s2x = b.Start.X - perpX*startHalfWidth
		s2y = b.Start.Y - perpY*startHalfWidth
	}

	// End point offsets (full width / 2)
	endHalfWidth := b.EndWidth / 2
	e1x := b.End.X + perpX*endHalfWidth
	e1y := b.End.Y + perpY*endHalfWidth
	e2x := b.End.X - perpX*endHalfWidth
	e2y := b.End.Y - perpY*endHalfWidth

	// Create rasterizer and draw the trapezoid
	rast := vector.NewRasterizer(img.Bounds().Dx(), img.Bounds().Dy())
	rast.DrawOp = draw.Over

	// Move to first vertex and draw the polygon
	rast.MoveTo(float32(s1x), float32(s1y))
	rast.LineTo(float32(e1x), float32(e1y))
	rast.LineTo(float32(e2x), float32(e2y))
	rast.LineTo(float32(s2x), float32(s2y))
	rast.ClosePath()

	// Rasterize to the image
	rast.Draw(img, img.Bounds(), image.NewUniform(toRGBA(rgb)), image.Point{})
}
