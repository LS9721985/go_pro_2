package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type binancePrice struct {
	Price float64 `json:"price,string"`
	Code  int64   `json:"code"`
}

type wallet map[string]float64

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		msgText := update.Message.Text
		chatID := update.Message.Chat.ID
		command := strings.Split(msgText, " ")

		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(chatID, "Валюта добавлена"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, err.Error()))
			}
			if _, ok := db[chatID]; !ok {
				db[chatID] = wallet{}
			}

			db[chatID][command[1]] += amount

			balanceText := fmt.Sprintf("%f", db[chatID][command[1]])
			bot.Send(tgbotapi.NewMessage(chatID, balanceText))

		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(chatID, "Валюта вычтена"))
			}
			amount, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, err.Error()))
			}
			if _, ok := db[chatID]; !ok {
				continue
			}
			if amount >= db[chatID][command[1]] {
				db[chatID][command[1]] = 0
			} else {
				db[chatID][command[1]] -= amount
			}

			balanceText := fmt.Sprintf("%f", db[chatID][command[1]])
			bot.Send(tgbotapi.NewMessage(chatID, balanceText))

		case "DEL":
			if len(command) != 2 {
				delete(db[chatID], command[1])
				bot.Send(tgbotapi.NewMessage(chatID, "Валюта удалена"))
			}

		case "SHOW":
			msg := ""
			var sum float64 = 0
			for key, amount := range db[chatID] {
				price, _ := getPrice(key)
				msg += fmt.Sprintf("%s: %.2f\n", key, amount*price)
				sum += amount * price
			}
			msg += fmt.Sprintf("Total (RUB): %.2f\n", sum)
			bot.Send(tgbotapi.NewMessage(chatID, msg))
		case "QUIT":
			break

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Команда не найдена"))
		}
	}
}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	var jsonResp binancePrice

	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}

	//if jsonResp.Code != 0 {

	//}

	priceRUB, _ := getRuble2USDPrice()
	price = jsonResp.Price * priceRUB
	return
}

func getRuble2USDPrice() (price float64, err error) {
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	if err != nil {
		return
	}

	defer resp.Body.Close()

	var jsonResp binancePrice

	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}

	//if jsonResp.Code != 0 {

	//}

	price = jsonResp.Price
	return

}

func getToken() string {
	return "1858238605:AAEdwnu8YXGV70_92Dzxg6BKOCjyj1QLtvg"
}
