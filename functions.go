package main

import "math"

// this file provides utility functions for the driver
// colour conversions and the like

// TemperatureToRGB converts a colour temperature in kelvins in range [1000, 40000] to RGB
// https://gist.github.com/paulkaplan/5184275
// From http://www.tannerhelland.com/4435/convert-temperature-rgb-algorithm-code/
func TemperatureToRGB(kelvin float64) (r, g, b uint8) {
	temp := kelvin / 100

	var red, green, blue float64

	if (temp <= 66) {

		red = 255

		green = temp
		green = 99.4708025861 * math.Log(green) - 161.1195681661

		if (temp <= 19) {
			blue = 0
		} else {
			blue = temp-10
			blue = 138.5177312231 * math.Log(blue) - 305.0447927307
		}

	} else {
		red = temp - 60;
		red = 329.698727446 * math.Pow(red, -0.1332047592)

		green = temp - 60
		green = 288.1221695283 * math.Pow(green, -0.0755148492)

		blue = 255
	}
	return clamp(red, 0, 255), clamp(green, 0, 255), clamp(blue, 0, 255)
}


func clamp(x, min, max float64) uint8 {

	if (x < min) {
		return uint8(min)
	}
	if (x > max) {
		return uint8(max)
	}
	return uint8(x)
}


// HSVToRGB converts an HSV triple to a RGB triple.
// from https://godoc.org/code.google.com/p/sadbox/color
// Ported from http://goo.gl/Vg1h9
func HSVToRGB(h, s, v float64) (r, g, b uint8) {
	var fR, fG, fB float64
	i := math.Floor(h * 6)
	f := h*6 - i
	p := v * (1.0 - s)
	q := v * (1.0 - f*s)
	t := v * (1.0 - (1.0-f)*s)
	switch int(i) % 6 {
	case 0:
		fR, fG, fB = v, t, p
	case 1:
		fR, fG, fB = q, v, p
	case 2:
		fR, fG, fB = p, v, t
	case 3:
		fR, fG, fB = p, q, v
	case 4:
		fR, fG, fB = t, p, v
	case 5:
		fR, fG, fB = v, p, q
	}
	r, g, b = float64ToUint8(fR), float64ToUint8(fG), float64ToUint8(fB)
	return
}

// float64ToUint8 converts a float64 to uint8.
// See: http://code.google.com/p/go/issues/detail?id=3423
func float64ToUint8(x float64) uint8 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 255
	}
	return uint8(int(x*255 + 0.5))
}
