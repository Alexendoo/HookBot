package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Alexendoo/Slidebot/android"
	"github.com/Alexendoo/Slidebot/config"
	"github.com/Alexendoo/Slidebot/github"
	"github.com/Alexendoo/Slidebot/lastfm"
	"github.com/Alexendoo/Slidebot/store"
	"github.com/bwmarrin/discordgo"
)

func main() {
	err := config.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = store.Open("bolt.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	dg, err := discordgo.New("Bot " + config.Tokens.Discord)
	if err != nil {
		fmt.Println(err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dg.Close()

	gh := &github.Handler{
		Discord: dg,
	}

	mux := http.NewServeMux()
	mux.Handle("/hook/github", gh)

	srv := http.Server{
		Addr:    "localhost:9000",
		Handler: mux,
	}
	go srv.ListenAndServe()
	defer srv.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Stopping...")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("%s (%s): %s\n", m.Author.Username, m.Author.ID, m.Content)

	if m.Author.ID == s.State.User.ID {
		return
	}

	words := strings.Fields(m.Content)
	if len(words) == 0 || words[0][0] != '.' {
		return
	}

	switch words[0][1:] {
	case "help":
		s.ChannelMessageSend(m.ChannelID, "https://git.io/Slidebot")
	case "lastfm", "last.fm", "last", "l":
		lastfm.RecentTrack(words[1:], s, m.Message)
	case "api", "android", "sdk":
		android.APILevel(words[1:], s, m.Message)
	case "echo":
		s.ChannelMessageSend(m.ChannelID, m.Content)
	}
}
