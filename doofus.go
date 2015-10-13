// Command doofus implements a Telegram bot that fetches Magic Card information.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

const (
	bot        = "DoofusBot"
	maxMatches = 15
)

func main() {
	bot, err := telebot.NewBot(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	messages := make(chan telebot.Message)
	bot.Listen(messages, 1*time.Minute)

	for message := range messages {
		q, ok := isSearch(message.Text)
		if !ok {
			continue
		}
		cards, err := search(q)
		if err != nil {
			bot.SendMessage(message.Chat, "Error fetching: "+err.Error(), nil)
			continue
		}
		msg := ""
		switch len(cards) {
		case 0:
			msg = fmt.Sprintf("I don't know what %q is.", q)
		case 1:
			msg = cards[0].String()
		default:
			if len(cards) > maxMatches {
				msg = fmt.Sprintf("I know %v cards like %q. Be more specific.", len(cards), q)
				break
			}
			msg = fmt.Sprintf("I know a few cards like %q:", q)
			for _, c := range cards {
				msg += "\n  " + c.Name
			}
		}
		bot.SendMessage(message.Chat, msg, nil)
	}
}

func isSearch(t string) (q string, ok bool) {
	const a, b = "/card ", "@" + bot + " "
	switch {
	case strings.HasPrefix(t, a):
		return strings.TrimPrefix(t, a), true
	case strings.HasPrefix(t, b):
		return strings.TrimPrefix(t, b), true
	}
	return "", false
}

func search(q string) ([]Card, error) {
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

type Edition struct {
	Set       string
	Image_URL string
}
