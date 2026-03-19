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

var genconfig = &genai.GenerateContentConfig{MaxOutputTokens: 150}

// variables found in config.json, which needs to exist
var (
	prompts         map[string]string
	config_location string
)

var s *discordgo.Session
var ctx context.Context
var client *genai.Client
var cfg *Config
var chats map[string]*genai.Chat

func init() {
	flag.StringVar(&config_location, "c", "", "path to configuration file")
	flag.Parse()
	_, err := os.Stat(config_location)
	if os.IsNotExist(err) {
		panic("config.json is missing")
	}
	cfg = read_config()
	s, err = discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	ctx = context.Background()
	client, err = genai.NewClient(ctx, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
		APIKey:  cfg.LLMToken,
	})
	if err != nil {
		log.Fatalf("glonk auth error!")
	}
	log.Println(cfg.Prompt)
	chats = make(map[string]*genai.Chat)
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
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
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
			log.Printf("Detected prompt from user %s: %s\n", m.Author.ID, m.Message.Content)
			parts := []genai.Part{
				*genai.NewPartFromText(generateFullPrompt(m.Message.Content)),
			}
			for _, v := range m.Attachments {
				if strings.HasPrefix(v.ContentType, "image") {
					parts = append(parts, *genai.NewPartFromBytes(getFile(v.URL), v.ContentType))
				}
			}
			guild := m.GuildID
			chanl, _ := s.Channel(m.ChannelID)
			if chanl.Type == discordgo.ChannelTypeDM || chanl.Type == discordgo.ChannelTypeGroupDM {
				guild = m.ChannelID + "_nonguild"
			}
			if _, ok := chats[guild]; !ok {
				log.Printf("Creating new chat for guild %s", guild)
				chats[guild], err = client.Chats.Create(ctx, cfg.GlonkModel, genconfig, nil)
				chats[guild].SendMessage(ctx,*genai.NewPartFromText(cfg.Prompt))
				check(err)
			}
			result, err := chats[guild].SendMessage(ctx, parts...)
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
