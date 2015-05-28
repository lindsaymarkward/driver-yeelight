package main

// Ninja Sphere driver for Yeelight Sunflower light bulbs
// originally copied from FakeDriver and modified

import (
	"fmt"
	"log"
	"time"

	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/events"
	"github.com/ninjasphere/go-ninja/support"
	"github.com/lindsaymarkward/go-yeelight"
)

var info = ninja.LoadModuleInfo("./package.json")

/*model.Module{
	ID:          "com.ninjablocks.fakedriver",
	Name:        "Fake Driver",
	Version:     "1.0.2",
	Description: "Just used to test go-ninja",
	Author:      "Elliot Shepherd <elliot@ninjablocks.com>",
	License:     "MIT",
}*/

type FakeDriver struct {
	support.DriverSupport
	config *FakeDriverConfig
}

type FakeDriverConfig struct {
	Initialised          bool
	NumberOfLights       int
}

func defaultConfig() *FakeDriverConfig {
	return &FakeDriverConfig{
		Initialised:          true,
		NumberOfLights:       2,
	}
}

func NewFakeDriver() (*FakeDriver, error) {

	driver := &FakeDriver{}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize fake driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export fake driver: %s", err)
	}

	userAgent := driver.Conn.GetServiceClient("$device/:deviceId/channel/user-agent")
	userAgent.OnEvent("pairing-requested", driver.OnPairingRequest)

	return driver, nil
}

func (d *FakeDriver) OnPairingRequest(pairingRequest *events.PairingRequest, values map[string]string) bool {
	log.Printf("Pairing request received from %s for %d seconds", values["deviceId"], pairingRequest.Duration)
	d.SendEvent("pairing-started", &events.PairingStarted{
		Duration: pairingRequest.Duration,
	})
	go func() {
		time.Sleep(time.Second * time.Duration(pairingRequest.Duration))
		d.SendEvent("pairing-ended", &events.PairingStarted{
			Duration: pairingRequest.Duration,
		})
	}()
	return true
}

func (d *FakeDriver) Start(config *FakeDriverConfig) error {
	log.Printf("Yeelight Driver Starting with config %v", config)

	d.config = config
	if !d.config.Initialised {
		d.config = defaultConfig()
	}

	lights, _ := yeelight.GetLights()
	fmt.Println(lights)

	for i := 0; i < d.config.NumberOfLights; i++ {
		log.Print("Creating new Yeelight!!!!!!!!!!!!       !!!!!!!")
		device := NewFakeLight(d, i)

		err := d.Conn.ExportDevice(device)
		if err != nil {
			log.Fatalf("Failed to export fake light %d: %s", i, err)
		}

		err = d.Conn.ExportChannel(device, device.onOffChannel, "on-off")
		if err != nil {
			log.Fatalf("Failed to export fake light on off channel %d: %s", i, err)
		}

		err = d.Conn.ExportChannel(device, device.brightnessChannel, "brightness")
		if err != nil {
			log.Fatalf("Failed to export fake light brightness channel %d: %s", i, err)
		}

		err = d.Conn.ExportChannel(device, device.colorChannel, "color")
		if err != nil {
			log.Fatalf("Failed to export fake color channel %d: %s", i, err)
		}
		// can I remove this?
		err = d.Conn.ExportChannel(device, device.temperatureChannel, "temperature")
		if err != nil {
			log.Fatalf("Failed to export fake light temperature channel %d: %s", i, err)
		}


	}

	// Bump the config prop by one... to test it updates
	config.NumberOfLights++

	// test!!
	response, _ := yeelight.SendCommand("C 50F5,255,255,255,100,0\r\n", yeelight.IP)
	fmt.Println(response)

	return d.SendEvent("config", config)
}

func (d *FakeDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")
}

type In struct {
	Name string
}

type Out struct {
	Age  int
	Name string
}

func (d *FakeDriver) Blarg(in *In) (*Out, error) {
	log.Printf("GOT INCOMING! %s", in.Name)
	return &Out{
		Name: in.Name,
		Age:  30,
	}, nil
}
