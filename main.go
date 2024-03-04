package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Gbot *tgbotapi.BotAPI
var Token string
var ChatId int64

const (
	EMOJI_BUTTON_START  = "\U000025B6  " // ▶
	EMOJI_BUTTON_END    = "  \U000025C0" // ◀
	EMOJI_BUTTON_RHOT   = "\U0001F321"   // 🌡
	EMOJI_BUTTON_NORMAL = "\U0001F31D"   // 🌝
	EMOJI_BUTTON_COLD   = "\U0001F976"   // 🥶
	EMOJI_BUTTON_HOT    = "\U0001F31E"   // 🌞

	BUTTON_TEXT_PRINT_INTRO  = EMOJI_BUTTON_START + "Да, хочу узнать, как он работает" + EMOJI_BUTTON_END
	BUTTON_TEXT_PRINT_RHOT   = "Лучше не выходить на улицу. Очень жарко" + EMOJI_BUTTON_RHOT
	BUTTON_TEXT_PRINT_NORMAL = "На улице комфортно" + EMOJI_BUTTON_NORMAL
	BUTTON_TEXT_PRINT_COLD   = "На улице прохладно" + EMOJI_BUTTON_COLD
	BUTTON_TEXT_PRINT_HOT    = "На улице жарко" + EMOJI_BUTTON_HOT

	BUTTON_CODE_PRINT_INTRO = "print_intro"
	BUTTON_CODE_SKIP_INTRO  = "skip_intro"

	TOKEN_NAME_IN_OS = "398844992"
)

func init() {
	_ = os.Setenv("TOKEN_NAME_IN_OS", "6838765270:AAEwcSFxYsfdA-QXI8j_bwazdrNfqC5vOi4")
	var err error
	Token = os.Getenv("TOKEN_NAME_IN_OS")

	if Gbot, err = tgbotapi.NewBotAPI(Token); err != nil {
		log.Panic(err)
	}

	Gbot.Debug = true

}

func isCallbackQuery(update *tgbotapi.Update) bool {
	return update.CallbackQuery != nil && update.CallbackQuery.Data != ""
}
func isStartMessage(update *tgbotapi.Update) bool {
	return update.Message != nil && update.Message.Text == "/start"
}

func printIntro(*tgbotapi.Update) {
	Gbot.Send(tgbotapi.NewMessage(ChatId,
		"Привет! Я бот, который поможет тебе узнать погоду в городе. Просто напиши мне название города, о котором ты хочешь узнать погоду, и я предоставлю тебе актуальную информацию.\n\n      Например, 'Moscow', если ты хочешь узнать погоду на сегодня.\n\n      Например, '7 Moscow', если тебе нужен прогноз на несколько дней вперед."))
}

func isTemperature(update *tgbotapi.Update) bool {
	return update.Message != nil
}

type WeatherData struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func Temperature(update *tgbotapi.Update) {
	apiKey := "2acb96138cb24ff884593705240103"
	city := update.Message.Text
	cityUp := strings.ToUpper(string(city[0])) + city[1:]
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, cityUp)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Ошибка при выполнении запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return
	}

	var data WeatherData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}
	roundedTemp := math.Round(data.Current.TempC)
	if roundedTemp >= 24 && roundedTemp <= 30 {
		tempt := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(BUTTON_TEXT_PRINT_HOT))
		Gbot.Send(tempt)
	}
	if roundedTemp >= 17 && roundedTemp < 24 {
		tempt := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(BUTTON_TEXT_PRINT_NORMAL))
		Gbot.Send(tempt)
	}
	if roundedTemp <= 10 {
		tempt := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(BUTTON_TEXT_PRINT_COLD))
		Gbot.Send(tempt)
	}
	if roundedTemp >= 32 {
		tempt := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprint(BUTTON_TEXT_PRINT_RHOT))
		Gbot.Send(tempt)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Текущая температура в %s: %.0f °C\n", cityUp, roundedTemp))
	_, err = Gbot.Send(msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

type WeatherDataMore struct {
	Forecast struct {
		Forecastday []struct {
			Date string `json:"date"`
			Day  struct {
				Maxtemp_c float64 `json:"maxtemp_c"`
			} `json:"day"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func TemperatureMore(update *tgbotapi.Update) {
	apiKey := "2acb96138cb24ff884593705240103"
	messageParts := strings.Fields(update.Message.Text)
	if len(messageParts) < 2 {
		fmt.Println("Invalid input")
		return
	}
	daysStr := messageParts[0]
	city := messageParts[1]

	cityUp := strings.ToUpper(string(city[0])) + city[1:]

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		fmt.Println("Error converting days to int:", err)
		return
	}

	url := fmt.Sprintf("http://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&days=%s", apiKey, city, daysStr)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}

	var data WeatherDataMore
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshalling data:", err)
		return
	}

	if days >= 8 {
		Gbot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите количество дней до 7."))
	} else if days <= 1 {
		Gbot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Пожалуйста, укажите количество дней от 2."))
	} else {
		for _, forecast := range data.Forecast.Forecastday {
			roundedTemp := math.Round(forecast.Day.Maxtemp_c)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Температура на %s в %s: %.0f °C\n", forecast.Date, cityUp, roundedTemp))
			_, err = Gbot.Send(msg)
			if err != nil {
				log.Println("Error sending message:", err)
			}
		}
	}
}
func getKeyboardRow(buttonText, buttonCode string) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(buttonText, buttonCode))
}
func askToPrintIntro() {
	msg := tgbotapi.NewMessage(ChatId, "Хочешь узанать как работает бот? ")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		getKeyboardRow(BUTTON_TEXT_PRINT_INTRO, BUTTON_CODE_PRINT_INTRO),
	)
	Gbot.Send(msg)
}

func updateProcessing(update *tgbotapi.Update) {
	choiceCode := update.CallbackQuery.Data
	log.Printf("[%T] %s", time.Now(), choiceCode)

	switch choiceCode {
	case BUTTON_CODE_PRINT_INTRO:
		printIntro(update)
	}
}
func main() {

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := Gbot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if isCallbackQuery(&update) {
			updateProcessing(&update)

		} else if isStartMessage(&update) {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			ChatId = update.Message.Chat.ID
			askToPrintIntro()
		} else if isTemperature(&update) {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			ChatId = update.Message.Chat.ID
			if unicode.IsDigit([]rune(update.Message.Text)[0]) {
				TemperatureMore(&update)
			} else {
				Temperature(&update)
			}
		}
	}
}
