package main

import(
	"fmt"
	"net/http"
	"time"
	"io/ioutil"
	"encoding/json"
	"database/sql"
	"strconv"
	"math"

	_ "github.com/go-sql-driver/mysql"
)
//структура данных по сообщению
type BigData struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Text 			string `json:"text"` 
			Date           int `json:"date"`
			ReplyToMessage struct {
				MessageID int `json:"message_id"`
				From      struct {
					ID        int64  `json:"id"`
					IsBot     bool   `json:"is_bot"`
					FirstName string `json:"first_name"`
					Username  string `json:"username"`
				} `json:"from"`
				Chat struct {
					ID        int    `json:"id"`
					FirstName string `json:"first_name"`
					Username  string `json:"username"`
					Type      string `json:"type"`
				} `json:"chat"`
				Date int    `json:"date"`
				Text string `json:"text"`
			} `json:"reply_to_message"`
			Location struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
		} `json:"message"`
	} `json:"result"`
}
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
//структура данных пользователя(при отправке погоды лично) 
type uFoundData struct{
	Username string `json:"username"`
	FirstName string `json:"first_name"`
}
//токен бота
const bot_token  = "5378816043:AAHIEkTEtBH922tAZtHYYVYJU4Rrpv4E2-s"
//наша основная функция
func main() {
	//подключаемся к бд
	db, _ := sql.Open("mysql", "root:inordic123@tcp(mysql:3306)/bot_data")
	//запускаем фунцкию для отправки погоды всем пользователем в отдельном потоке
	go weather(db)
	//запускаем функцию для записи новых юзеров и основного функционала бота
	register(db)
}
//регистрация пользователя и измение или добавление координат
func register(db *sql.DB){
	//создаем массив апод айди юзеров
	users := []int{}
	//делаем оффсет
	offset := 0
	//переменная под айди сообщения на который ответили
	findMessageId := 1
	//в цикле отправляем запросы для получения новых сообщений и записи их в бд
	for range time.Tick(time.Second){
		//отправляем запрос и сохраняем данные полученные из него в перемнную
		resp, _ := http.Get("https://api.telegram.org/bot" + bot_token + "/getUpdates?offset=" + strconv.Itoa(offset))
		//получаем тело запроса
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		//создаем структуру под все сообщения
		data := BigData{}
		//переливаем данные из тела запроса в структуру
		json.Unmarshal(bodyBytes, &data)
		//в цикле бежим по всем сообщениям
		for i := 0; i < len(data.Result); i++{
			//записываем в переменные основные данные о пришедшем сообщении
			userId := data.Result[i].Message.From.ID
			cTime := data.Result[i].Message.Date
			message := data.Result[i].Message.Text
			firstName := data.Result[i].Message.From.FirstName
			userName := data.Result[i].Message.From.Username
			latitude := 0.0
			longitude := 0.0
			latitude =	data.Result[i].Message.Location.Latitude
			longitude = data.Result[i].Message.Location.Longitude
			messageId := data.Result[i].Message.MessageID
			replyedMessageId := data.Result[i].Message.ReplyToMessage.MessageID
			//проверяем есть ли юзер в бд
			if !inArray(userId, users) {
				//если нет то регистрируем и отправляем приветсвенный текст
				addUser(userId, cTime, userName, firstName, db, 0.0, 0.0)
				users = append(users, userId)
				text := "Привет " + firstName + ", чтобы начать введи команду /info"
				sendMessage(text, userId)
			} else{
				if latitude != 0 && longitude != 0{
					//если пользователь отправил локацию записываем её и отправляем ему погоду по ней
					addUser(userId, cTime, userName, firstName, db, latitude, longitude) 
					text := "поздравляем, локация успешно добавлена/изменена"
					sendMessage(text, userId)
					weatherOnId(db, userId)
				} else if message == "/find" {
					//если пользователь хочет использовать функцию по поиску людей рядом:
					text := "Ответьте на это сообщение нужным вам расстоянием в метрах"
					sendMessage(text, userId)
					//записываем айди данного сообщения, на него пользователь должен кинуть ответ в виде радиуса поиска в метрах
					findMessageId = messageId + 1
					fmt.Println("findmessage", findMessageId)
				} else if message == "/info"{
					//сообщение основной информации по боту
					text := "Привет, наш бот создан для получения погоды, а также поиска людей поблизости по вашей локации. Краткое руководство:  - чтобы установить локацию отправте в чат свою локацию, - чтобы её изменить сделайте то же самое, - нажмите на кнопку меню чтобы увидеть все команды,  - вы будете получать прогноз погоды каждые 30 минут, а также при каждой отправке своей локации в бота"
					sendMessage(text, userId)
				} else if replyedMessageId == findMessageId{
					//ловим событие по ответу на сообщение из /find
					distance, _ := strconv.Atoi(message)
					if distance != 0{
						//запускаем функцию для отправки людей поблизости
						sendNearbyPe(userId, distance, db)
					}
				} else{
					//обработка ошибки: если пользователь отправил обычный текст отправляем ему сообщение об ошибке
					text := "ничего не понял..., попробуй еще раз"
					sendMessage(text, userId)
				}
			} 
			

			//сохраняем в базу новое сообщение
			addMessage(userId, cTime, message, db)

			//обновляем значения оффсет
			offset = data.Result[i].UpdateID + 1
		}
	}
}
//функция для проверки есть ли то или инное в масиве
func inArray(needle int, haystack []int ) (bool) {
	//для каждого юзера в массиве смотрим совпадает ли его айди с пришедшим айди
	for i := 0; i < len(haystack); i++{
		if needle == haystack[i]{
			return  true
		}
	}
	return false
}
//функция для добаления сообщений в бд
func addMessage(user_id int, time int, text string, db *sql.DB){
	//отправляем запрос 
	db.Exec("INSERT INTO `messages`(`time`, `content`, `user_id`) VALUES(?, ?, ?)", time, text, user_id)
}
//функция для добаления юзеров в бд
func addUser(user_id int, time int, username string, first_name string, db *sql.DB, lat float64, lon float64){
	if lat == 0.0 && lon == 0.0{
		//если сообщение - не координаты  добавляем юзера в бд 
		db.Exec("INSERT INTO `users`(`registration_time`, `username`, `id`, `first_name`, `latitude`, `longitude`) VALUES(?, ?, ?, ?, 0.0, 0.0)", time, username, user_id, first_name)
	} else{
		//в противном случае добавляем координаты к пользователю
		db.Exec("UPDATE `users` SET `latitude` = ?, `longitude` = ? WHERE `id`=?", lat, lon, user_id)
	}
}
//функция для нахождения людей поблизости
func sendNearbyPe(chat_id int, distance int, db *sql.DB){
	//получаем долготу и широту пользователя по айди
	u_data, _ := db.Query("SELECT `latitude`, `longitude` FROM `users` WHERE `id`=?", chat_id)
	var latitude float64 = 0
	var longitude float64 = 0
	for u_data.Next(){
		u_data.Scan(&latitude, &longitude)
	}
	if latitude == 0 && longitude == 0{
		//обработка ошибки: если локация не настроена отправляем ошибку
		text := "вы не добавили свою локацию, отправте её в чат и попробуйте еще раз"
		sendMessage(text, chat_id)
	} else{
		//иначе переводим радиус в граду для широты и долготы
		aroundLat := getAroundLoc(latitude, distance)
		aroundLon := getAroundLoc(longitude, distance)
		foundUsersData := []uFoundData{}
		//отправялем запрос по поиску людей рядом
		m_data, err := db.Query("SELECT `username`, `first_name` FROM `users` WHERE `latitude` between ? and ? AND `longitude` between ? and ?", latitude - aroundLat, latitude + aroundLat, longitude - aroundLon, longitude + aroundLon)
		if err != nil{
			fmt.Println(err)
		}
		for m_data.Next(){
			foundUser := uFoundData{}
			m_data.Scan(&foundUser.Username, &foundUser.FirstName)
			foundUsersData = append(foundUsersData, foundUser)
		}
		text := "Люди рядом: "
		for i := 0; i < len(foundUsersData); i++{
			//в цикле добавляем в сообщение каждого нашедшегося пользователся (имя и юзернейм)
			if foundUsersData[i].Username == ""{
				//обработка ошибки: если юзернейм пуст то добавляем юзера без него
				text += string(foundUsersData[i].FirstName) + ",  "
			} else{
				//иначе отправляем и имя и юзернейм
				text += string(foundUsersData[i].FirstName) + " @" + string(foundUsersData[i].Username) + ",  "
			}
		}
		sendMessage(text, chat_id)
	}
}
//функция для отправки погоды раз в 30 минут
func weather(db *sql.DB){
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
			req.Header.Add("X-RapidAPI-Key", "043ba3497emsh35fe7d49bdc695ap18d919jsnb86bc5dc874b")
			req.Header.Add("X-RapidAPI-Host", "weatherbit-v1-mashape.p.rapidapi.com")
			//отправляем запрос к апи погоды по координатам
			res, _ := http.DefaultClient.Do(req)
			defer res.Body.Close()
			bodyWeather, _ := ioutil.ReadAll(res.Body)
			wData := weatherData{}
			json.Unmarshal(bodyWeather, &wData)
			text := `Локация: ` + wData.Data[i].CityName + `, Температура : ` + fmt.Sprintf("%.1f", wData.Data[i].Temp) + `°С`
			//отправлляем сообщение
			sendMessage(text, users[i].Id)	
		}
	}
}
//функция для получения погоды сразу при отправке сообщения
func weatherOnId(db *sql.DB, chat_id int){
	userData, _ := db.Query("SELECT `id`, `latitude`, `longitude`  FROM `users` WHERE `id`= ?", chat_id)
	user := uWeatherData{}
	for userData.Next(){
		userData.Scan(&user.Id, &user.Lat, &user.Lon)
	}
	uLat := fmt.Sprintf("%f", user.Lat)
	uLon := fmt.Sprintf("%f", user.Lon)
	url := "https://weatherbit-v1-mashape.p.rapidapi.com/current?lat=" + uLat + "&lon=" + uLon 
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-RapidAPI-Key", "043ba3497emsh35fe7d49bdc695ap18d919jsnb86bc5dc874b")
	req.Header.Add("X-RapidAPI-Host", "weatherbit-v1-mashape.p.rapidapi.com")
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	bodyWeather, _ := ioutil.ReadAll(res.Body)
	wData := weatherData{}
	json.Unmarshal(bodyWeather, &wData)
	text := `Локация: ` + wData.Data[0].CityName + `, Температура : ` + fmt.Sprintf("%.1f", wData.Data[0].Temp) + `°С`
	sendMessage(text, user.Id)		
}
//универсальная функция отправки сообщения
func sendMessage(text string, chat_id int){
	//отправляем сообщения
	http.Get("https://api.telegram.org/bot" + bot_token +"/sendMessage?chat_id=" + strconv.Itoa(chat_id) + "&text=" + text)
}
//функция вычисления примерного расстояноия от пользователя
func getAroundLoc(degrees float64, distance int) float64{
	const nPi = math.Pi
	const EarthRadius = 6371210
	var Distance float64 = float64(distance)
	//переводим градусы в радианы
	radLoc := degrees * nPi / 180
	//вычисляем дельту
	delta := nPi / 180 * EarthRadius * math.Cos(radLoc)
	//вычиисляем радиус поиска
	around := Distance / delta
	return around
}
