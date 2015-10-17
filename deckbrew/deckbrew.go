package deckbrew

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func Search(q string) ([]Card, error) {
	v := url.Values{"name": {q}}
	r, err := http.Get("http://api.deckbrew.com/mtg/cards?" + v.Encode())
	if err != nil {
		return nil, err
	}
	var resp []Card
	err = json.NewDecoder(r.Body).Decode(&resp)
	r.Body.Close()
	if err != nil {
		return nil, err
	}
	// Do exact match.
	if len(resp) > 1 {
		q = strings.ToLower(q)
		for _, c := range resp {
			if strings.ToLower(c.Name) == q {
				return []Card{c}, nil
			}
		}
	}
	return resp, nil
}

type Card struct {
	Name      string
	Types     []string
	Power     string
	Toughness string
	Cost      string
	Text      string
	Editions  []Edition
}

type Edition struct {
	Set       string
	Image_URL string
}

func (c Card) String() string {
	s := fmt.Sprintf("%v %v", c.Name, c.Cost)
	for _, t := range c.Types {
		if t == "creature" {
			s += fmt.Sprintf(" (%v/%v)", c.Power, c.Toughness)
		}
	}
	if c.Text != "" {
		s += "\n" + c.Text
	}
	if e := c.Editions; len(e) > 0 {
		s += "\n" + e[0].Image_URL
	}
	return s
}
