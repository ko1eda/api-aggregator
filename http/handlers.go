package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	koala "github.com/ko1eda/apiaggregator"
)

// Handler
func (s *Server) handleGetProviderInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		loc := chi.URLParam(r, "id")

		// zero value for our datatype
		var pi *koala.ProviderInfo
		switch loc {
		// grill
		case "1":
			val, err := s.XmlGrillProvider.GetProviderInfo()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("{\"error\":\"Trouble connecting to grill provider, please try again!\"}"))
				return
			}
			pi = val
		// eatary
		case "2":
			val, err := s.JsonEateryProvider.GetProviderInfo()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("{\"error\":\"Trouble connecting to eatery provider, please try again!\"}"))
				return
			}
			pi = val

		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\":\"No items found for this location!\"}"))
			return
		}

		w.WriteHeader(http.StatusOK)

		// encode the result
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(pi)
	}
}

// This gets the full menu and returns the json
// If it hits a location id we don't know it returns an error
// For more routes we could add a map that holds our ID's and
// the corresponding function and load that into a config at runtime with a flag
func (s *Server) handleGetFullMenu() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		loc := chi.URLParam(r, "id")

		var menu *koala.Menu
		switch loc {
		case "1":
			// this runs a goroutine  under the scenes usually I would
			// pull this aysnc functionality up to the handler
			// put to keep it short I have left behavior inside the provider
			val, err := s.XmlGrillProvider.GetFullMenu()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("{\"error\":\"Trouble connecting to grill provider, please try again!\"}"))
				return
			}
			menu = val
		case "2":
			// this runs a goroutine  under the scenes usually I would
			// pull this aysnc functionality up to the handler
			// put to keep it short I have left behavior inside the provider
			val, err := s.JsonEateryProvider.GetFullMenu()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("{\"error\":\"Trouble connecting to eatery provider, please try again!\"}"))
				return
			}
			menu = val

		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\":\"No items found for this location!\"}"))
			return
		}

		w.WriteHeader(http.StatusOK)

		// encode the result
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(menu)
	}
}
