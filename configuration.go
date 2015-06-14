package main

// initially copied from driver-samsung-tv
// This file contains most of the code for the UI (i.e. what appears in the Labs)

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

func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:        "",
			Label:       "Yeelight Sunflower Bulbs",
			DisplayIcon: "lightbulb-o", // DisplayIcon should have a value from Font Awesome, without the "fa-" at the start
		},
	}, nil
}

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

		// save all/only new names to names map
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
		return c.presets() // maybe want to return to list

	case "deletePreset":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		presetToDelete = values["name"]
		return c.confirmDeletePreset()
//		c.driver.DeletePreset(values["name"])
//		return c.presets()

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
		yeelight.TurnOffAllLights(c.driver.config.IP)
		// update state of all lights for UI
		onOff := false
		for _, device := range c.driver.devices {
			device.UpdateLightState(&devices.LightDeviceState{OnOff: &onOff})
		}
		return c.list()

	case "reset":
		return c.confirmReset()

	case "confirmReset":
		var values map[string][]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		presetNames := c.driver.config.PresetNames
		presets := c.driver.config.Presets

		c.driver.config = DefaultConfig()
		c.driver.config.Initialised = false

		if containsString(values["options"], "keepPresets") {
			c.driver.config.PresetNames = presetNames
			c.driver.config.Presets = presets
		}
		c.driver.SendEvent("config", c.driver.config)

		// TODO: Find a way to restart driver here
		//		 maybe like this (from samsung-tv)?
		// no, this just exits - doesn't restart
		//		go func() {
		//			time.Sleep(time.Second * 2)
		//			os.Exit(0)
		//		}()

		return c.list()

	case "confirmDeletePreset":
		c.driver.DeletePreset(presetToDelete)
		return c.presets()

	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
}

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

	screen := suit.ConfigurationScreen{
		Title: "Yeelight - Rename/Reset Lights",
		Sections: []suit.Section{
			suit.Section{
				Title:    "Rename Lights",
				Subtitle: "Set nice names that make you happy",
				Contents: lightInputs,
			},
			// IP address setting - might want... not now
			//			suit.Section{
			//				Title: "Set IP",
			//				Contents: []suit.Typed{
			//					suit.InputText{
			//						Name:        "setIP",
			//						Before:      "Current IP",
			//						Placeholder: "IP address",
			//						Value:       c.driver.config.Hub.IP,
			//					},
			//				},
			//			},
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Title:    "Reset - ",
						Subtitle: "Clear all bulbs and (optionally) presets", // might separate these some day
						Value:    "Reset the lights (and restart the driver) if you add new bulbs.",
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
				Label:        "Reset",
				Name:         "reset",
				DisplayClass: "warning",
				DisplayIcon:  "warning",
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

func (c *configService) newPreset() (*suit.ConfigurationScreen, error) {
	//	var lights []suit.RadioGroupOption

	lights := []suit.OptionGroupOption{suit.OptionGroupOption{Title: "All lights", Value: "all"}}
	// create text field for each light
	for _, lightID := range c.driver.config.LightIDs {
		selected := false
		// Note: this is inefficient as each call to IsOn gets all lights (but not very slow)
		if isOn, _ := c.driver.devices[lightID].IsOn(); isOn {
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
				Title:    "Define Settings",
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

func (c *configService) presets() (*suit.ConfigurationScreen, error) {
	presets := []suit.ActionListOption{}

	// create action option for each preset
//	for name, _ := range c.driver.config.Presets {
	for _, name := range c.driver.config.PresetNames {
		//		title := name
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
				Subtitle: "Click to activate. To edit an existing preset, create a new one with the same name.",
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

// list displays the main screen with lights to control and links to other actions
func (c *configService) list() (*suit.ConfigurationScreen, error) {
	lightActions := []suit.ActionListOption{}

	// create action option for each light
	for _, lightID := range c.driver.config.LightIDs {
		title := c.driver.config.Names[lightID] + " (" + lightID + ") On"
		if isOn, _ := c.driver.devices[lightID].IsOn(); isOn {
			title += " *"
		}
		lightActions = append(lightActions, suit.ActionListOption{
			Title: title,
			Value: lightID,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "Yeelight Sunflower - Main",
		Sections: []suit.Section{
			// On/Off buttons for controlling or finding which lights are which
			suit.Section{
				Title:    "Switch Lights",
				Subtitle: "* indicates light is currently on",
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "lightID", // the field name for which light was clicked
						Options: lightActions,
						PrimaryAction: &suit.ReplyAction{
							Name:         "on",
							Label:        "On",
							DisplayIcon:  "toggle-on",
							DisplayClass: "success", // this doesn't change the default - can't change, it seems
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
			//			suit.Section{
			//				Contents: []suit.Typed{
			//
			//				},
			//			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			suit.ReplyAction{
				Label: "Rename/Reset",
				Name:  "rename",
				//				DisplayClass: "warning",
				DisplayIcon: "pencil",
			},
			suit.ReplyAction{
				Label:        "Presets",
				Name:         "presets",
				DisplayClass: "info",
				DisplayIcon:  "list-ul",
			},
			suit.ReplyAction{
				Label:        "All Off",
				Name:         "allOff",
				DisplayClass: "danger",
				DisplayIcon:  "toggle-off",
			},
		},
	}
	return &screen, nil
}

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
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "list",
			},
		},
	}, nil
}

func (c *configService) confirmReset() (*suit.ConfigurationScreen, error) {
	options := []suit.OptionGroupOption{suit.OptionGroupOption{Title: "Keep presets?", Value: "keepPresets"}}
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
						Name: "options",
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