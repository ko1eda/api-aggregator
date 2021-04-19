package koalaJsonEatery

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	koala "github.com/ko1eda/apiaggregator"
	"github.com/ko1eda/apiaggregator/http"
)

// we do a runtime check to ensure our item implenets our AsyncProviderService
var _ http.AsyncProvider = &KoalaJsonEatery{}

// We could add a config and pass our endpoints and add them to a map on our provider
// We are using files for now
var (
	menuUrl     = "./goldenfiles/json-eatery-menu.json"
	locationUrl = "./goldenfiles/json-eatery-location.json"
)

// Represents our JsonEatery data provider
type KoalaJsonEatery struct {
	LocationID string
	client     http.HttpGetter
}

// Return a new jsonEatery with sensible defaults and variadic modifier params
func NewProvider(c http.HttpGetter, opts ...func(*KoalaJsonEatery)) *KoalaJsonEatery {
	k := &KoalaJsonEatery{client: c, LocationID: "2"}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

// This function can be passed to our generator function will add an id to our Eatery
// We can pass in a location ID dynamically instead of hardcoding
// we could create a config file to configure this at build time
// should our location api change
func WithLocationID(ID string) func(*KoalaJsonEatery) {
	return func(k *KoalaJsonEatery) {
		k.LocationID = ID
	}
}

// This Stucture we used to parse the nested location info
type koalaJsonEateryLocationParser struct {
	Locations []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Address struct {
			City    string `json:"locality"`
			Zip     string `json:"postal_code"`
			Country string `json:"country"`
			State   string `json:"administrative_district_level_1"`
		} `json:"address"`
		PaymentMethods []string `json:"capabilities"`
		Telephone      string   `json:"phone_number"`
		StoreHours     struct {
			Periods []struct {
				DayOfWeek string `json:"day_of_week"`
				Opens     string `json:"start_local_time"`
				Closes    string `json:"end_local_time"`
			} `json:"periods"`
		} `json:"business_hours"`
		Coordinates struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
	} `json:"locations"`
}

// Hits our location endpoint create the necessary config with a helper function
// and return the provider info
func (k *KoalaJsonEatery) GetProviderInfo() (*koala.ProviderInfo, error) {
	// resp, err := k.client.Get(locationUrl)
	f, err := os.Open(locationUrl)

	if err != nil {
		return nil, fmt.Errorf("LocationFetchErr: Could not get location %+v", err)
	}

	defer f.Close()

	p := &koalaJsonEateryLocationParser{}

	if err := json.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("JsonDecodeErr: Could not decode menu %+v", err)
	}

	provider := createProviderInfo(p, k.LocationID)

	return provider, nil
}

// This function is unexported, we can use this function to test that our data is being transformed
// correctly without having to necessarily worry about the implementation of our GetProviderInfo function
// Given more time I would probalby add more abstraction, perhaps adding the ability for each provider to parse its own
// data and then return an interface to make this even easier to work with
func createProviderInfo(p *koalaJsonEateryLocationParser, ID string) *koala.ProviderInfo {
	info := &koala.ProviderInfo{}
loop:
	for _, p := range p.Locations {
		if p.ID == ID {
			info.ID = p.ID
			info.Name = p.Name
			info.City = p.Address.City
			info.State = p.Address.State
			info.Country = p.Address.Country
			info.Zip = p.Address.Zip
			info.Telephone = p.Telephone
			info.Name = p.Name
			info.Longitude = p.Coordinates.Longitude
			info.Latitude = p.Coordinates.Latitude
			for _, hrs := range p.StoreHours.Periods {
				hour := &koala.ProviderHour{}
				hour.DayOfWeek = hrs.DayOfWeek
				hour.Opens = hrs.Opens
				hour.Closes = hrs.Closes
				info.StoreHours = append(info.StoreHours, hour)
			}
			// Normalize the payment method field to more closely match the output of our xml grill
			for _, method := range p.PaymentMethods {
				info.PaymentMethods = append(info.PaymentMethods, strings.ReplaceAll(method, "_", " "))
			}
			break loop
		}
	}

	return info
}

// We use this to unmarshal the incoming json for a menu
type koalaJsonEateryMenuParser struct {
	Objects []struct {
		Type         string `json:"type"`
		ID           string `json:"id"`
		Deleted      bool   `json:"is_deleted"`
		CategoryData struct {
			Name string `json:"name"`
		} `json:"category_data"`
		ModifierListData struct {
			Name      string `json:"name"`
			Modifiers []struct {
				ID           string `json:"id"`
				ModifierData struct {
					Name      string `json:"name"`
					PriceData struct {
						Cost     float64 `json:"amount"`
						Currency string  `json:"currency"`
					} `json:"price_money"`
				} `json:"modifier_data"`
			} `json:"modifiers"`
		} `json:"modifier_list_data"`
		ItemData struct {
			Name             string `json:"name"`
			Description      string `json:"description"`
			Visibility       string `json:"visibility"`
			CategoryID       string `json:"category_id"`
			ModifierListInfo []struct {
				ModifierListID string `json:"modifier_list_id"`
			} `json:"modifier_list_info"`
			Variations []struct {
				Type      string `json:"type"`
				SKU       string `json:"sku"`
				PriceData struct {
					Cost     float64 `json:"amount"`
					Currency string  `json:"currency"`
				} `json:"price_money"`
			}
		} `json:"item_data"`
	}
}

