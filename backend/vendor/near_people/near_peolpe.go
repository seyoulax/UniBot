package near_people

import(
	"math"
	"oStuff"
	"database/sql"
	"fmt"
)
//структура данных пользователя(при отправке погоды лично) 
type uFoundData struct{
	Username string `json:"username"`
	FirstName string `json:"first_name"`
} 
//функция для нахождения людей поблизости
func SendNearbyPe(chat_id int, distance int, db *sql.DB, bot_token string){
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
		oStuff.SendMessage(text, chat_id, bot_token)
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
		oStuff.SendMessage(text, chat_id, bot_token)
	}
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