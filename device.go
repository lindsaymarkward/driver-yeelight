package main

// Yeelight device for Ninja Sphere driver

import (
	"fmt"
	"log"

	"github.com/lindsaymarkward/go-ninja/devices"
	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/model"
)

type YeelightDevice struct {
	*devices.LightDevice
	IP        string // driver config has this, but can't access it easily from light
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

	// use go-ninja's light type to create a LightDevice
	lightDevice, err := devices.CreateLightDevice(d, infoModel, d.Conn)
	if err != nil {
		log.Printf("Error creating light device %v\n", id)
	}

	// ?? test - set ThingID in Device
	// can get access to it but setting it doesn't do anything
	//	fmt.Printf("\nInfo: %v\n", lightDevice.GetDeviceInfo())
	//	fmt.Printf("\nLight name %v has ThingID %v\n", name, lightDevice.GetDeviceInfo().ThingID)
	//	lightDevice.GetDeviceInfo().ThingID = &name
	//	fmt.Printf("\n- Now name %v has ThingID %v\n", name, *lightDevice.GetDeviceInfo().ThingID)

	// ApplyLightState runs for a number of actions, including when airwheeling for brightness and color,
	// takes the state based on action and sends appropriate Yeelight command
	lightDevice.ApplyLightState = func(state *devices.LightDeviceState) error {
		// for some reason, nothing prints in here...
		log.Printf("Applying Light State: %v\n", *state)
		if state.OnOff != nil {
			err = yeelight.SetOnOff(lightDevice.GetDeviceInfo().NaturalID, *state.OnOff, ip)
			// send brightness to match on/off state
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
			// send on/off state to match brightness with minimum brightness before going off
			onOff := true
			if *state.Brightness < 0.08 {
				*state.Brightness = 0
				onOff = false
			}
			state.OnOff = &onOff

		}
		if state.Color != nil {
			r, g, b := yeelight.HSVToRGB(*state.Color.Hue, *state.Color.Saturation, 1)
			err = yeelight.SetColor(lightDevice.GetDeviceInfo().NaturalID, r, g, b, ip)
		}
		// update the state for the UI
		lightDevice.UpdateLightState(state)
		return err
	}

	// to determine if a light is on
	lightDevice.ApplyIsOn = func() (bool, error) {
		isOn, err := yeelight.IsOn(lightDevice.GetDeviceInfo().NaturalID, ip)
		return isOn, err
	}

	// enable channels that Yeelight supports
	if err := lightDevice.EnableOnOffChannel(); err != nil {
		log.Printf("Could not enable on-off channel. %v", err)
	}
	if err := lightDevice.EnableBrightnessChannel(); err != nil {
		log.Printf("Could not enable brightness channel. %v", err)
	}
	// don't enable temperature channel - then temperature -> hue is handled by LightDevice
	if err := lightDevice.EnableColorChannel("hue"); err != nil {
		log.Printf("Could not enable color channel. %v", err)
	}

	return &YeelightDevice{LightDevice: lightDevice, IP: ip}
}
