package main

// Configuration for the UI (i.e. what appears in Labs)
// initially copied from driver-samsung-tv

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/lindsaymarkward/go-ninja/devices"
	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

// global variable avoids hidden fields and another unmarshal call
var presetToDelete string

type configService struct {
	driver *YeelightDriver
}

type savePresetData struct {
	Name     string   `json:name`
	LightIDs []string `json:lightIDs`
}

// GetActions is called by the Ninja Sphere system and returns the actions that this driver performs
func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:        "",
			Label:       "Yeelight Sunflower Bulbs",
			DisplayIcon: "lightbulb-o", // DisplayIcon should have a value from Font Awesome, without the "fa-" at the start
		},
	}, nil
}

// Configure is the main function that is called for every action
func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	log.Printf("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))
	switch request.Action {
	case "": // the case when coming from "main menu"
		return c.list()
	case "list":
		return c.list()

	case "saveRename":
		//		log.Printf("\nSaving with Data: %v\n", string(request.Data))
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		// save all/only new names to names map. values that start with "id" are the light name fields
		names := make(map[string]string)
		for id, newName := range values {
			if strings.HasPrefix(id, "id") {
				names[strings.TrimLeft(id, "id")] = newName
			}
		}

		err = c.driver.Rename(names)
		if err != nil {
			return c.error(fmt.Sprintf("Could not rename lights: %s", err))
		}
		return c.list()

	case "presets":
		return c.presets()

	case "newPreset":
		return c.newPreset()

	case "savePreset":
		log.Printf("\nSaving with Data: %v\n", string(request.Data))
		values := &savePresetData{}
		err := json.Unmarshal(request.Data, values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		err = c.driver.SavePreset(values)
		if err != nil {
			return c.error(fmt.Sprintf("Could not save preset: %s", err))
		}
		return c.presets()

	case "presetOn":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		c.driver.ActivatePreset(values["name"])
		return c.presets()

	case "deletePreset":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		// set global variable then go to confirmation screen
		presetToDelete = values["name"]
		return c.confirmDeletePreset()

	case "rename":
		return c.rename()

	case "on":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		c.driver.devices[values["lightID"]].SetOnOff(true)
		return c.list()

	case "off":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		c.driver.devices[values["lightID"]].SetOnOff(false)

		return c.list()

	case "allOff":
		// turn off all lights and if no error, update state of all lights for UI
		if c.driver.TurnOffAllLights() == nil {
			onOff := false
			for _, device := range c.driver.devices {
				device.UpdateLightState(&devices.LightDeviceState{OnOff: &onOff})
			}
		}
		return c.list()

	case "resetAction":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		// handle the ActionList choice - reset or scan/update
		switch values["choice"] {
		case "reset":
			return c.confirmReset()
		case "scanNew":
			if err := c.driver.ScanLightsToConfig(); err != nil {
				return c.error(fmt.Sprintf("%v", err))
			}
			c.driver.SendEvent("config", c.driver.config)
			// TODO: (as in other places) This doesn't make new devices, just config entries... Somehow need to make new devices without re-making existing devices
			//			c.driver.CreateDevicesFromConfig()
			return c.list()

		default:
			return c.error(fmt.Sprintf("Unknown option %v\n", values["choice"]))
		}

	case "refresh":
		if err := c.driver.ScanLightsToConfig(); err != nil {
			return c.error(fmt.Sprintf("%v", err))
		}
		c.driver.SendEvent("config", c.driver.config)
		return c.list()

	case "setip":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		c.driver.config.IP = values["ip"]
		// ?? set initialised to true?
		c.driver.SendEvent("config", c.driver.config)
		err = c.driver.ScanLightsToConfig()
		if err != nil {
			return c.error(fmt.Sprintf("%v", err))
		}
		c.driver.CreateDevicesFromConfig()
		return c.list()

	case "confirmReset":
		var values map[string][]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		presetNames := c.driver.config.PresetNames
		presets := c.driver.config.Presets

		c.driver.config = DefaultConfig()

		if containsString(values["options"], "keepPresets") {
			c.driver.config.PresetNames = presetNames
			c.driver.config.Presets = presets
		}
		// scan for new lights
		if err := c.driver.ScanLightsToConfig(); err != nil {
			return c.error(fmt.Sprintf("%v", err))
		}
		c.driver.SendEvent("config", c.driver.config)
		return c.list()

	case "confirmDeletePreset":
		c.driver.DeletePreset(presetToDelete)
		return c.presets()

	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
}

// rename is a config screen containing text entry fields for each light name, plus buttons to reset or scan
func (c *configService) rename() (*suit.ConfigurationScreen, error) {
	lightInputs := []suit.Typed{}
	// create text field for each light
	for _, lightID := range c.driver.config.LightIDs {
		name := "id" + lightID // create name field from ID so each name is unique
		lightInputs = append(lightInputs, suit.InputText{
			Name:        name,
			Before:      lightID,
			Placeholder: "Custom name",
			Value:       c.driver.config.Names[lightID],
		})
	}
	choices := []suit.ActionListOption{}
	choices = append(choices, suit.ActionListOption{Title: "Reset Lights", Value: "reset"})
	choices = append(choices, suit.ActionListOption{Title: "Scan for New Lights", Value: "scanNew"})

	screen := suit.ConfigurationScreen{
		Title: "Yeelight - Rename/Reset Lights",
		Sections: []suit.Section{
			suit.Section{
				Title:    "Rename Lights",
				Subtitle: "Set nice names that make you happy",
				Contents: lightInputs,
			},
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Value: "Reset clears all bulbs and (optionally) presets, then scans for current lights",
					},
					suit.StaticText{
						Value: "Driver version: " + c.driver.DriverSupport.Info.Version,
					},
					suit.ActionList{
						Name:    "choice",
						Options: choices,
						PrimaryAction: &suit.ReplyAction{
							Name: "resetAction",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "list",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "saveRename",
				DisplayClass: "success",
				DisplayIcon:  "save",
			},
		},
	}
	return &screen, nil
}

// newPreset is a config screen for creating new presets/scenes, including selecting which lights to include
func (c *configService) newPreset() (*suit.ConfigurationScreen, error) {
	onLights := c.determineOnLights()
	lights := []suit.OptionGroupOption{suit.OptionGroupOption{Title: "All lights", Value: "all"}}
	// create option check box for each light
	for _, lightID := range c.driver.config.LightIDs {
		selected := false

		if c.isLightOn(lightID, onLights) {
			selected = true
		}
		lights = append(lights, suit.OptionGroupOption{
			Title:    c.driver.config.Names[lightID],
			Value:    lightID,
			Selected: selected,
		})
	}
	screen := suit.ConfigurationScreen{
		Title: "Yeelight - Create New Preset",
		Sections: []suit.Section{
			suit.Section{
				Title:    "Choose Settings",
				Subtitle: "With your lights as you want the scene, select which lights to include below, and click Save. (Lights that are currently on are pre-selected.) Including lights that are currently off will turn them off when the preset is activated.",
				Contents: []suit.Typed{
					suit.InputText{
						Name:        "name",
						Before:      "Preset Name",
						Placeholder: "(unique name)",
					},
					suit.OptionGroup{
						Title:   "Lights to Include",
						Name:    "lightIDs",
						Options: lights,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "presets",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "savePreset",
				DisplayClass: "success",
				DisplayIcon:  "save",
			},
		},
	}
	return &screen, nil
}

// presets is a config screen that displays current presets and allows you to activate them,
// plus a button to create a new preset
func (c *configService) presets() (*suit.ConfigurationScreen, error) {
	presets := []suit.ActionListOption{}

	// create action option for each preset
	for _, name := range c.driver.config.PresetNames {
		presets = append(presets, suit.ActionListOption{
			Title: name,
			Value: name,
		})
	}
	screen := suit.ConfigurationScreen{
		Title: "Yeelight - Presets",
		Sections: []suit.Section{
			suit.Section{
				Title:    "Current Presets",
				Subtitle: "Click to activate scene. To change an existing preset, create a new one with the same name.",
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "name", // the field name for which preset was clicked
						Options: presets,
						PrimaryAction: &suit.ReplyAction{
							Name:        "presetOn",
							Label:       "On",
							DisplayIcon: "toggle-on",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "deletePreset",
							Label:        "Delete",
							DisplayIcon:  "trash",
							DisplayClass: "danger",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Back",
				Name:  "list",
			},
			suit.ReplyAction{
				Label:        "New Preset",
				Name:         "newPreset",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}
	return &screen, nil
}

// list displays the main screen with lights to control, plus buttons for other main actions
func (c *configService) list() (*suit.ConfigurationScreen, error) {
	var screen suit.ConfigurationScreen
	err := c.driver.CheckHub()
	if err != nil {
		// if hub is not responding, produce custom error screen with option to refresh/rescan
		log.Printf("Error: %v\n", err)
		screen = suit.ConfigurationScreen{
			Title: "Yeelight Sunflower - Error",
			Sections: []suit.Section{
				suit.Section{
					Title: "Hub is not responding. Check that the Yeelight hub is connected and switched on, then click Refresh below.",
					Contents: []suit.Typed{
						suit.StaticText{
							Value: "Driver version: " + c.driver.DriverSupport.Info.Version,
						},
						suit.StaticText{
							Value: "You could try setting the IP manually...",
						},
						suit.InputText{
							Title: "IP address",
							Name:  "ip",
							Value: c.driver.config.IP,
						},
					},
				},
			},
			Actions: []suit.Typed{
				suit.CloseAction{
					Label: "Close",
				},
				suit.ReplyAction{
					Label:        "Refresh (Scan for hub & bulbs)",
					Name:         "refresh",
					DisplayIcon:  "refresh",
					DisplayClass: "warning",
				},
				suit.ReplyAction{
					Label:        "Set IP",
					Name:         "setip",
					DisplayIcon:  "wifi",
					DisplayClass: "success",
				},
			},
		}
	} else {
		// hub is alive so make normal list
		var lightActions []suit.ActionListOption
		onLights := c.determineOnLights()

		// create action option for each light
		for _, lightID := range c.driver.config.LightIDs {
			title := c.driver.config.Names[lightID]
			if c.isLightOn(lightID, onLights) {
				title += " *"
			}
			lightActions = append(lightActions, suit.ActionListOption{
				Title:    title,
				Subtitle: lightID,
				Value:    lightID,
			})
		}

		screen = suit.ConfigurationScreen{
			Title: "Yeelight Sunflower - Main",
			Sections: []suit.Section{
				// On/Off buttons for controlling all lights
				suit.Section{
					Title:    "Switch Lights",
					Subtitle: "* indicates light is currently on",
					Contents: []suit.Typed{
						suit.ActionList{
							Name:    "lightID", // the field name for which light was clicked
							Options: lightActions,
							PrimaryAction: &suit.ReplyAction{
								Name:        "on",
								Label:       "On",
								DisplayIcon: "toggle-on",
							},
							SecondaryAction: &suit.ReplyAction{
								Name:         "off",
								Label:        "Off",
								DisplayIcon:  "toggle-off",
								DisplayClass: "danger",
							},
						},
					},
				},
				// extra button for turning off all lights
				suit.Section{
					Contents: []suit.Typed{
						suit.ActionList{
							Name:    "allOff",
							Options: []suit.ActionListOption{suit.ActionListOption{Title: "Turn All Lights Off"}},
							PrimaryAction: &suit.ReplyAction{
								Name:        "allOff",
								DisplayIcon: "toggle-off",
							},
						},
					},
				},
			},
			Actions: []suit.Typed{
				suit.CloseAction{
					Label: "Close",
				},
				suit.ReplyAction{
					Label:       "Rename/Reset",
					Name:        "rename",
					DisplayIcon: "pencil",
				},
				suit.ReplyAction{
					Label:        "Presets",
					Name:         "presets",
					DisplayClass: "info",
					DisplayIcon:  "list-ul",
				},
			},
		}
	}
	return &screen, nil
}

// error is a generic error message screen
func (c *configService) error(message string) (*suit.ConfigurationScreen, error) {

	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Error",
						Subtitle:     message,
						DisplayClass: "danger",
					},
					suit.StaticText{
						Value: "Driver version: " + c.driver.DriverSupport.Info.Version,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Back",
				Name:  "list",
			},
		},
	}, nil
}

// confirmReset is a config screen for confirming/cancelling reset of driver configuration
func (c *configService) confirmReset() (*suit.ConfigurationScreen, error) {
	options := []suit.OptionGroupOption{suit.OptionGroupOption{Title: "Keep presets?", Value: "keepPresets", Selected: true}}
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Confirm Reset",
						Subtitle:     "Do you really want to reset the configuration? This will clear all custom light names and presets. Choose the option below to preserve presets and just clear bulbs",
						DisplayClass: "danger",
						DisplayIcon:  "warning",
					},
					suit.OptionGroup{
						Name:    "options",
						Options: options,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label:       "Cancel",
				Name:        "list",
				DisplayIcon: "close",
			},
			suit.ReplyAction{
				Label:        "Confirm - Reset",
				Name:         "confirmReset",
				DisplayClass: "warning",
				DisplayIcon:  "check",
			},
		},
	}, nil
}

// confirmDeletePreset is a config screen to confirm/cancel deleting a preset
func (c *configService) confirmDeletePreset() (*suit.ConfigurationScreen, error) {
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Confirm Delete Preset",
						Subtitle:     "Do you really want to delete this preset?",
						DisplayClass: "danger",
						DisplayIcon:  "warning",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label:       "Cancel",
				Name:        "presets",
				DisplayIcon: "close",
			},
			suit.ReplyAction{
				Label:        "Confirm",
				Name:         "confirmDeletePreset",
				DisplayClass: "warning",
				DisplayIcon:  "check",
			},
		},
	}, nil
}

// determineOnLights gets current light values and returns a list of the IDs of lights that are on
func (c *configService) determineOnLights() []string {
	var lightData []yeelight.Light
	var onLightIDs []string
	// get light data
	lightData, _ = yeelight.GetLights(c.driver.config.IP)
	for _, light := range lightData {
		if light.Level > 0 {
			onLightIDs = append(onLightIDs, light.ID)
		}
	}
	return onLightIDs
}

// isLightOn returns true if lightID is in list of lights that are on (passed in)
func (c *configService) isLightOn(lightID string, lights []string) bool {
	if containsString(lights, lightID) {
		return true
	}
	return false
}
