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
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	bot            *linebot.Client
	dianaHost      string
	MapleServerMap = map[string]bool{
		"izcr": true, // 愛麗西亞
		"izr":  true, // 艾莉亞
		"plt":  true, // 普力特
		"ld":   true, // 琉德
		"yen":  true, // 優依那
		"slc":  true, // 殺人鯨
	}
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

	eventByte, err := json.Marshal(events)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	url := "http://%s/callback_lineTemplate?events=%s"
	resp, err := http.Get(fmt.Sprintf(url, dianaHost, string(eventByte)))
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(resp)
	log.Print(eventByte)
	log.Print(string(eventByte))
}
