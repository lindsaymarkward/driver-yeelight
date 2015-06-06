package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

type Yeelight struct {
	devices.LightDevice
	ip                string
	state             bool
	driver            ninja.Driver
	info              *model.Device
	thing             *model.Thing
	device            *devices.LightDevice
	sendEvent         func(event string, payload interface{}) error
	onOffChannel      *channels.OnOffChannel
	brightnessChannel *channels.BrightnessChannel
	colorChannel      *channels.ColorChannel
	transitionChannel *channels.TransitionChannel
	identifyChannel   *channels.IdentifyChannel
}

func NewYeelight(driver *YeelightDriver, id string) *Yeelight {
	// only give programmed name if it doesn't already have one
	var name string
	if driver.config.Names[id] == "" {
		name = fmt.Sprintf("Lt %v", id)
		driver.config.Names[id] = name
	} else {
		name = driver.config.Names[id]
	}

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

	lightDevice, err := devices.CreateLightDevice(driver, infoModel, driver.Conn)
	if err != nil {
		log.Printf("Error creating light device")
	}

	light := &Yeelight{
		ip:     driver.config.Hub.IP, // "192.168.1.59",
		driver: driver,
		device: lightDevice,
		info:   infoModel,
		thing: &model.Thing{
			Name: name,
		},
	}

	// what channels
	light.onOffChannel = channels.NewOnOffChannel(light)
	light.brightnessChannel = channels.NewBrightnessChannel(light)
	light.colorChannel = channels.NewColorChannel(light)
	// these 2 are only here to test/fix the crash on setBatch
	light.transitionChannel = channels.NewTransitionChannel(light)
	light.identifyChannel = channels.NewIdentifyChannel(light)

	// ApplyLightState runs when airwheeling...
	light.device.ApplyLightState = func(state *devices.LightDeviceState) error {
		// for some reason, nothing prints in here...
		//		fmt.Printf("\n\nWOW!! %v\n\n", state)
//		log.Printf("Batch: %v", state)
		if state.OnOff != nil {
			err := light.SetOnOff(*state.OnOff)
			if err != nil {
				return err
			}
		}
		if state.Brightness != nil {
			light.SetBrightness(*state.Brightness)
//			log.Printf("\nBrightness: %v\n\n", *state.Brightness)
		}
		// TODO: Colour state changing
		if state.Color != nil {
			light.SetColor(state.Color)
		}

		return nil
	}

	return light
}

func (l *Yeelight) ApplyLightState(state *devices.LightDeviceState) error {
	fmt.Printf("\n\nnow it's WOW!! %v\n\n", state)
	return nil
}

func (l *Yeelight) GetDeviceInfo() *model.Device {
	return l.info
}

func (l *Yeelight) GetDriver() ninja.Driver {
	return l.driver
}

// these functions are where the action happens - send commands to the Yeelight bulbs
// TODO: update app/model status when these change... I think??

func (l *Yeelight) SetTransition(state int) error {
	fmt.Printf("Really?\n")
	return nil
}

func (l *Yeelight) SetOnOff(state bool) error {
	log.Printf("Turning %t", state)
	// turn light on/off (yeelight.SetOnOff handles state choice)
	yeelight.SetOnOff(l.info.NaturalID, state, l.ip)
	l.state = state
	l.onOffChannel.SendState(state)
	if l.state {
		l.brightnessChannel.SendState(1)
	} else {
		l.brightnessChannel.SendState(0)
	}
	return nil
}

func (l *Yeelight) ToggleOnOff() error {
	log.Println("\n\nToggling!\n")
	yeelight.ToggleOnOff(l.info.NaturalID, l.ip)
	l.state = !l.state
	l.onOffChannel.SendState(l.state)
	if l.state {
		l.brightnessChannel.SendState(1)
	} else {
		l.brightnessChannel.SendState(0)
	}
	return nil
}

func (l *Yeelight) SetColor(state *channels.ColorState) error {
	//	log.Printf("Setting color state to %#v", state)
	var r, g, b uint8
	if state.Mode == "temperature" {
		// temperature is in the range [2000, 6500]
		//		log.Printf("Temp: %v", *state.Temperature)
		r, g, b = TemperatureToRGB(*state.Temperature)
	} else {
		// state must be "hue"
		//		log.Printf("Hue: %v, Sat: %v", *state.Hue, *state.Saturation)
		r, g, b = HSVToRGB(*state.Hue, *state.Saturation, 1)
	}
	log.Printf("Setting colour to RGB = %v, %v, %v\n", r, g, b)

	yeelight.SetColor(l.info.NaturalID, r, g, b, l.ip)
	l.colorChannel.SendState(state)

	return nil
}

// SetBrightness takes a brightness value (0-1) and calls yeelight.SetBrightness to... set the brightness
func (l *Yeelight) SetBrightness(value float64) error {
	log.Printf("Setting brightness to %f", value)
	yeelight.SetBrightness(l.info.NaturalID, value, l.ip)
	if value == 0 {
		l.state = false
	} else {
		l.state = true
	}
	l.onOffChannel.SendState(l.state)
	l.brightnessChannel.SendState(value)
	return nil
}

func (l *Yeelight) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	l.sendEvent = sendEvent
}

var reg, _ = regexp.Compile("[^a-z0-9]")

// Exported by service/device schema
func (l *Yeelight) SetName(name *string) (*string, error) {
	log.Printf("\n\n\nSetting device name to %s\n\n\n", *name)

	safe := reg.ReplaceAllString(strings.ToLower(*name), "")
	if len(safe) > 5 {
		safe = safe[0:5]
	}

	// Why is this here??
	log.Printf("Pretending we can only set 5 lowercase alphanum. Name now: %s", safe)

	l.sendEvent("renamed", safe)

	return &safe, nil
}
