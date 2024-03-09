package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type episode struct {
	Season  int    `json:"season"`
	Episode int    `json:"episode"`
	Name    string `json:"name"`
	AirDate string `json:"air_date"`
	Seen    bool   `json:"seen"`
}

type tvshow struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Status      string    `json:"status"`
	Episodes    []episode `json:"episodes"`
}

type showjson struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type showresults struct {
	Shows []showjson `json:"tv_shows"`
}

const jsonpath = "json/"

func searchShows(showname string) showresults {
	r, e := http.Get(fmt.Sprintf("https://www.episodate.com/api/search?q=%s&page=1", showname))

	if e != nil {
		fmt.Println("no response")
	}

	defer r.Body.Close()

	body, e := io.ReadAll(r.Body)

	if e != nil {
		fmt.Println("no response")
	}

	var results showresults

	if e := json.Unmarshal(body, &results); e != nil {
		fmt.Println("Unmarshal failed")
	}

	return results

}

func downloadShow(showid int) error {
	r, e := http.Get(fmt.Sprintf("https://www.episodate.com/api/show-details?q=%d", showid))

	if e != nil {
		return e
	}

	defer r.Body.Close()

	body, e := io.ReadAll(r.Body)
	if e != nil {
		return e
	}

	var result struct {
		Tvshow tvshow `json:"tvShow"`
	}

	if e := json.Unmarshal(body, &result); e != nil {
		fmt.Println("Unmarshal failed")
	}

	f, e := os.Create(jsonpath + strconv.Itoa(showid))

	if e != nil {
		return e
	}

	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	if e := encoder.Encode(result.Tvshow); e != nil {
		return e
	}

	sh, _ := readShow(strconv.Itoa(showid))

	filterlist := []string{"<b>", "</b>"}

	for _, i := range filterlist {
		sh.Description = strings.ReplaceAll(sh.Description, i, "")
	}

	for _, i := range sh.Episodes {
		i.Seen = false
	}

	writeShow(sh, strconv.Itoa(showid))

	return nil

}

func listShows() []showjson {
	files, _ := os.ReadDir(jsonpath)

	var shows []showjson
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		t, _ := readShow(file.Name())
		shows = append(shows, showjson{t.ID, t.Name})
	}

	return shows
}

func readShow(filename string) (tvshow, error) {
	var show tvshow

	file, e := os.Open(jsonpath + filename)
	if e != nil {
		return show, e
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if e := decoder.Decode(&show); e != nil {
		return show, e
	}

	return show, nil
}

func writeShow(show tvshow, filename string) error {
	file, e := os.Create(jsonpath + filename)
	if e != nil {
		return e
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if e := encoder.Encode(show); e != nil {
		return e
	}

	return nil
}
