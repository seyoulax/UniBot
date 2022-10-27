package oStuff

import(
	"net/http"
	"strconv"
)

//функция для проверки есть ли то или инное в масиве
func InArray(needle int, haystack []int ) (bool) {
	//для каждого юзера в массиве смотрим совпадает ли его айди с пришедшим айди
	for i := 0; i < len(haystack); i++{
		if needle == haystack[i]{
			return  true
		}
	}
	return false
}
//универсальная функция отправки сообщения
func SendMessage(text string, chat_id int, bot_token string){
	//отправляем сообщения
	http.Get("https://api.telegram.org/bot" + bot_token +"/sendMessage?chat_id=" + strconv.Itoa(chat_id) + "&text=" + text)
}
