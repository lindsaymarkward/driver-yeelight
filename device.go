package main

// Yeelight device for Ninja Sphere driver

import (
	"fmt"
	"log"

	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/channels"
)

type YeelightDevice struct {
	*devices.LightDevice
	IP string // driver config has this, but can't access it easily from light
	sendEvent func(event string, payload interface{}) error
}

// NewYeelightDevice creates a light device, given a driver and an id (hex code used by Yeelight hub)
func NewYeelightDevice(d *YeelightDriver, id string, ip string) *YeelightDevice {

	name := d.config.Names[id]
	infoModel := &model.Device{
		NaturalID:     fmt.Sprintf("%s", id),
		NaturalIDType: "light",
		Name:          &name,
		Signatures: &map[string]string{
			"ninja:manufacturer": "Qingdao Yeelink",
			"ninja:productName":  "Yeelight",
			"ninja:productType":  "Light",
			"ninja:thingType":    "light",
		},
	}

	// create a "proper" LightDevice using go-ninja's "light" device type
	lightDevice, err := devices.CreateLightDevice(d, infoModel, d.Conn)
	if err != nil {
		log.Printf("Error creating light device %v\n", id)
	}

	// ?? test - set ThingID in Device
	//	fmt.Printf("\nInfo: %v\n", lightDevice.GetDeviceInfo())
	//	fmt.Printf("\nLight name %v has ThingID %v\n", name, lightDevice.GetDeviceInfo().ThingID)
	//	lightDevice.GetDeviceInfo().ThingID = &name
	//	fmt.Printf("\n- Now name %v has ThingID %v\n", name, *lightDevice.GetDeviceInfo().ThingID)

	// Functions for connecting built-in events to Yeelight commands

	// ApplyLightState runs for a number of actions, including when airwheeling for brightness and color
	lightDevice.ApplyLightState = func(state *devices.LightDeviceState) error {
		// for some reason, nothing prints in here...
		log.Printf("Applying Light State: %v\n", *state)
		if state.OnOff != nil {
			err = yeelight.SetOnOff(lightDevice.GetDeviceInfo().NaturalID, *state.OnOff, ip)
//			lightDevice.UpdateOnOffState(*state.OnOff)
//			if *state.OnOff {
////				lightDevice.UpdateBrightnessState(1)
//				state.Brightness = &float64(1.0)
//			} else {
//				state.Brightness = &float64(0.0)
//				lightDevice.UpdateBrightnessState(0)
//			}
			var brightness float64
			if *state.OnOff {
				brightness = 1.0
			} else {
				brightness = 0.0
			}
			state.Brightness = &brightness
		}
		if state.Brightness != nil {
			// state.Brightness is a float value between 0-1
			err = yeelight.SetBrightness(lightDevice.GetDeviceInfo().NaturalID, *state.Brightness, ip)
//			lightDevice.UpdateBrightnessState(*state.Brightness)
//			if *state.Brightness > 0 {
////				lightDevice.UpdateOnOffState(true)
//				state.OnOff = true
//			} else {
////				lightDevice.UpdateOnOffState(false)
//				state.OnOff = false
//			}
			onOff := true
			if *state.Brightness == 0 {
				onOff = false
			}
			state.OnOff = &onOff

		}
		if state.Color != nil {
			err = lightDevice.ApplyColor(state.Color)
//			lightDevice.UpdateColorState(*state.Color)
		}
		lightDevice.SetLightState(state)
		return err
	}

	// ApplyColor takes a color state struct and sets the colour of the light
	// the color state struct contains the "Mode" of light that has been set,
	// which can either be "hue" or "temperature" for Yeelights
	lightDevice.ApplyColor = func(state *channels.ColorState) error {
		log.Printf("Applying Color: %v\n", *state)
		var r, g, b uint8
		if state.Mode == "temperature" {
			// temperature is in the range [2000, 6500]
			r, g, b = yeelight.TemperatureToRGB(*state.Temperature)
		} else {
			// state must be "hue"
			r, g, b = yeelight.HSVToRGB(*state.Hue, *state.Saturation, 1)
		}
		log.Printf("Setting colour to RGB = %v, %v, %v\n", r, g, b)

		return yeelight.SetColor(lightDevice.GetDeviceInfo().NaturalID, r, g, b, ip)
	}

	// ApplyOnOff takes a bool and sets the light to on or off,
	// updating brightness state as well (0 or 1, which is maximum)
	lightDevice.ApplyOnOff = func(state bool) error {
		log.Printf("Applying onOff %v\n", state)
		err = yeelight.SetOnOff(lightDevice.GetDeviceInfo().NaturalID, state, ip)

		return err
	}

	// enable channels that Yeelight supports
	if err := lightDevice.EnableOnOffChannel(); err != nil {
		log.Printf("Could not enable on-off channel. %v", err)
	}
	if err := lightDevice.EnableBrightnessChannel(); err != nil {
		log.Printf("Could not enable brightness channel. %v", err)
	}
	if err := lightDevice.EnableColorChannel("hue"); err != nil {
		log.Printf("Could not enable color channel. %v", err)
	}

	return &YeelightDevice{LightDevice: lightDevice, IP: ip}
}