// TODO: Move these goroutines up to the handler so that we remove any of the magic
// thats going on behind the seens, these functions are async and could block for the duration of the transport timeout
// so putting the async calls at a higher level might be benficial for future developers
func (k *KoalaJsonEatery) GetFullMenu() (*koala.Menu, error) {
	menuChan := make(chan *koala.Menu, 1)
	providerChan := make(chan *koala.ProviderInfo, 1)

	f, err := os.Open(menuUrl)

	if err != nil {
		return nil, fmt.Errorf("MenuFetchErr: Could not fetch menu: %+v", err)
	}

	p := &koalaJsonEateryMenuParser{}
	if err := json.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("JsonDecodeErr: Could not decode menu %+v", err)
	}
	// run a go routine to parse the menu
	// close the channel when done
	go func(p *koalaJsonEateryMenuParser) {
		menuChan <- createMenu(p)
		close(menuChan)
	}(p)

	// response 2
	f, err = os.Open(locationUrl)
	if err != nil {
		return nil, fmt.Errorf("LocationFetchErr: Could not fetch location: %+v", err)
	}

	// Parse
	p2 := &koalaJsonEateryLocationParser{}
	if err := json.NewDecoder(f).Decode(&p2); err != nil {
		return nil, fmt.Errorf("JsonDecodeErr: Could not decode menu %+v", err)
	}

	// Run this in another goroutine and then aggregate the data below
	go func(p2 *koalaJsonEateryLocationParser) {
		providerChan <- createProviderInfo(p2, k.LocationID)
		close(providerChan)
	}(p2)

	f.Close()

	// read from our channels getting the returned data
	menu := <-menuChan

	providerInfo := <-providerChan

	menu.ProviderInfo = providerInfo

	// return combined full menu with the provider info
	return menu, nil
}

// Get the data in a common format that we can share between our implementations
func createMenu(p *koalaJsonEateryMenuParser) *koala.Menu {
	cats := make(map[string]koala.Category)
	items := []*koala.MenuItem{}
	modifiers := make(map[string]koala.Modifier)
	for _, value := range p.Objects {
		mi := &koala.MenuItem{}
		mod := &koala.Modifier{}
		cat := &koala.Category{}
		// switch on our top level items
		// asinging the values we want to our transformed menu item
		switch value.Type {
		case "ITEM":
			mi.ID = value.ID
			mi.Name = value.ItemData.Name
			mi.Description = value.ItemData.Description
			mi.CategoryID = value.ItemData.CategoryID
			mi.Disabled = value.Deleted
			// if we have a modifier list attached to this
			// the we need to set its list id so we can map it later
			for _, v := range value.ItemData.ModifierListInfo {
				mi.ModifierListID = v.ModifierListID
			}
		case "CATEGORY":
			cat.ID = value.ID
			cat.Name = value.CategoryData.Name
			cats[value.ID] = *cat
		case "MODIFIER_LIST":
			mod.ID = value.ID
			mod.Disabled = value.Deleted
			mod.Name = value.ModifierListData.Name
			for _, v := range value.ModifierListData.Modifiers {
				mod.Options = append(
					mod.Options,
					struct {
						Name string  `xml:"name,attr" json:"name"`
						Cost float64 `xml:"cost,attr" json:"cost"`
					}{
						Name: v.ModifierData.Name,
						Cost: v.ModifierData.PriceData.Cost,
					},
				)
			}
			// we want to map our modifier list id
			// to our modifier and options
			// then we can associate it
			// with menu item below
			modifiers[value.ID] = *mod
		}

		// o(1) lookup for our categories
		// lets map them to our menuItem
		if c, hit := cats[mi.CategoryID]; hit {
			mi.Category = &c
		}
		// same as above if we have a matching modifier
		// then we want to attach it to our menu item.
		// ignoring varations for sake of brevity
		if mod, hit := modifiers[mi.ModifierListID]; hit {
			mi.Modifiers = append(mi.Modifiers, &mod)
		}
		// if our menu item has a zero value
		// for its id we don't store it (we probably hit another case and menu item is a 0 value struct)
		if mi.ID != "" {
			items = append(items, mi)
		}
	}
	// return our menu
	return &koala.Menu{MenuItems: items}
}
