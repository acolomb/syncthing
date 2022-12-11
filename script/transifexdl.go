// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type statsResponse struct {
	Data []resourceLanguageStats
}

type resourceLanguageStats struct {
	ID   string `json:"id"`
	Attr stat   `json:"attributes"`
}

type stat struct {
	Translated   int `json:"translated_strings"`
	Untranslated int `json:"untranslated_strings"`
	Reviewed     int `json:"reviewed_strings"`
	Total        int `json:"total_strings"`
}

type translation struct {
	Content string
}

func main() {
	log.SetFlags(log.Lshortfile)

	if t := userPass(); t == "" {
		log.Fatal("Need environment variable TRANSIFEX_TOKEN")
	}

	curValidLangs := map[string]bool{}
	for _, lang := range loadValidLangs() {
		curValidLangs[lang] = true
	}
	log.Println(curValidLangs)

	resp := req("https://rest.api.transifex.com/resource_language_stats?filter[project]=o:syncthing:p:syncthing&filter[resource]=o:syncthing:p:syncthing:r:gui")

	var stats statsResponse
	err := json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	names := make(map[string]string)

	var langs []string
	for _, stat := range stats.Data {
		origCode := stat.ID[strings.LastIndex(stat.ID, ":")+1:]
		code := strings.Replace(origCode, "_", "-", 1)
		pct := 100 * stat.Attr.Translated / stat.Attr.Total
		if pct < 75 || !curValidLangs[code] && pct < 95 {
			log.Printf("Skipping language %q (too low completion ratio %d%%)", code, pct)
			os.Remove("lang-" + code + ".json")
			continue
		}

		langs = append(langs, code)
		names[code] = languageName(code)
		if code == "en" {
			continue
		}

		log.Printf("Updating language %q", code)

		resp, err = downloadTranslationFile(origCode, "default")
		if err != nil {
			log.Fatal(err)
		}
		var t translation
		err := json.NewDecoder(resp.Body).Decode(&t)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		fd, err := os.Create("lang-" + code + ".json")
		if err != nil {
			log.Fatal(err)
		}
		fd.WriteString(t.Content)
		fd.Close()
	}

	saveValidLangs(langs)
	saveLanguageNames(names)
}

type asyncDownloadRequest struct {
	Data struct {
		Attr struct {
			Mode     string `json:"mode"`
			FileType string `json:"file_type"`
		} `json:"attributes"`
		Relationships struct {
			Language struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"language"`
			Resource struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"resource"`
		} `json:"relationships"`
		Type string `json:"type"`
	} `json:"data"`
}

type asyncDownloadResponse struct {
	Data struct {
		Attr struct {
			Status string `json:"status"`
		} `json:"attributes"`
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"data"`
	Errors []map[string]string `json:"errors"`
}

func downloadTranslationFile(code, mode string) (*http.Response, error) {
	var r asyncDownloadRequest
	r.Data.Attr.Mode = mode
	r.Data.Attr.FileType = "json"
	r.Data.Relationships.Language.Data.ID = "l:" + code
	r.Data.Relationships.Language.Data.Type = "languages"
	r.Data.Relationships.Resource.Data.ID = "o:syncthing:p:syncthing:r:gui"
	r.Data.Relationships.Resource.Data.Type = "resources"
	r.Data.Type = "resource_translations_async_downloads"

	requestBody, _ := json.Marshal(r)

	resp := reqPost("https://rest.api.transifex.com/resource_translations_async_downloads", requestBody)
	location := resp.Header.Get("Content-Location")
	var a asyncDownloadResponse
checkAgain:
	err := json.NewDecoder(resp.Body).Decode(&a)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode == 202 {
		log.Println(" code responded", resp.StatusCode)
		log.Println("Status is", a.Data.Attr.Status, "location", location)
		switch a.Data.Attr.Status {
		case "succeeded":
			resp = req(location)
			return resp, nil

		case "pending", "processing":
			log.Println("Retrying in one second")
			time.Sleep(1 * time.Second)
			resp = req(location)
			if resp.StatusCode != 200 {
				goto checkAgain
			}

		default:
			return nil, errors.New("Failed response")
		}
	}
	return resp, nil
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

func userPass() string {
	token := os.Getenv("TRANSIFEX_TOKEN")
	return token
}

func req(url string) *http.Response {
	token := userPass()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("RESPONSE:\n%s", string(respDump))
	}

	return resp
}

func reqPost(url string, content []byte) *http.Response {
	token := userPass()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(content))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("REQUEST:\n%s", string(reqDump))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RESPONSE:\n%s", string(respDump))

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
	exp := regexp.MustCompile(`\[([a-zA-Z@",-]+)\]`)
	if matches := exp.FindSubmatch(bs); len(matches) == 2 {
		langs = strings.Split(string(matches[1]), ",")
		for i := range langs {
			// Remove quotes
			langs[i] = langs[i][1 : len(langs[i])-1]
		}
	}

	return langs
}

type languageResponse struct {
	Data struct {
		Code string
		Name string
	}
}

func languageName(code string) string {
	var lang languageResponse
	resp := req("https://rest.api.transifex.com/languages/l:" + code)
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(&lang)
	if err != nil {
		log.Fatal(err)
	}
	if lang.Data.Name == "" {
		return code
	}
	return lang.Data.Name
}
