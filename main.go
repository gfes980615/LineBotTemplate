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
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"gitlab.paradise-soft.com.tw/glob/utils/network"

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
	log.Println(r)
	if err != nil {
		log.Print(err.Error())
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	client := network.NewClient("https://script.google.com/macros/s/AKfycbzDtZfQHmr0YJF7F_m2ZfatU7Hu-FwTpBTwQfYXqZAv7P1JnHQ/exec")
	params := map[string]string{
		"msg": "2",
	}
	buf, err := client.Get(params)
	if err != nil {
		log.Print(err.Error())
		return
	}

	type Tmp struct {
		Msg string
	}

	test := Tmp{}
	if err := json.Unmarshal(buf, &test); err != nil {
		log.Print(err.Error())
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch event.Message.(type) {
			case *linebot.TextMessage:
				// quota, err := bot.GetMessageQuota().Do()
				// if err != nil {
				// 	log.Println("Quota err:", err)
				// }
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("test: "+test.Msg)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}
