package lib

import (
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
	"github.com/lucasb-eyer/go-colorful"
)

// GenerateGamutMask generates a wheel (as *image.RGBA64) of Gamut Mask with a size of maskWidth, maskHeight
func GenerateGamutMask(img image.Image, maskWidth, maskHeight int) (wheel *image.RGBA64) {
	bounds := img.Bounds()

	wheel = image.NewRGBA64(image.Rect(0, 0, maskWidth, maskHeight))

	context := gg.NewContext(maskWidth, maskHeight)
	context.DrawEllipse(float64(maskWidth)/2, float64(maskHeight)/2, float64(maskWidth)/2, float64(maskHeight)/2)
	context.SetRGB(0, 0, 0)
	context.Fill()

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, a := context.Image().At(x, y).RGBA()
			wheel.SetRGBA64(int(x), int(y),
				color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(a)})
			//wheel.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}

	// Fill with black background
	//draw.Draw(wheel, bounds, &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)

	// if true {
	// 	return wheel
	// }

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			h, s, v := hsv(r, g, b)
			// Rotating by -math.Pi/2 so Red appears on top
			x := math.Cos(h*math.Pi/180-math.Pi/2)*s*float64(maskWidth)/2.0 + float64(maskWidth)/2.0
			y := math.Sin(h*math.Pi/180-math.Pi/2)*s*float64(maskHeight)/2.0 + float64(maskHeight)/2.0

			current := wheel.RGBA64At(int(x), int(y))
			_, _, currentV := hsv(uint32(current.R), uint32(current.G), uint32(current.B))
			if currentV < v {
				wheel.SetRGBA64(int(x), int(y),
					color.RGBA64{uint16(r), uint16(g), uint16(b), uint16(0xFFFF)})
			}
		}
	}
	return wheel
}

func hsv(r, g, b uint32) (h, s, v float64) {
	c := colorful.Color{
		R: float64(r) / float64(0xFFFF),
		G: float64(g) / float64(0xFFFF),
		B: float64(b) / float64(0xFFFF)}
	h, s, v = c.Hsv()
	return
}
