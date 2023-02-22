package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/go-redis/redis"
)

type Config struct {
	RequestsPerSecond int64 `yaml:"RequestsPerSecond"`
}

func getConf() (*Config, error) {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	c := &Config{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c, err
}

/**
 * Indica si se rechaza el request por haber superado la cantidad de requests permitidos por segundo (por IP)
 */
func isRateLimited(client *redis.Client, ad string, limite int64) bool {
	val, _ := client.Get(ad).Result()

	if val == "" {
		client.Set(ad, 1, time.Second*60).Result()
		return false
	}

	val_int, _ := strconv.ParseInt(val, 10, 64)

	if val_int >= limite {
		return true
	}

	client.Incr(ad)
	return false
}

func main() {
	c, err := getConf()

	fmt.Printf("Starting serverrr at port 8080\n")
	fmt.Printf("Config of %d\n", c.RequestsPerSecond)

	client := redis.NewClient(&redis.Options{
		Addr:     "redis-server:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if isRateLimited(client, r.RemoteAddr, c.RequestsPerSecond) {
			http.Error(w, "Rate limited", http.StatusTooManyRequests)
			return

		}
		fmt.Fprintf(w, "Hello %s!", r.RemoteAddr)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
