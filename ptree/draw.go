package ptree

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"fortio.org/terminal/ansipixels"
	"fortio.org/terminal/ansipixels/tcolor"
	"golang.org/x/image/vector"
)

func DrawTree(img draw.Image, c *Canvas, useLines bool) {
	var notset tcolor.RGBColor
	var rast *vector.Rasterizer
	if !useLines {
		// Reuse single rasterizer for all branches
		rast = vector.NewRasterizer(img.Bounds().Dx(), img.Bounds().Dy())
		rast.DrawOp = draw.Over
	}
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
			drawBranchPolygon(img.(*image.RGBA), b, rgb, rast)
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

func drawBranchPolygon(img *image.RGBA, b *Branch, rgb tcolor.RGBColor, rast *vector.Rasterizer) {
	perpX, perpY := b.Perpendicular()
	if perpX == 0 && perpY == 0 {
		return
	}

	startHalfWidth := b.StartWidth / 2
	endHalfWidth := b.EndWidth / 2

	// Calculate the 4 vertices of the trapezoid
	var s1x, s1y, s2x, s2y float64
	if b.IsTrunk {
		// Flat bottom for trunk
		s1x, s1y = b.Start.X+startHalfWidth, b.Start.Y
		s2x, s2y = b.Start.X-startHalfWidth, b.Start.Y
	} else {
		s1x, s1y = b.Start.X+perpX*startHalfWidth, b.Start.Y+perpY*startHalfWidth
		s2x, s2y = b.Start.X-perpX*startHalfWidth, b.Start.Y-perpY*startHalfWidth
	}

	e1x, e1y := b.End.X+perpX*endHalfWidth, b.End.Y+perpY*endHalfWidth
	e2x, e2y := b.End.X-perpX*endHalfWidth, b.End.Y-perpY*endHalfWidth

	// Calculate tight bounding box
	minX := min(s1x, s2x, e1x, e2x)
	maxX := max(s1x, s2x, e1x, e2x)
	minY := min(s1y, s2y, e1y, e2y)
	maxY := max(s1y, s2y, e1y, e2y)

	// Add 1px margin for anti-aliasing and clamp to image bounds
	imgBounds := img.Bounds()
	x0 := max(0, minX-1)
	y0 := max(0, minY-1)
	x1 := min(float64(imgBounds.Dx()), maxX+1)
	y1 := min(float64(imgBounds.Dy()), maxY+1)

	if x0 >= x1 || y0 >= y1 {
		return // Completely offscreen
	}
	x0Int := int(math.Floor(x0))
	y0Int := int(math.Floor(y0))
	x1Int := int(math.Ceil(x1))
	y1Int := int(math.Ceil(y1))
	wInt := x1Int - x0Int
	hInt := y1Int - y0Int

	// Reset rasterizer to tight bounding box
	rast.Reset(wInt, hInt)
	// Translate coordinates to local space
	dx, dy := float32(x0Int), float32(y0Int)
	rast.MoveTo(float32(s1x)-dx, float32(s1y)-dy)
	rast.LineTo(float32(e1x)-dx, float32(e1y)-dy)
	rast.LineTo(float32(e2x)-dx, float32(e2y)-dy)
	rast.LineTo(float32(s2x)-dx, float32(s2y)-dy)
	rast.ClosePath()

	// Rasterize to the bounding box region
	subImg := img.SubImage(image.Rect(x0Int, y0Int, x1Int, y1Int)).(*image.RGBA)
	rast.Draw(subImg, subImg.Bounds(), image.NewUniform(toRGBA(rgb)), image.Point{})
}
