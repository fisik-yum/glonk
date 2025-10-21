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
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "glonk_info",
		Description: "list glonk model",
	},
	{
		Name:        "glonk_profile",
		Type:        discordgo.ChatApplicationCommand,
		Description: "set glonk prompt",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "profile",
				Required:    true,
				Description: "change profile",
			},
		},
	},
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"glonk_info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		e := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("```\nglonk model: %s\nglonk profile: %s\n```", glonk_model, profile),
			},
		})
		check(e)
	},
	"glonk_profile": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		p := i.ApplicationCommandData().Options[0].StringValue()
		if _, ok := prompts[p]; ok {
			profile = p
			e := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Changed profile to `%s`", p),
				},
			})
			check(e)
		} else {
			e := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "An error occured!",
				},
			})
			check(e)
		}

	},
}
