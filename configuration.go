package main

// initially copied from driver-samsung-tv
// This file contains most of the code for the UI (i.e. what appears in the Labs)

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/lindsaymarkward/go-yeelight"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

type configService struct {
	driver *YeelightDriver
}

func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:        "",
			Label:       "Yeelight",
			DisplayIcon: "lightbulb-o", // DisplayIcon should have a value from fontawesome, without the "fa-" at the start
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
	case "save":
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
		// IP?? set it? seems unnecessary

		err = c.driver.Rename(names)

		if err != nil {
			return c.error(fmt.Sprintf("Could not rename lights: %s", err))
		}

		// go back instead? - how ??
		return c.list()

	case "on":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		//		yeelight.SetOnOff(values["lightID"], true, c.driver.config.Hub.IP)
		c.driver.devices[values["lightID"]].SetOnOff(true)
		return c.list()

	case "off":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		//		yeelight.SetOnOff(values["lightID"], false, c.driver.config.Hub.IP)
		c.driver.devices[values["lightID"]].SetOnOff(false)

		return c.list()

	case "allOff":
		yeelight.TurnOffAllLights(c.driver.config.Hub.IP)
		// sendState of all lights
		for _, device := range c.driver.devices {
			device.onOffChannel.SendState(false)
		}
		return c.list()

	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
}

func (c *configService) list() (*suit.ConfigurationScreen, error) {

	lightActions := []suit.ActionListOption{}
	lightInputs := []suit.Typed{}

	// create text field and action button for each light
	for _, lightID := range c.driver.config.Hub.LightIDs {
		name := "id" + lightID // create name field from ID so each name is unique
		lightInputs = append(lightInputs, suit.InputText{
			Name:        name,
			Before:      lightID,
			Placeholder: "Custom name",
			Value:       c.driver.config.Names[lightID],
		})
		lightActions = append(lightActions, suit.ActionListOption{
			Title: c.driver.config.Names[lightID] + " (" + lightID + ") On",
			Value: lightID,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "Yeelight",
		Sections: []suit.Section{
			suit.Section{
				Title:    "Rename Lights",
				Contents: lightInputs,
			},

			// On/Off buttons for testing to find which lights are which
			suit.Section{
				Title: "Switch Lights",
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "lightID", // the field name for which light was clicked
						Options: lightActions,
						PrimaryAction: &suit.ReplyAction{
							Name:         "on",
							Label:        "On",
							DisplayIcon:  "toggle-on",
							DisplayClass: "success", // this doesn't change the default
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
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			// TODO: Add a reset/rescan button - refactor part of Start out and use that
			suit.ReplyAction{
				Label:        "All Off",
				Name:         "allOff",
				DisplayClass: "danger",
				DisplayIcon:  "toggle-off",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "save",
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
