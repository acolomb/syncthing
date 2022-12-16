// Copyright (C) 2022 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

type stat struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Total      int    `json:"total"`
	Translated int    `json:"translated"`
	Fuzzy      int    `json:"fuzzy"`
}

type translation map[string]string

func main() {
	log.SetFlags(log.Lshortfile)

	if t := authToken(); t == "" {
		log.Fatal("Need environment variable WEBLATE_TOKEN")
	}

	curValidLangs := map[string]bool{}
	for _, lang := range loadValidLangs() {
		curValidLangs[lang] = true
	}
	log.Println(curValidLangs)

	resp := req("https://hosted.weblate.org/exports/stats/syncthing/gui/?format=json")

	var stats []stat
	err := json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	names := make(map[string]string)

	var langs []string
	for _, stat := range stats {
		code := strings.Replace(stat.Code, "_", "-", 1)
		pct := 100 * stat.Translated / stat.Total
		if pct < 75 || !curValidLangs[code] && pct < 95 {
			log.Printf("Skipping language %q (too low completion ratio %d%%)", code, pct)
			os.Remove("lang-" + code + ".json")
			continue
		}

		langs = append(langs, code)
		names[code] = stat.Name
		if code == "en" {
			continue
		}

		log.Printf("Updating language %q", code)

		resp := req("https://hosted.weblate.org/api/translations/syncthing/gui/" + stat.Code + "/file/")
		bs, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		var t translation
		if err := json.Unmarshal(bs, &t); err != nil {
			log.Fatal(err)
		}

		fd, err := os.Create("lang-" + code + ".json")
		if err != nil {
			log.Fatal(err)
		}
		fd.Write(bs)
		fd.Close()
	}

	saveValidLangs(langs)
	saveLanguageNames(names)
}

func saveValidLangs(langs []string) {
	sort.Strings(langs)
	fd, err := os.Create("valid-langs.js")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(fd, "var validLangs = ")
	json.NewEncoder(fd).Encode(langs)
	fd.Close()
}

func saveLanguageNames(names map[string]string) {
	fd, err := os.Create("prettyprint.js")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(fd, "var langPrettyprint = ")
	json.NewEncoder(fd).Encode(names)
	fd.Close()
}

func authToken() string {
	token := os.Getenv("WEBLATE_TOKEN")
	return token
}

func req(url string) *http.Response {
	token := authToken()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func loadValidLangs() []string {
	fd, err := os.Open("valid-langs.js")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()
	bs, err := io.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}

	var langs []string
	exp := regexp.MustCompile(`\[([a-zA-Z@",-_]+)\]`)
	if matches := exp.FindSubmatch(bs); len(matches) == 2 {
		langs = strings.Split(string(matches[1]), ",")
		for i := range langs {
			// Remove quotes
			langs[i] = langs[i][1 : len(langs[i])-1]
		}
	}

	return langs
}
