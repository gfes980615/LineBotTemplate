// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	"github.com/line/line-bot-sdk-go/linebot"

	_ "github.com/joho/godotenv/autoload"
)

var (
	bot       *linebot.Client
	dianaHost string
)

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	dianaHost = os.Getenv("diana_host")
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		log.Print(err.Error())
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:

				if message.Text == "a" {
					daily := getEveryDaySentence()
					bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(daily)).Do()
					return
				}

				if message.Text == "plt" {
					bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(getDianaResponse(message.Text))).Do()
				}

				id, transferErr := strconv.ParseInt(message.Text, 10, 64)
				text := getGoogleExcelValueById(id)
				if transferErr != nil {
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(transferErr.Error())).Do(); err != nil {
						log.Print(err)
					}
					return
				}
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
	// fmt.Println(getEveryDaySentence())
}

func getDianaResponse(message string) string {
	url := "http://%s/diana/currency/value?server=%s"
	resp, err := http.Get(fmt.Sprintf(url, dianaHost, message))
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return err.Error()
	}
	type Message struct {
		Message string `json:"message"`
	}

	tmp := &Message{}
	json.Unmarshal(bytes, tmp)

	return tmp.Message
}

func getGoogleExcelValueById(id int64) string {
	url := "https://script.google.com/macros/s/AKfycbzDtZfQHmr0YJF7F_m2ZfatU7Hu-FwTpBTwQfYXqZAv7P1JnHQ/exec?msg=" + fmt.Sprintf("%d", id)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("err:\n" + err.Error())
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read error", err)
		return ""
	}

	type Tmp struct {
		Msg interface{}
	}

	test := Tmp{}
	if err := json.Unmarshal(body, &test); err != nil {
		log.Print(err.Error())
		return ""
	}

	switch reflect.TypeOf(test.Msg).Kind() {
	case reflect.Int:
		return fmt.Sprintf("%d", test.Msg.(int))
	case reflect.Int8:
		return fmt.Sprintf("%d", test.Msg.(int8))
	case reflect.Int16:
		return fmt.Sprintf("%d", test.Msg.(int16))
	case reflect.Int32:
		return fmt.Sprintf("%d", test.Msg.(int32))
	case reflect.Int64:
		return fmt.Sprintf("%d", test.Msg.(int64))
	case reflect.String:
		return test.Msg.(string)
	case reflect.Float64:
		return fmt.Sprintf("%.f", test.Msg.(float64))
	case reflect.Float32:
		return fmt.Sprintf("%.f", test.Msg.(float32))
	default:
		fmt.Println(reflect.TypeOf(test.Msg).Kind())
		return "unknow type"
	}

	return "unexcept error"
}

var (
	baseURL = "https://www.1juzi.com/"
)

const (
	baseRegexp = `<li><a href="/([a-z]+)/">(.{4,6})</a></li>`
	subRegexp  = `<li><h3><a href="(.{0,20})" title="(.{0,10})" target="_blank">`
)

type URLStruct struct {
	CategoryName string
	URL          string
}

func getEveryDaySentence() string {
	juziSubURL := setJuziURL(baseURL, baseRegexp)
	r := getRandomNumber(len(juziSubURL))
	subListURL := setJuziURL(baseURL+juziSubURL[r].URL, subRegexp)
	lr := getRandomNumber(len(subListURL))
	result := getPageSource(baseURL + subListURL[lr].URL)
	rp := regexp.MustCompile(`<p>([0-9]+)、(.*?)</p>`)
	items := rp.FindAllStringSubmatch(result, -1)
	ir := getRandomNumber(len(items))
	return fmt.Sprintf("%s > %s:\n\n%s", juziSubURL[r].CategoryName, subListURL[lr].CategoryName, items[ir][2])
}

func getRandomNumber(number int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int() % number
}

func setJuziURL(url string, regex string) []URLStruct {
	juziSubURL := []URLStruct{}
	result := getPageSource(url)
	rp := regexp.MustCompile(regex)
	items := rp.FindAllStringSubmatch(result, -1)
	for _, item := range items {
		tmp := URLStruct{CategoryName: item[2], URL: item[1]}
		juziSubURL = append(juziSubURL, tmp)
	}

	return juziSubURL
}

func getPageSource(url string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Http get err:", err)

	}
	if resp.StatusCode != 200 {
		fmt.Println("Http status code:", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read error", err)
	}
	result := ConvertToString(string(body), "gbk", "utf-8")

	return strings.Replace(result, "\n", "", -1)
}

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}
