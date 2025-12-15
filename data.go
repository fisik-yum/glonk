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
	"encoding/json"
	"os"
)

type owner struct {
	BotToken      string            `json:"bot_token"`
	LLMToken      string            `json:"llm_token"`
	Prompts       map[string]string `json:"prompts"`
	Profile string            `json:"default_profile"`
}

func read_config() { // main config file for end user
	f, err := os.ReadFile("config.json")
	check(err)
	var userData owner
	err = json.Unmarshal([]byte(f), &userData)
	check(err)
	bot_token = userData.BotToken
	llm_token = userData.LLMToken
	if bot_token == "" || llm_token == "" {
		panic("A token is missing. Check config.json")
	}
	prompts = userData.Prompts
	if prompts == nil {
		panic("Prompt map can't be empty!")
	}
	profile=userData.Profile
	if profile== ""{
		panic("Default prompt can't be empty!")
	}
}

// use the *current* global `prompt_string` as a template base
func generateFullPrompt(msg string) string {
	return prompts[profile] + msg
}

func check(e error) {
	if e != nil {
		panic(e.Error())
	}
}
