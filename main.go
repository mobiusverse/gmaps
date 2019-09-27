package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Placement struct {
	City      string  `json:"city"`
	State     string  `json:"state"`
	Country   string  `json:"country"`
	Code      string  `json:"code"`
	Accuracy  float64 `json:"accuracy"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	Zone      string  `json:"zone"`
	Url       string  `json:"url"`
	Desc      string  `json:"desc"`
}

var (
	placement = flag.String("placement", "", "search placement")
	longitude = flag.Float64("longitude", 0, "longitude")
	latitude  = flag.Float64("latitude", 0, "latitude")
	uri1      = "https://www.google.com/search?tbm=map&authuser=0&hl=en&gl=en&pb=!1b1&q=%s"
	uri2      = "https://www.google.com/maps/preview/reveal?authuser=0&hl=en&gl=en&pb=!3m2!2d%f!3d%f"
	client    = &http.Client{}
)

func main() {
	flag.Parse()
	uri := ""
	if "" != *placement {
		values, _ := url.ParseQuery(*placement)
		uri = fmt.Sprintf(uri1, values.Encode())
	} else if 0 != *longitude && 0 != *latitude {
		uri = fmt.Sprintf(uri2, *longitude, *latitude)
	}
	if "" == uri {
		log.Println("empty input")
		return
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println("unknown err:", err)
		}
	}()
	response, err := client.Get(uri)
	if err == nil && response != nil && response.StatusCode == 200 {
		bytes, err := ioutil.ReadAll(response.Body)
		if err == nil && len(bytes) > 4 {
			s := string(bytes)
			if s[:4] == ")]}'" {
				s = s[4:]
			}
			var tbm []interface{}
			if err := json.Unmarshal([]byte(s), &tbm); err == nil {
				var pl *Placement
				if strings.Contains(uri, "search") {
					pl = parse(tbm[0].([]interface{})[1].([]interface{})[0].([]interface{})[14].([]interface{}))
					pl.Accuracy = tbm[1].([]interface{})[0].([]interface{})[0].(float64)
				} else {
					pl = parse(tbm[2].([]interface{}))
				}
				marshal, _ := json.MarshalIndent(pl, "", "\t")
				println(string(marshal))
			} else {
				log.Println("unmarshal err:", s)
			}
		}
	} else {
		log.Println("response wrong:", response, err)
	}
}

func parse(item []interface{}) *Placement {
	pl := new(Placement)
	if nil != item && len(item) > 44 {
		if nil != item[9] {
			pl.Latitude = item[9].([]interface{})[2].(float64)
			pl.Longitude = item[9].([]interface{})[3].(float64)
		}
		if nil != item[18] {
			pl.Address = item[18].(string)
			split := strings.Split(pl.Address, ",")
			for i, j := 0, len(split)-1; i < j; i, j = i+1, j-1 {
				split[i], split[j] = split[j], split[i]
			}
			address := make([]string, 0)
			for _, val := range split {
				if !regexp.MustCompile("\\d+-\\d+").MatchString(val) {
					address = append(address, val)
				} else {
					pl.Code = val
				}
			}
			for i, val := range address {
				if i == 0 {
					pl.Country = val
				} else if i == 1 {
					pl.State = val
				} else if i == 2 {
					pl.City = val
				}
			}
		}
		if nil != item[30] {
			pl.Zone = item[30].(string)
		}
		if nil != item[42] {
			pl.Url = item[42].(string)
		}
		if nil != item[44] {
			pl.Desc = item[44].([]interface{})[2].([]interface{})[0].([]interface{})[0].(string)
			pl.Desc = strings.ReplaceAll(strings.ReplaceAll(pl.Desc, "\r", " "), "\n", " ")
		}
	}
	return pl
}
