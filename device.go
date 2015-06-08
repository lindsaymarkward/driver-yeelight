package main

// Yeelight device for Ninja Sphere driver

import (
	"fmt"
	"log"

	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

type YeelightDevice struct {
	*devices.LightDevice        // why isn't this as a pointer - don't we need to mutate it?
	IP                   string // driver config has this, but can't access it easily from light
	//	state             bool	// LightDevice has this
		sendEvent         func(event string, payload interface{}) error
	//	onOffChannel      *channels.OnOffChannel
	//	brightnessChannel *channels.BrightnessChannel
	//	colorChannel      *channels.ColorChannel
	//	transitionChannel *channels.TransitionChannel
	//	identifyChannel   *channels.IdentifyChannel
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

	// create a "proper" LightDevice using Ninja's built-in type
	lightDevice, err := devices.CreateLightDevice(d, infoModel, d.Conn)
	if err != nil {
		log.Printf("Error creating light device %v\n", id)
	}

	// Functions for connecting built-in events to Yeelight commands

	// ApplyLightState runs when airwheeling...
	lightDevice.ApplyLightState = func(state *devices.LightDeviceState) error {
		// for some reason, nothing prints in here...
		log.Printf("Batch: %v", state)
		if state.OnOff != nil {
			err = yeelight.SetOnOff(lightDevice.GetDeviceInfo().NaturalID, *state.OnOff, ip)
		}
		if state.Brightness != nil {
			// state.Brightness is a float value between 0-1
			err = yeelight.SetBrightness(lightDevice.GetDeviceInfo().NaturalID, *state.Brightness, ip)
		}
		if state.Color != nil {
			err = lightDevice.ApplyColor(state.Color)
		}
		return err
	}

	// ApplyColor takes a color state struct and sets the colour of the light
	// the color state struct contains the "Mode" of light that has been set,
	// which can either be "hue" or "temperature" for Yeelights
	lightDevice.ApplyColor = func(state *channels.ColorState) error {
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

	// enable channels that Yeelight supports
	if err := lightDevice.EnableOnOffChannel(); err != nil {
		log.Printf("Could not enable on-off channel. %v", err)
	}
	if err := lightDevice.EnableBrightnessChannel(); err != nil {
		log.Printf("Could not enable brightness channel. %v", err)
	}
	if err := lightDevice.EnableColorChannel("temperature", "hue"); err != nil {
		log.Printf("Could not enable color channel. %v", err)
	}

	return &YeelightDevice{LightDevice: lightDevice, IP: ip}
}


// doesn't seem to be necessary. In lifx, not samsung-tv
// (this doesn't SendState events)
//func (l *YeelightDevice) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
//	l.sendEvent = sendEvent
//}

//var reg, _ = regexp.Compile("[^a-z0-9]")
//
//// Exported by service/device schema
//func (l *YeelightDevice) SetName(name *string) (*string, error) {
//	log.Printf("\n\n\nSetting device name to %s\n\n\n", *name)
//
//	safe := reg.ReplaceAllString(strings.ToLower(*name), "")
//	if len(safe) > 5 {
//		safe = safe[0:5]
//	}
//
//	// Why is this here??
//	log.Printf("Pretending we can only set 5 lowercase alphanum. Name now: %s", safe)
//
//	l.sendEvent("renamed", safe)
//
//	return &safe, nil
//}
