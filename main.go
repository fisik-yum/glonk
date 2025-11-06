/*
   glonk - glonk
   Copyright (C) 2025  fisik_yum

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"google.golang.org/genai"
)

// variables found in config.json, which needs to exist
var (
	bot_token       string
	llm_token       string
	prompts         map[string]string
	profile         string
	config_location string
)
var glonk_model = "gemini-2.5-flash-lite-preview-09-2025"

var s *discordgo.Session
var c *genai.Client
var ctx context.Context

func init() {
	flag.StringVar(&config_location, "c", "", "path to configuration file")
	flag.Parse()
	_, err := os.Stat(config_location)
	if os.IsNotExist(err) {
		panic("config.json is missing")
	}
	read_config()

	s, err = discordgo.New("Bot " + bot_token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	c, err = genai.NewClient(ctx, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
		APIKey:  llm_token,
	})
	if err != nil {
		log.Fatalf("glonk auth error!")
	}
	ctx = context.Background()
}

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	s.ShouldReconnectOnError = true
	s.Identify.Intents = 12800

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		log.Printf("Registering command: %v", v.Name)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	log.Println("Adding glonk Handler")
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		IsMentioned := false

		for _, user := range m.Mentions {
			if user.ID == s.State.User.ID {
				IsMentioned = true
				break // Exit the loop once we've found the mention
			}
		}

		if IsMentioned {

			parts := []*genai.Part{
				genai.NewPartFromText(generateFullPrompt(m.Message.Content)),
			}

			for _, v := range m.Attachments {
				if strings.HasPrefix( v.ContentType,"image") {
					parts = append(parts, genai.NewPartFromBytes(getFile(v.URL), v.ContentType))
				}
			}

			log.Printf("Detected prompt from user %s: %s\n", m.Author.ID, m.Message.Content)
			result, err := c.Models.GenerateContent(
				ctx,
				glonk_model,
				[]*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)},
				&genai.GenerateContentConfig{
					MaxOutputTokens: 150,
				},
			)
			check(err)
			s.ChannelMessageSendReply(m.ChannelID, result.Text(), m.Reference())

		}
	})

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
	log.Println("Gracefully shutting down.")
}
