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
	Initialised bool
	IP          string
	LightIDs    []string // need slices in addition to maps due to being ordered
	Names       map[string]string
	PresetNames []string
	Presets     map[string]*Preset
}

type Preset struct {
	//	Name   string	// name is the key in the map of Presets
	Lights []yeelight.Light // TODO: consider storing only what we need, not everything (LQI, Effect...)
}

// DefaultConfig sets a default configuration for the YeelightDriverConfig with no lights
func DefaultConfig() *YeelightDriverConfig {
	return &YeelightDriverConfig{
		Initialised: false,
		IP:          "",
		LightIDs:    make([]string, 0),
		Names:       make(map[string]string),
		Presets:     make(map[string]*Preset),
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

	return driver, nil
}

// Start runs when the driver is started - called by the Ninja system (not the driver itself),
// gets the hub and light details, sets the configuration
func (d *YeelightDriver) Start(config *YeelightDriverConfig) error {
	log.Printf("Yeelight Driver Starting with config %v", config)
	d.config = config

	if !d.config.Initialised {
		// Preserve presets if they exist
		presetNames := config.PresetNames
		presets := config.Presets
		d.config = DefaultConfig()
		if len(presetNames) > 0 {
			log.Printf("Preserving presets: %v\n", presetNames)
			d.config.Presets = presets
			d.config.PresetNames = presetNames
		}

		// find hub and lights, save to config
		// if it worked, the driver is initialised
		if err := d.ScanLightsToConfig(); err != nil {
			d.config.Initialised = true
		}
	}
	log.Printf("\nLightIDs: %v\nNames: %v\n", d.config.LightIDs, d.config.Names)

	// create devices from the lights stored in the config
	// this creates devices even if the hub is not online so they can be used when it does come online
	d.CreateDevicesFromConfig()

	// TODO: trying to set ThingIDs so we can set Thing.Name
	// can get access to it but setting it doesn't do anything
	// I think I need a sendEvent like for config

	//	thingClient := d.Conn.GetServiceClient("$home/services/ThingModel")
	//	things := make([]*model.Thing, 0)
	//	keptThings := make([]*model.Thing, 0, len(things))
	//	if err := thingClient.Call("fetchAll", nil, &things, 10*time.Second); err != nil {
	//		fmt.Printf("ERROR getting things, %v\n", err)
	//	}
	//
	//	fmt.Printf("\n\n")
	//	for _, thing := range things {
	//		if thing.Type == "light" {
	//			keptThings = append(keptThings, thing)
	//			fmt.Printf("> %v (%v)\n", *thing.DeviceID, thing.Name)
	//			if thing.Name == "Lt 143F" {
	//				thing.Name = "Yee 143F"
	//				fmt.Printf("\n\n!! %v\n", thing.Name)
	//			}
	//		}
	//	}

	// Provide configuration (Labs) service
	d.Conn.MustExportService(&configService{d}, "$driver/"+info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	return d.SendEvent("config", config)
}

// ScanLightsToConfig finds the Yeelight hub on the network, gets the lights and saves them to the config
func (d *YeelightDriver) ScanLightsToConfig() error {
	// search for hub and get IP address
	ip, err := yeelight.DiscoverHub()
	if err != nil {
		log.Printf("ERROR discovering Yeelight hub: %v", err)
		return err
	}
	// found hub, get lights and set config details
	log.Printf("Hub found at %s\n", ip)
	lights, _ := yeelight.GetLights(ip)
	// TODO: Consider how to remove a light. Maybe offer delete option to ignore these lights (which would still appear in the Yeelight iPhone app)
	// Create entries in Names map (light IDs from lights slice as keys) and LightIDs slice
	for _, light := range lights {
		// set default name for new lights, like "Yee238B"
		if !containsString(d.config.LightIDs, light.ID) {
			d.config.Names[light.ID] = "Yee" + light.ID
			d.config.LightIDs = append(d.config.LightIDs, light.ID)
		}
	}
	// save IP address to config and "initialise" driver
	d.config.IP = ip
	d.config.Initialised = true
	log.Printf("Found these (%d) lights: %v at IP %v", len(lights), d.config.LightIDs, ip)
	return err
}

// I think Stop should be different ??
func (d *YeelightDriver) Stop() error {
	return fmt.Errorf("This driver does not support being stopped. YOU HAVE NO POWER HERE.")
}

// CreateDevicesFromConfig creates a new device (exporting it) for each light in the config
func (d *YeelightDriver) CreateDevicesFromConfig() error {
	// create device for each light and add it to devices map in driver
	for id, _ := range d.config.Names {
		log.Printf("Creating new Yeelight, %v", id)
		device := NewYeelightDevice(d, id)
		d.devices[id] = device
	}
	return nil
}

// Rename takes a map of id->name and changes the display names for each light
func (d *YeelightDriver) Rename(names map[string]string) error {
	d.config.Names = names
	// as well as the driver config, we also need to set the device names
	for id, newName := range names {
		d.devices[id].GetDeviceInfo().Name = &newName
	}
	// save the new configuration
	return d.SendEvent("config", d.config)
}

// SavePreset takes the data from the configuration and saves a new preset as a slice of light values
func (d *YeelightDriver) SavePreset(values *savePresetData) error {
	lightStates, err := yeelight.GetLights(d.config.IP)
	if err != nil {
		return err
	}
	// handle "all" selection
	lightsToSet := values.LightIDs
	if values.LightIDs[0] == "all" {
		lightsToSet = d.config.LightIDs
	}
	// save name to slice of names so we can have consistent order on presets page
	// unless it already exists (updating the preset)
	if !containsString(d.config.PresetNames, values.Name) {
		d.config.PresetNames = append(d.config.PresetNames, values.Name)
	}
	// create blank preset to save to
	d.config.Presets[values.Name] = &Preset{Lights: make([]yeelight.Light, len(values.LightIDs))}
	// for each light in preset
	for _, lightID := range lightsToSet {
		for _, light := range lightStates {
			if lightID == light.ID {
				d.config.Presets[values.Name].Lights = append(d.config.Presets[values.Name].Lights, light)
				break
			}
		}
	}

	log.Printf("Saving preset: %v\n", values.Name)
	//	log.Printf("Current presets: %v\n", d.config.Presets)
	// save the new configuration
	return d.SendEvent("config", d.config)
}

// DeletePreset takes the name of a preset and deletes it from the config
func (d *YeelightDriver) DeletePreset(name string) error {
	// delete from map
	delete(d.config.Presets, name)
	// delete from slice
	i := pos(d.config.PresetNames, name)
	d.config.PresetNames = append(d.config.PresetNames[:i], d.config.PresetNames[i+1:]...)

	// save the new configuration
	return d.SendEvent("config", d.config)

}

// ActivatePreset takes the name of a preset and sets the lights to match the values stored
// only changes the lights the preset stores values for
func (d *YeelightDriver) ActivatePreset(name string) error {
	log.Printf("Activating preset: %v", name)
	lights := d.config.Presets[name].Lights
	var err error
	for _, light := range lights {
		err = yeelight.SetLight(light.ID, light.R, light.G, light.B, light.Level, d.config.IP)
	}
	return err
}

// CheckHub calls Heartbeat which pings the Yeelight hub to see if it's alive,
// returns either nil error if it's responsive or error if the ack is not received from the hub.
func (d *YeelightDriver) CheckHub() error {
	return yeelight.Heartbeat(d.config.IP)
}

// TurnOffAllLights turns off all bulbs
func (d *YeelightDriver) TurnOffAllLights() error {
	return yeelight.TurnOffAllLights(d.config.IP)
}

func containsString(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// pos finds the position of a value in a slice, returns -1 if not found
func pos(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}
