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
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/line/line-bot-sdk-go/linebot"

	_ "github.com/joho/godotenv/autoload"
)

var bot *linebot.Client

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
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
				id, transferErr := strconv.ParseInt(message.Text, 10, 64)
				text := getGoogleExcelValueById(id)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(text)).Do(); err != nil {
					log.Print(err)
				}
				if transferErr != nil {
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(transferErr.Error())).Do(); err != nil {
						log.Print(err)
					}
					return
				}

			}
		}
	}
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", test.Msg.(int))
	case reflect.String:
		return test.Msg.(string)
	default:
		return "unknow type"
	}

	return "except error"
}
