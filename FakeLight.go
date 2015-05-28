package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/lindsaymarkward/go-yeelight"
)

type Yeelight struct {
	driver             ninja.Driver
	info               *model.Device
	sendEvent          func(event string, payload interface{}) error
	onOffChannel       *channels.OnOffChannel
	brightnessChannel  *channels.BrightnessChannel
	colorChannel       *channels.ColorChannel
	temperatureChannel *channels.TemperatureChannel
}

func NewYeelight(driver ninja.Driver, id string) *Yeelight {
	name := fmt.Sprintf("Light %v", id)

	light := &Yeelight{
		driver: driver,
		info: &model.Device{
			// should I store the light ID here or outside info??
			// Is NaturalID meant to be something in particular??
			NaturalID:     fmt.Sprintf("%s", id), // was: fmt.Sprintf("light%d", id),
			NaturalIDType: "yeelight", // I have no idea about this. was "fake"
			Name:          &name,
			Signatures: &map[string]string{
				"ninja:manufacturer": "Qingdao Yeelink",
				"ninja:productName":  "Yeelight",
				"ninja:productType":  "Light",
				"ninja:thingType":    "light",
			},
		},
	}

	// what channels?? remove temp
	light.onOffChannel = channels.NewOnOffChannel(light)
	light.brightnessChannel = channels.NewBrightnessChannel(light)
	light.colorChannel = channels.NewColorChannel(light)
	light.temperatureChannel = channels.NewTemperatureChannel(light) // remove??

	// I'm pretty sure this is just testing events.
	// Could potentially use heartbeat here if it's useful??
	//	go func() {
	//
	//		var temp float64
	//		for {
	//			time.Sleep(5 * time.Second)
	//			temp += 0.5
	//			light.temperatureChannel.SendState(temp)
	//		}
	//	}()

	return light
}

func (l *Yeelight) GetDeviceInfo() *model.Device {
	return l.info
}

func (l *Yeelight) GetDriver() ninja.Driver {
	return l.driver
}

// is this where the action happens??

func (l *Yeelight) SetOnOff(state bool) error {
	log.Printf("Turning %t", state)
	// turn light on/off (yeelight.SetOnOff handles state choice)
	yeelight.SetOnOff(l.info.NaturalID, state)
	return nil
}

func (l *Yeelight) ToggleOnOff() error {
	log.Println("Toggling")
	return nil
}

func (l *Yeelight) SetColor(state *channels.ColorState) error {
	log.Printf("setting color state to %v", state)
	return nil
}

func (l *Yeelight) SetBrightness(state float64) error {
	log.Printf("setting brightness to %f", state)
	return nil
}

func (l *Yeelight) SetEventHandler(sendEvent func(event string, payload interface{}) error) {
	l.sendEvent = sendEvent
}

var reg, _ = regexp.Compile("[^a-z0-9]")

// Exported by service/device schema
func (l *Yeelight) SetName(name *string) (*string, error) {
	log.Printf("Setting device name to %s", *name)

	safe := reg.ReplaceAllString(strings.ToLower(*name), "")
	if len(safe) > 5 {
		safe = safe[0:5]
	}

	// Why is this here??
	log.Printf("Pretending we can only set 5 lowercase alphanum. Name now: %s", safe)

	l.sendEvent("renamed", safe)

	return &safe, nil
}
