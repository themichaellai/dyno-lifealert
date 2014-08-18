package main

import (
	"encoding/json"
	"fmt"
	"github.com/bgentry/heroku-go"
	"github.com/garyburd/redigo/redis"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const redisLastRestartedKey string = "herokurestarter:lastrestarted"

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

func saveTimestampRedis(redisClient redis.Conn) {
	_, err := redisClient.Do("SET", redisLastRestartedKey, time.Now().Unix())
	check(err)
}

func getTimestampRedis(redisClient redis.Conn) time.Time {
	seconds, err := redis.Int64(redisClient.Do("GET", redisLastRestartedKey))
	check(err)
	return time.Unix(seconds, 0)
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

	redisClient, err := redis.Dial("tcp", ":6379")
	check(err)

	args := os.Args[1:]
	if len(args) > 0 {
		getApps(heroku)
		return
	}
	m := martini.Classic()
	m.Use(auth.BasicFunc(func(_, password string) bool {
		return auth.SecureCompare(password, settings["password"].(string))
	}))

	m.Get("/", func() string {
		var time = getTimestampRedis(redisClient)
		return fmt.Sprintf("help I've crashed and can't get up %s",
			strconv.FormatInt(time.Unix(), 10))
	})

	m.Post("/restart", func() string {
		if err := heroku.DynoRestartAll(settings["dyno_id"].(string)); err != nil {
			panic(err)
		}
		go saveTimestampRedis(redisClient)
		return "success"
	})

	m.Run()
}
