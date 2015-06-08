package main

// Ninja Sphere driver for Yeelight Sunflower light bulbs

import (
	"fmt"
	"log"

	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")

type YeelightDriver struct {
	support.DriverSupport
	config  *YeelightDriverConfig
	devices map[string]*YeelightDevice
}

type YeelightDriverConfig struct {
	IP          string
	LightIDs    []string // needed in addition to map of Names due to being ordered
	Names       map[string]string
	Initialised bool
}

// defaultConfig sets a default configuration for the YeelightDriverConfig with no lights
func DefaultConfig() *YeelightDriverConfig {
	return &YeelightDriverConfig{
		Initialised: true,
		LightIDs:    make([]string, 0),
		IP:          "",
		Names:       make(map[string]string),
	}
}

// NewYeelightDriver creates a new driver with an empty map of names
// initialises and exports Ninja stuff
func NewYeelightDriver() (*YeelightDriver, error) {

	driver := &YeelightDriver{
		// make map of devices so we can add lights to it
		devices: make(map[string]*YeelightDevice),
	}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize Yeelight driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export Yeelight driver: %s", err)
	}

	//	userAgent := driver.Conn.GetServiceClient("$device/:deviceId/channel/user-agent")
	//	userAgent.OnEvent("pairing-requested", driver.OnPairingRequest)

	return driver, nil
}

// Start runs when the driver is started - called by the Ninja system (not the driver itself)
// gets the hub and light details, sets the configuration
func (d *YeelightDriver) Start(config *YeelightDriverConfig) error {
	log.Printf("Yeelight Driver Starting with config %v", config)

	d.config = config
	if !d.config.Initialised {
		// search for hub and get IP address
		ip, err := yeelight.DiscoverHub()
		if err != nil {
			log.Panicf("ERROR discovering Yeelight hub: %v", err)
			d.config = DefaultConfig()
		} else {
			// found hub, get lights and set config details
			lights, _ := yeelight.GetLights(ip)
			log.Printf("\nLights:\n%v\n", lights)
			// Create Names map entries with light IDs from lights array as keys
			for _, light := range lights {
				// set default name for new lights
				d.config.Names[light.ID] = "Yee" + light.ID
				d.config.LightIDs = append(d.config.LightIDs, light.ID)

			}
			// save IP address to config
			d.config.IP = ip
			d.config.Initialised = true
			log.Printf("Found these (%d) lights: %v at IP %v", len(lights), d.config.LightIDs, ip)
		}
	} else {
		fmt.Println(d.config.LightIDs)
	}

	log.Printf("\n\nLightIDs: %v\nd.config.Names: %v\n\n", d.config.LightIDs, d.config.Names)

	// create device for each light and add it to devices map in driver
	for id, _ := range d.config.Names {
		log.Printf("Creating new Yeelight, %v", id)
		device := NewYeelightDevice(d, id, d.config.IP)
		d.devices[id] = device
	}

	// Provide configuration page (labs)
	d.Conn.MustExportService(&configService{d}, "$driver/"+info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	return d.SendEvent("config", config)
}

// I think Stop should be different
func (d *YeelightDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")
}

// Rename takes a map of id->name and changes the display names for each light
func (d *YeelightDriver) Rename(names map[string]string) error {
	d.config.Names = names
	// as well as the driver config, we also need to set the device names
	for id, newName := range names {
		//		fmt.Printf("\nid %v, name %v\n", id, newName)
		d.devices[id].GetDeviceInfo().Name = &newName
		//		d.devices[id].sendEvent("renamed", &newName)

	}
	// save the new configuration
	return d.SendEvent("config", d.config)
}

// I think I don't need this! (not in samsung-tv or lifx)
//func (d *YeelightDriver) OnPairingRequest(pairingRequest *events.PairingRequest, values map[string]string) bool {
//	log.Printf("Pairing request received from %s for %d seconds", values["deviceId"], pairingRequest.Duration)
//	d.SendEvent("pairing-started", &events.PairingStarted{
//		Duration: pairingRequest.Duration,
//	})
//	go func() {
//		time.Sleep(time.Second * time.Duration(pairingRequest.Duration))
//		d.SendEvent("pairing-ended", &events.PairingStarted{
//			Duration: pairingRequest.Duration,
//		})
//	}()
//	return true
//}
