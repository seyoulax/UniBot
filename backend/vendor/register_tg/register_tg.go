package register_tg

import(
	"io/ioutil"
	"net/http"
	"encoding/json"
	"time"
	"weather"
	"near_people"
	"strconv"
	"oStuff"
	"database/sql"
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
func Register(db *sql.DB, ApiKey string, bot_token string){
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
			if !oStuff.InArray(userId, users) {
				//если нет то регистрируем и отправляем приветсвенный текст
				addUser(userId, cTime, userName, firstName, db, 0.0, 0.0)
				users = append(users, userId)
				text := "Привет " + firstName + ", чтобы начать введи команду /info"
				oStuff.SendMessage(text, userId, bot_token)
			} else{
				if latitude != 0 && longitude != 0{
					//если пользователь отправил локацию записываем её и отправляем ему погоду по ней
					addUser(userId, cTime, userName, firstName, db, latitude, longitude) 
					text := "поздравляем, локация успешно добавлена/изменена"
					oStuff.SendMessage(text, userId, bot_token)
					weather.WeatherOnId(db, userId, ApiKey, bot_token)
				} else if message == "/find" {
					//если пользователь хочет использовать функцию по поиску людей рядом:
					text := "Ответьте на это сообщение нужным вам расстоянием в метрах"
					oStuff.SendMessage(text, userId, bot_token)
					//записываем айди данного сообщения, на него пользователь должен кинуть ответ в виде радиуса поиска в метрах
					findMessageId = messageId + 1
				} else if message == "/info"{
					//сообщение основной информации по боту
					text := "Привет, наш бот создан для получения погоды, а также поиска людей поблизости по вашей локации. Краткое руководство:  - чтобы установить локацию отправте в чат свою локацию, - чтобы её изменить сделайте то же самое, - нажмите на кнопку меню чтобы увидеть все команды,  - вы будете получать прогноз погоды каждые 30 минут, а также при каждой отправке своей локации в бота"
					oStuff.SendMessage(text, userId, bot_token)
				} else if replyedMessageId == findMessageId{
					//ловим событие по ответу на сообщение из /find
					distance, _ := strconv.Atoi(message)
					if distance != 0{
						//запускаем функцию для отправки людей поблизости
						near_people.SendNearbyPe(userId, distance, db, bot_token)
					}
				} else{
					//обработка ошибки: если пользователь отправил обычный текст отправляем ему сообщение об ошибке
					text := "ничего не понял..., попробуй еще раз"
					oStuff.SendMessage(text, userId, bot_token)
				}
			} 
			

			//сохраняем в базу новое сообщение
			addMessage(userId, cTime, message, db)

			//обновляем значения оффсет
			offset = data.Result[i].UpdateID + 1
		}
	}
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