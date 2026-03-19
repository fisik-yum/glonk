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

type Config struct {
	BotToken   string `json:"bot_token"`
	LLMToken   string `json:"llm_token"`
	Prompt     string `json:"prompt"`
	GlonkModel string `json:"glonk_model"`
}

func read_config() *Config { // main config file for end user
	f, err := os.ReadFile("config.json")
	check(err)
	var cfg Config
	err = json.Unmarshal([]byte(f), &cfg)
	check(err)
	if cfg.BotToken == "" || cfg.LLMToken == "" {
		panic("A token is missing. Check config.json")
	}
	if cfg.Prompt == "" {
		panic("Prompt map can't be empty!")
	}
	if cfg.GlonkModel == "" {
		panic("Invalid Model")
	}
	return &cfg
}

// use the *current* global `prompt_string` as a template base
func generateFullPrompt(msg string) string {
	return msg
}

func check(e error) {
	if e != nil {
		panic(e.Error())
	}
}
