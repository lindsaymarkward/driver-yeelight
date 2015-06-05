package main

// initially copied from driver-samsung-tv
// This file contains most of the code for the UI (i.e. what appears in the Labs)

import (
	"fmt"

	"log"

	"encoding/json"

	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
	"strings"
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
	case "list":
		return c.list()
	case "":
		if len(c.driver.config.Hub.LightIDs) > 0 {
			return c.list()
		}
		fallthrough
	case "new":
		return c.edit(YeelightDriverConfig{})
	case "edit":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)
		fmt.Printf("\n\n%#v\n\n", vals)
		//		config := c.driver.config.get(vals["tv"])
		config := c.driver.config
		if config == nil {
			return c.error(fmt.Sprintf("Could not edit: %s", vals["tv"]))
		}

		return c.edit(*config)

	// TODO: need to set Name in model.Device
	case "toggle":
		log.Printf(string(request.Data))
		return c.list() // ??

		//	case "delete":
		//
		//		var vals map[string]string
		//		json.Unmarshal(request.Data, &vals)
		//
		//		err := c.driver.deleteTV(vals["tv"])
		//
		//		if err != nil {
		//			return c.error(fmt.Sprintf("Failed to delete tv: %s", err))
		//		}
		//
		//		return c.list()
	case "save":
		log.Printf("\nSaving with Data: %v\n", string(request.Data))
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
	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
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

func (c *configService) list() (*suit.ConfigurationScreen, error) {

	//	var lights []suit.ActionListOption
	//
	//	for _, lightID := range c.driver.config.Hub.LightIDs {
	//		lights = append(lights, suit.ActionListOption{
	//			Title: lightID,
	//			Value: c.driver.config.Names[lightID],
	//		})
	//	}
	lightInputs := []suit.Typed{
	}
	for _, lightID := range c.driver.config.Hub.LightIDs {
		name := "id" + lightID // create name field from ID so each name is unique
		lightInputs = append(lightInputs, suit.InputText{
			Name:        name,
			Before:      lightID,
			Placeholder: "Custom name",
			Value:       c.driver.config.Names[lightID],
		})
	}
	screen := suit.ConfigurationScreen{
		Title: "Yeelight",
		Sections: []suit.Section{
			suit.Section{
				Title:    "Rename Lights",
				Contents: lightInputs,
				//				Contents: []suit.Typed{
				//					suit.ActionList{
				//						Name:    "lights",
				//						Options: lights,
				//						PrimaryAction: &suit.ReplyAction{
				//							Name:        "edit",
				//							DisplayIcon: "pencil",
				//						},
				//						SecondaryAction: &suit.ReplyAction{
				//							Name:        "toggle",
				//							Label:       "Toggle",
				//							DisplayIcon: "lightbulb",
				//						},
				//					},
				//				},
			},
			suit.Section{
				Contents: []suit.Typed{
					suit.InputText{
						Name:        "ip",
						Before:      "IP",
						Placeholder: "IP address",
						Value:       c.driver.config.Hub.IP,
					},
// test button??
//					suit.ReplyAction{
//						Name:        "toggle",
//						Label:       "Toggle",
//						DisplayIcon: "lightbulb-o",
//					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

func (c *configService) edit(config YeelightDriverConfig) (*suit.ConfigurationScreen, error) {

	title := "Editing Yeelight"
	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.InputHidden{
						Name:  "id",
						Value: "", // config.??
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "My TV",
						Value:       "TEST NAME", //config.Name,
					},
					suit.InputText{
						Name:        "host",
						Before:      "Host",
						Placeholder: "IP",
						Value:       config.Hub.IP,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Cancel",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}
	return &screen, nil
}

//func i(i int) *int {
//	return &i
//}
//
//func contains(s []string, e string) bool {
//	for _, a := range s {
//		if a == e {
//			return true
//		}
//	}
//	return false
//}
