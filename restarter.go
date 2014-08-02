package main

import (
	"encoding/json"
	"fmt"
	"github.com/bgentry/heroku-go"
	"io/ioutil"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readSettings() map[string]interface{} {
	dat, err := ioutil.ReadFile("settings.json")
	check(err)
	var settingsJson map[string]interface{}
	if err := json.Unmarshal(dat, &settingsJson); err != nil {
		panic(err)
	}
	return settingsJson
}

func getApps(client heroku.Client) {
	if apps, err := client.AppList(nil); err != nil {
		fmt.Println("Failed to get app list,", err)
	} else {
		for _, app := range apps {
			fmt.Println(app.Name, app.Id)
		}
	}
}

func main() {
	settings := readSettings()
	heroku := heroku.Client{
		Username: settings["user"].(string), Password: settings["api_key"].(string)}
	args := os.Args[1:]
	if len(args) > 0 {
		getApps(heroku)
	} else {
		if err := heroku.DynoRestartAll(settings["dyno_id"].(string)); err != nil {
			panic(err)
		}
	}
}
