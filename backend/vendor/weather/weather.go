package weather

import(
	"fmt"
	"time"
	"net/http"
	"encoding/json"
	"oStuff"
	"io/ioutil"
	"database/sql"
)
//структура данных погоды
type weatherData struct{
	Data []struct {
		WindCdir     string  `json:"wind_cdir"`
		Rh           int     `json:"rh"`
		Pod          string  `json:"pod"`
		Lon          float64 `json:"lon"`
		Pres         float64 `json:"pres"`
		Timezone     string  `json:"timezone"`
		ObTime       string  `json:"ob_time"`
		CountryCode  string  `json:"country_code"`
		Clouds       int     `json:"clouds"`
		Vis          int     `json:"vis"`
		WindSpd      float64 `json:"wind_spd"`
		Snow         int     `json:"snow"`
		WindCdirFull string  `json:"wind_cdir_full"`
		Slp          int     `json:"slp"`
		Datetime     string  `json:"datetime"`
		Ts           int     `json:"ts"`
		HAngle       int     `json:"h_angle"`
		Aqi          int     `json:"aqi"`
		Uv           int     `json:"uv"`
		WindDir      int     `json:"wind_dir"`
		ElevAngle    float64 `json:"elev_angle"`
		Ghi          int     `json:"ghi"`
		Dhi          int     `json:"dhi"`
		Precip       int     `json:"precip"`
		Station      string  `json:"station"`
		Sunset       string  `json:"sunset"`
		Temp         float64 `json:"temp"`
		Sunrise      string  `json:"sunrise"`
		AppTemp      float64 `json:"app_temp"`
		SolarRad     int     `json:"solar_rad"`
		Weather      struct {
			Code        int    `json:"code"`
			Icon        string `json:"icon"`
			Description string `json:"description"`
		} `json:"weather"`
		Lat       float64  `json:"lat"`
		Gust      float64  `json:"gust"`
		CityName  string   `json:"city_name"`
		StateCode string   `json:"state_code"`
		Sources   []string `json:"sources"`
		Dni       int      `json:"dni"`
		Dewpt     float64  `json:"dewpt"`
	} `json:"data"`
	Count int `json:"count"`
}
//структура данных пользователя(при отправке погоды всем) 
type uWeatherData struct {
	Id int 
	Lat float64
	Lon float64
}
func WeatherLoop(db *sql.DB, ApiKey string, bot_token string){
	// отправляем погоду всем пользователям раз в 30 минут
	for range time.Tick(30 * time.Minute){
		//выбираем всех пользователей у которых указана локация
		usersData, _ := db.Query("SELECT `id`, `latitude`, `longitude`  FROM `users` WHERE `latitude` != 0.0 AND `longitude` != 0.0")
		users := []uWeatherData{}
		for usersData.Next(){
			user := uWeatherData{}
			usersData.Scan(&user.Id, &user.Lat, &user.Lon)
			users = append(users, user)
		}
		//в цикле для каждого пользователя отправляем ему сообщение с информацией о погоду
		for i := 0; i < len(users); i++{
			//переводим данные float в стринг с определенной точностью
			uLat := fmt.Sprintf("%f", users[i].Lat)
			uLon := fmt.Sprintf("%f", users[i].Lon)
			url := "https://weatherbit-v1-mashape.p.rapidapi.com/current?lat=" + uLat + "&lon=" + uLon 
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Add("X-RapidAPI-Key", ApiKey)
			req.Header.Add("X-RapidAPI-Host", "weatherbit-v1-mashape.p.rapidapi.com")
			//отправляем запрос к апи погоды по координатам
			res, _ := http.DefaultClient.Do(req)
			defer res.Body.Close()
			bodyWeather, _ := ioutil.ReadAll(res.Body)
			wData := weatherData{}
			json.Unmarshal(bodyWeather, &wData)
			text := `Локация: ` + wData.Data[i].CityName + `, Температура : ` + fmt.Sprintf("%.1f", wData.Data[i].Temp) + `°С`
			//отправлляем сообщение
			oStuff.SendMessage(text, users[i].Id, bot_token)	
		}
	}
}
func WeatherOnId(db *sql.DB, chat_id int, ApiKey string, bot_token string){
	userData, _ := db.Query("SELECT `id`, `latitude`, `longitude`  FROM `users` WHERE `id`= ?", chat_id)
	user := uWeatherData{}
	for userData.Next(){
		userData.Scan(&user.Id, &user.Lat, &user.Lon)
	}
	uLat := fmt.Sprintf("%f", user.Lat)
	uLon := fmt.Sprintf("%f", user.Lon)
	url := "https://weatherbit-v1-mashape.p.rapidapi.com/current?lat=" + uLat + "&lon=" + uLon 
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-RapidAPI-Key", ApiKey)
	req.Header.Add("X-RapidAPI-Host", "weatherbit-v1-mashape.p.rapidapi.com")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	bodyWeather, _ := ioutil.ReadAll(res.Body)
	wData := weatherData{}
	json.Unmarshal(bodyWeather, &wData)
	text := `Локация: ` + wData.Data[0].CityName + `, Температура : ` + fmt.Sprintf("%.1f", wData.Data[0].Temp) + `°С`
	oStuff.SendMessage(text, user.Id, bot_token)		
}