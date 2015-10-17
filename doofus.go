// Command doofus implements a Telegram bot that fetches Magic Card information.
package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/broady/mtgprice/mtgprice"
	"github.com/nf/doofus/deckbrew"
	"github.com/tucnak/telebot"
)

const (
	bot        = "DoofusBot"
	maxMatches = 15
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cards, err := mtgprice.Open(mtgprice.Opts{Filename: "mtgprice.kv", CardData: "AllCards.json"})
	if err != nil {
		log.Fatal("loading cards:", err)
	}
	closeOnTerm(cards)

	bot, err := telebot.NewBot(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	messages := make(chan telebot.Message)
	bot.Listen(messages, 1*time.Minute)

	for msg := range messages {
		if err := handleMessage(cards, bot, msg); err != nil {
			log.Printf("Error handling message %v: %v", msg, err)
		}
	}
}

func closeOnTerm(c io.Closer) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Printf("shutting down...")
		if err := c.Close(); err != nil {
			log.Fatalf("clean up error: %v", err)
		}
		os.Exit(1)
	}()
}

func handleMessage(cards *mtgprice.Client, bot *telebot.Bot, m telebot.Message) error {
	reply := func(s string) error { return bot.SendMessage(m.Chat, s, nil) }

	if m.Text == "/dobis" {
		return reply(dobis[rand.Intn(len(dobis))])
	}

	q, ok := isSearch(m.Text)
	if !ok {
		return nil
	}

	t0 := time.Now()
	result, err := cards.Query(q)
	if err != nil {
		return reply("Error: " + err.Error())
	}
	s := ""
	switch len(result) {
	case 0:
		s = fmt.Sprintf("I don't know what %q is.", q)
	case 1:
		ci := result[0]

		img := make(chan string)
		go func() {
			m, err := deckbrew.Search(ci.Name)
			if err != nil || len(m) == 0 || len(m[0].Editions) == 0 {
				img <- ""
				return
			}
			img <- "\n" + m[0].Editions[0].Image_URL
		}()

		c, err := cards.RichInfo(ci.Name)
		if err != nil {
			return reply("Error: " + err.Error())
		}

		s = fmt.Sprintf("%s\nTCG %v%s", c.Detail(), c.TCGPrice, <-img)
		fmt.Println(time.Since(t0))
	default:
		if len(result) > maxMatches {
			s = fmt.Sprintf("I know %v cards like %q. Be more specific.", len(result), q)
			break
		}
		s = fmt.Sprintf("I know a few cards like %q:", q)
		for _, c := range result {
			s += "\n  " + c.Name
		}
	}
	return reply(s)
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

var dobis = []string{
	"Just two boys doing business.",
	"We're Dobis P.R.",
	"That's who we are.",
	"Doing business.",
	`Uh, "Dear Mr. Weebs, We are very, very, very, very excited to meet you. We are Dobis." That's all I've got so far.`,
	"I'm Tim Heidecker.  This is Eric Wareheim.  We are Dobis P.R., and we're here to tell you about our plan to revitalize the S'wallow Valley Mall.",
	"That's the pride of Dobis.",
	"Dobis in the house.",
	"Official Dobis reps here, stopping in for a meet-and-greet.",
	"Taquito, we just had a Dobis meeting, and we've decided to let you run the mall fountain.",
	"I guess I'll just stay here and work on Dobis.",
	"Guys, tomorrow is our big day, and Dobis couldn't be more jazzed about it.  But right now I want to have a little fun, with the permission of my best friend Tim.\nMuyo permissiono granted.",
}
