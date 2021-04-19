package koala

// These are the root data types for our application
type Menu struct {
	ProviderInfo *ProviderInfo `json:"provider_info,omitempty"`
	MenuItems    []*MenuItem   `json:"menu_items,omitempty"`
}

// A menu Item Represents a single menu aggregated menu item
// that we will return to the user
type MenuItem struct {
	ID             string      `json:"id,omitempty"`
	Name           string      `json:"name,omitempty"`
	Description    string      `json:"description,omitempty"`
	Disabled       bool        `json:"disabled,omitempty"`
	CategoryID     string      `json:"category_id,omitempty"`
	Category       *Category   `json:"category,omitempty"`
	ModifierListID string      `json:"-"`
	Modifiers      []*Modifier `json:"modifiers,omitempty"`
}
type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

// Modifier represents a modification to the original item
// Add two xml tags to make the decoding easier for times sake would
// try and keep this implementation out of final code on a real project.
type Modifier struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Disabled bool    `json:"disabled"`
	Cost     float64 `json:"cost,omitempty"`
	Options  []struct {
		Name string  `xml:"name,attr" json:"name"`
		Cost float64 `xml:"cost,attr" json:"cost"`
	} `json:"options"`
}

// This is our store info
type ProviderInfo struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	StreetAddress  string          `json:"street_address,omitempty"`
	City           string          `json:"city"`
	State          string          `json:"state"`
	Country        string          `json:"country"`
	Zip            string          `json:"zip"`
	Telephone      string          `json:"telephone"`
	Longitude      float64         `json:"longitude"`
	Latitude       float64         `json:"latitude"`
	StoreHours     []*ProviderHour `json:"store_hours"`
	PaymentMethods []string        `json:"payment_methods"`
}

// Coupled with Provider Info
type ProviderHour struct {
	Type      string `json:"type"`
	DayOfWeek string `json:"day_of_week"`
	Opens     string `json:"opens"`
	Closes    string `json:"closes"`
}
