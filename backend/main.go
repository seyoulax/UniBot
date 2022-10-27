package main

import(
	"database/sql"
	"weather"
	"register_tg"
	"io/ioutil"
	"encoding/json"

	_ "github.com/go-sql-driver/mysql"
)
//структура конфигурации
type Config struct {
	Database struct {
		User         string `json:"user"`
		Password     string `json:"password"`
		DatabaseName string `json:"database_name"`
	} `json:"database"`
	BotToken string `json:"bot_token"`
	WeatherApiKey string `json:"weatherApiKey"`
}
var bot_token string = ""
//наша основная функция
func main() {
	config := Config{}
	//считываем данные из файла
	ConfigJson, _ := ioutil.ReadFile("config/config.json") 
	//раскодируем в структру
	json.Unmarshal(ConfigJson, &config)
	//записываем ключ в переменную
	bot_token  = config.BotToken
	//подключаемся к бд
	db, _ := sql.Open("mysql", config.Database.User + ":" + config.Database.Password + "@tcp(mysql:3306)/" + config.Database.DatabaseName + "")
	//запускаем фунцкию для отправки погоды всем пользователем в отдельном потоке
	go weather.WeatherLoop(db, config.WeatherApiKey, bot_token)
	//запускаем функцию для записи новых юзеров и основного функционала бота
	register_tg.Register(db, config.WeatherApiKey, bot_token)
}
