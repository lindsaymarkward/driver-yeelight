package main

// Ninja Sphere driver for Yeelight Sunflower light bulbs
// originally copied from FakeDriver and modified

import (
	"fmt"
	"log"
	"time"

	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/events"
	"github.com/ninjasphere/go-ninja/support"
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

type YeelightDriver struct {
	support.DriverSupport
	config *YeelightDriverConfig
}

type YeelightDriverConfig struct {
	Hub         yeelight.Hub
	Initialised bool
}

// defaultConfig sets a default configuration for the YeelightDriverConfig with no lights
func defaultConfig() *YeelightDriverConfig {
	return &YeelightDriverConfig{
		Initialised: true,
		Hub: yeelight.Hub{
			IP:       yeelight.IP,
			LightIDs: make([]string, 0),
		},
	}
}

// setConfig creates a configuration struct for the YeelightDriverConfig with the given ip and light IDs
func setConfig(ip string, lightIDs []string) *YeelightDriverConfig {
	return &YeelightDriverConfig{
		Initialised: true,
		Hub: yeelight.Hub{
			IP:       ip,
			LightIDs: lightIDs,
		},
	}
}

func NewYeelightDriver() (*YeelightDriver, error) {

	driver := &YeelightDriver{}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize Yeelight driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export Yeelight driver: %s", err)
	}

	userAgent := driver.Conn.GetServiceClient("$device/:deviceId/channel/user-agent")
	userAgent.OnEvent("pairing-requested", driver.OnPairingRequest)

	return driver, nil
}

func (d *YeelightDriver) OnPairingRequest(pairingRequest *events.PairingRequest, values map[string]string) bool {
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

// Start (I believe) runs when the driver is started - called by the Ninja system (not the driver itself)
// gets the hub and light details, sets the configuration
func (d *YeelightDriver) Start(config *YeelightDriverConfig) error {
	log.Printf("Yeelight Driver Starting with config %v", config)

	var lightIDs []string

	d.config = config
	if !d.config.Initialised {
		// search for hub and get IP address
		ip, err := yeelight.DiscoverHub()
		if err != nil {
			log.Panicf("ERROR discovering Yeelight hub: %v", err)
			d.config = defaultConfig()
		} else {
			// found hub, set config details
			lights, _ := yeelight.GetLights()
			log.Printf("\nLights:\n%v\n", lights)
			// get just light IDs from lights array
			lightIDs = make([]string, len(lights))
			for i := 0; i < len(lights); i++ {
				lightIDs[i] = lights[i].ID
			}
			d.config = setConfig(ip, lightIDs)
			log.Printf("Found these (%d) lights: %v at IP %v", len(lights), lightIDs, ip)
		}
	}

	for i := 0; i < len(d.config.Hub.LightIDs); i++ {
		//	for i := 0; i < 0; i++ {
		log.Printf("Creating new Yeelight, %v", lightIDs[i])
		device := NewYeelight(d, lightIDs[i])

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

	//	// test!!
	//	response, _ := yeelight.SendCommand("C 50F5,255,255,255,100,0\r\n", yeelight.IP)
	//	fmt.Println(response)

	return d.SendEvent("config", config)
}

func (d *YeelightDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")
}

type In struct {
	Name string
}

type Out struct {
	Age  int
	Name string
}

func (d *YeelightDriver) Blarg(in *In) (*Out, error) {
	log.Printf("GOT INCOMING! %s", in.Name)
	return &Out{
		Name: in.Name,
		Age:  30,
	}, nil
}
