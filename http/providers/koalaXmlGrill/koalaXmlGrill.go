package koalaXmlGrill

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	koala "github.com/ko1eda/apiaggregator"
	"github.com/ko1eda/apiaggregator/http"
)

// we do a runtime check to ensure our item implenets our AsyncProviderService
var _ http.AsyncProvider = &KoalaXmlGrill{}

// Relative to root of project
var (
	menuUrl = "./goldenfiles/xml-grill-data.xml"
)

// Represents our JsonEatery data provider
type KoalaXmlGrill struct {
	LocationID string
	client     http.HttpGetter
}

// Return a new jsonEatery with sensible defaults and variadic modifier params
func NewProvider(c http.HttpGetter, opts ...func(*KoalaXmlGrill)) *KoalaXmlGrill {
	k := &KoalaXmlGrill{client: c, LocationID: "1"}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

// This function can be passed to our generator function will add an id to our Eatery
// we can pass in a location ID dynamically instead of hardcoding
// we could create a config file to configure this at build time
// should our location api change
func WithLocationID(ID string) func(*KoalaXmlGrill) {
	return func(k *KoalaXmlGrill) {
		k.LocationID = ID
	}
}

// Parse our xml attributes
// Unlike json we can parse nested elements with > tag modifier
// We don't habve to make this struct as deeply nested to pull the data we want
type xmlLocationParser struct {
	ID             string   `xml:"id,attr"`
	Name           string   `xml:"name,attr"`
	StreetAddress  string   `xml:"streetaddress,attr"`
	City           string   `xml:"city,attr"`
	State          string   `xml:"state,attr"`
	Country        string   `xml:"country,attr"`
	Zip            string   `xml:"zip,attr"`
	Telephone      string   `xml:"telephone,attr"`
	Longitude      float64  `xml:"longitude,attr"`
	Latitude       float64  `xml:"latitude,attr"`
	PaymentMethods []string `xml:"billingdetails>billingmethods>billingmethod"`
	StoreHours     []struct {
		DayOfWeek string `xml:"day,attr"`
		Opens     string `xml:"from,attr"`
		Closes    string `xml:"to,attr"`
		Type      string `xml:"type,attr"`
	} `xml:"hours>period"`
}

// Get the provider info
func (k *KoalaXmlGrill) GetProviderInfo() (*koala.ProviderInfo, error) {
	f, err := os.Open(menuUrl)

	if err != nil {
		return nil, fmt.Errorf("LocationFetchErr: Could not get location %+v", err)
	}

	defer f.Close()

	p := &xmlLocationParser{}
	if err := xml.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("XmlDecodeErr: Could not decode menu %+v", err)
	}

	provider := createProviderInfo(p)

	return provider, nil
}

func createProviderInfo(p *xmlLocationParser) *koala.ProviderInfo {
	// we normalize the data here
	hrs := []*koala.ProviderHour{}
	for _, hr := range p.StoreHours {
		// make this look more like our json grill
		hr.Opens += ":00"
		hr.Closes += ":00"
		hr.DayOfWeek = strings.ToUpper(hr.DayOfWeek)[:3]

		// convert our hours struct into the proper format
		v := &koala.ProviderHour{
			DayOfWeek: hr.DayOfWeek,
			Closes:    hr.Closes,
			Opens:     hr.Opens,
			Type:      hr.Type,
		}

		hrs = append(hrs, v)
	}

	return &koala.ProviderInfo{
		ID:             p.ID,
		Name:           p.Name,
		StreetAddress:  p.StreetAddress,
		City:           strings.ToUpper(p.City),
		State:          p.State,
		Country:        p.Country,
		Zip:            p.Zip,
		Telephone:      p.Telephone,
		Longitude:      p.Longitude,
		Latitude:       p.Latitude,
		StoreHours:     hrs,
		PaymentMethods: p.PaymentMethods,
	}
}

// Im certain I could find a more efficient way to parse this
// But for times sake this will have to do.
type xmlMenuParser struct {
	Menu []struct {
		Category []struct {
			Name     string `xml:"name,attr"`
			ID       string `xml:"id,attr"`
			ItemData struct {
				Name             string `xml:"name,attr"`
				Description      string `xml:"description,attr"`
				ModifierListData []struct {
					Description string `xml:"description,attr"`
					Modifiers   []struct {
						ID           string  `xml:"id,attr"`
						Name         string  `xml:"name,attr"`
						Cost         float64 `xml:"cost,attr"`
						ModifierData []struct {
							ID              string `xml:"id,attr"`
							ModifierOptions []struct {
								Name string  `xml:"name,attr" json:"name"`
								Cost float64 `xml:"cost,attr" json:"cost"`
							} `xml:"options>option"`
						} `xml:"modifiers>optiongroup"`
					} `xml:"options>option"`
				} `xml:"modifiers>optiongroup"`
			} `xml:"products>product"`
		} `xml:"categories>category"`
	} `xml:"menu"`
}

// Get full menu, this runs some go routines, and uses buffered channel of 1 to return the data from the endpoints
// ayncronously,
func (k *KoalaXmlGrill) GetFullMenu() (*koala.Menu, error) {
	// resp, err := k.client.Get(menuUrl)
	menuChan := make(chan *koala.Menu, 1)
	providerChan := make(chan *koala.ProviderInfo, 1)

	f, err := os.Open(menuUrl)

	if err != nil {
		return nil, fmt.Errorf("MenuFetchErr: Could not fetch menu: %+v", err)
	}

	p2 := &xmlMenuParser{}
	if err := xml.NewDecoder(f).Decode(&p2); err != nil {
		return nil, fmt.Errorf("XmlDecodeErr: Could not decode menu %+v", err)
	}

	// run a go routine to parse the menu
	// close the channel when done
	go func(p2 *xmlMenuParser) {
		menuChan <- createMenu(p2)
		close(menuChan)
	}(p2)

	f.Close()

	f, err = os.Open(menuUrl)
	// make another call though
	// we can probably achieve this with a noop closer
	// so we don't have to make two calls
	// resp, err = k.client.Get(menuUrl)
	if err != nil {
		return nil, fmt.Errorf("LocationFetchErr: Could not fetch location: %+v", err)
	}

	defer f.Close()
	p := &xmlLocationParser{}
	if err := xml.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("XmlDecodeErr: Could not decode menu %+v", err)
	}

	go func(p *xmlLocationParser) {
		providerChan <- createProviderInfo(p)
		close(providerChan)
	}(p)

	menu := <-menuChan

	providerInfo := <-providerChan

	menu.ProviderInfo = providerInfo

	return menu, nil
}

// Create a menu, Improvements would be parsing this without a nested loop
func createMenu(p *xmlMenuParser) *koala.Menu {
	items := []*koala.MenuItem{}
	mods := []*koala.Modifier{}
	// deeply nested for loop
	// refactor
	for _, categories := range p.Menu {
		for _, item := range categories.Category {
			mi := &koala.MenuItem{}
			cat := &koala.Category{}
			cat.Name = item.Name
			cat.ID = item.ID
			mi.Name = item.ItemData.Name
			mi.Description = item.ItemData.Description
			mi.Category = cat
			for _, mod := range item.ItemData.ModifierListData {
				for _, f := range mod.Modifiers {
					m := &koala.Modifier{}
					m.ID = f.ID
					m.Cost = f.Cost
					m.Name = f.Name
					for _, omg := range f.ModifierData {
						m.Options = omg.ModifierOptions
						mods = append(mods, m)
						mi.Modifiers = mods
						break
					}
				}
			}
			items = append(items, mi)
		}
	}

	return &koala.Menu{MenuItems: items}
}
