package main

// initially copied from driver-samsung-tv
// This file contains most of the code for the UI (i.e. what appears in the Labs)
/*
import (
	"encoding/json"
	"fmt"

	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

type configService struct {
	driver *Driver
}

func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:  "",
			Label: "Samsung TVs",
		},
	}, nil
}

func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	log.Infof("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "list":
		return c.list()
	case "":
		if len(c.driver.config.TVs) > 0 {
			return c.list()
		}
		fallthrough
	case "new":
		return c.edit(TVConfig{})
	case "edit":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)
		config := c.driver.config.get(vals["tv"])

		if config == nil {
			return c.error(fmt.Sprintf("Could not find tv with id: %s", vals["tv"]))
		}

		return c.edit(*config)
	case "delete":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)

		err := c.driver.deleteTV(vals["tv"])

		if err != nil {
			return c.error(fmt.Sprintf("Failed to delete tv: %s", err))
		}

		return c.list()
	case "save":
		var cfg TVConfig
		err := json.Unmarshal(request.Data, &cfg)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		err = c.driver.saveTV(cfg)

		if err != nil {
			return c.error(fmt.Sprintf("Could not save tv: %s", err))
		}

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

	var tvs []suit.ActionListOption

	for _, tv := range c.driver.config.TVs {
		tvs = append(tvs, suit.ActionListOption{
			Title: tv.Name,
			//Subtitle: tv.ID,
			Value: tv.ID,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "Samsung TVs",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "tv",
						Options: tvs,
						PrimaryAction: &suit.ReplyAction{
							Name:        "edit",
							DisplayIcon: "pencil",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "delete",
							Label:        "Delete",
							DisplayIcon:  "trash",
							DisplayClass: "danger",
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
				Label:        "New TV",
				Name:         "new",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

func (c *configService) edit(config TVConfig) (*suit.ConfigurationScreen, error) {

	title := "New Samsung TV"
	if config.ID != "" {
		title = "Editing Samsung TV"
	}

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.InputHidden{
						Name:  "id",
						Value: config.ID,
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "My TV",
						Value:       config.Name,
					},
					suit.InputText{
						Name:        "host",
						Before:      "Host",
						Placeholder: "IP or Hostname",
						Value:       config.Host,
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

func i(i int) *int {
	return &i
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
*/