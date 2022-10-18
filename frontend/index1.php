<?php
// определяем кодировку
header('Content-type: text/html; charset=utf-8');
// Создаем объект бота
$bot = new Bot();
// Обрабатываем пришедшие данные
$bot->init('php://input');

/**
 * Class Bot
 */
class Bot
{
    // <bot_token> - созданный токен для нашего бота от @BotFather
    private $botToken = "5327059939:AAGr9otM_gS8FWzzuuHePa93zhnSJPCSnqg";
    // адрес для запросов к API Telegram
    private $apiUrl = "https://api.telegram.org/bot";

    public function init($data)
    {
        // создаем массив из пришедших данных от API Telegram
        $arrData = $this->getData($data);

        // лог
        // $this->setFileLog($arrData);

        if (array_key_exists('message', $arrData)) {
            $chat_id = $arrData['message']['chat']['id'];
            $message = $arrData['message']['text'];

        } elseif (array_key_exists('callback_query', $arrData)) {
            $chat_id = $arrData['callback_query']['message']['chat']['id'];
            $message = $arrData['callback_query']['data'];
        }

        $justKeyboard = $this->getKeyBoard([[["text" => "Голосовать"], ["text" => "Помощь"]]]);

        $inlineKeyboard = $this->getInlineKeyBoard([[
            ['text' => hex2bin('F09F918D') . ' 0', 'callback_data' => 'vote_1_0_0'],
            ['text' => hex2bin('F09F918E') . ' 0', 'callback_data' => 'vote_0_0_0']
        ]]);

        switch ($message) {
            case '/start':
                $dataSend = array(
                    'text' => "Приветствую, давайте начнем нашу практику. Нажмите на кнопку Голосовать.",
                    'chat_id' => $chat_id,
                    'reply_markup' => $justKeyboard,
                );
                $this->requestToTelegram($dataSend, "sendMessage");
                break;
            case 'Голосовать':
                $dataSend = array(
                    'text' => "Выберите один из вариантов",
                    'chat_id' => $chat_id,
                    'reply_markup' => $inlineKeyboard,
                );
                $this->requestToTelegram($dataSend, "sendMessage");
                break;
            case 'Помощь':
                $dataSend = array(
                    'text' => "Просто нажмите на кнопку Голосовать.",
                    'chat_id' => $chat_id,
                );
                $this->requestToTelegram($dataSend, "sendMessage");
                break;
            case (preg_match('/^vote/', $message) ? true : false):
                $params = $this->setParams($message);
                $dataSend = array(
                    'reply_markup' => $params[0],
                    'message_id' => $arrData['callback_query']['message']['message_id'],
                    'chat_id' => $chat_id,
                );
                $this->changeVote($dataSend, $params[1], $arrData['callback_query']['id']);
                break;
            default:
                $dataSend = array(
                    'text' => "Не запланированная реакция, может просто нажмете на кнопку Голосовать.",
                    'chat_id' => $chat_id,
                );
                $this->requestToTelegram($dataSend, "sendMessage");
                break;
        }
    }

    /** Меняем клавиатуру Vote
     * @param $data
     * @param $emogi
     * @param $callback_query_id
     */
    private function changeVote($data, $emoji, $callback_query_id)
    {
        $text = $this->requestToTelegram($data, "editMessageReplyMarkup")
            ? "Вы проголосовали " . hex2bin($emoji)
            : "Непредвиденная ошибка, попробуйте позже.";

        $this->requestToTelegram([
            'callback_query_id' => $callback_query_id,
            'text' => $text,
            'cache_time' => 30,
        ], "answerCallbackQuery");
    }

    /** Устанавливаем новые значения Vote
     * @param $data
     * @return string
     */
    private function setParams($data)
    {
        $params = explode("_", $data);
        $params[1] ? $params[2]++ : $params[3]++;
        $arr[] = $this->getInlineKeyBoard([[
            ['text' => hex2bin('F09F918D') . ' ' . $params[2], 'callback_data' => 'vote_1_' . $params[2] . '_' . $params[3]],
            ['text' => hex2bin('F09F918E') . ' ' . $params[3], 'callback_data' => 'vote_0_' . $params[2] . '_' . $params[3]]
        ]]);
        $arr[] = $params[1] ? 'F09F918D' : 'F09F918E';
        return $arr;
    }

    /**
     * создаем inline клавиатуру
     * @return string
     */
    private function getInlineKeyBoard($data)
    {
        $inlineKeyboard = array(
            "inline_keyboard" => $data,
        );
        return json_encode($inlineKeyboard);
    }

    /**
     * создаем клавиатуру
     * @return string
     */
    private function getKeyBoard($data)
    {
        $keyboard = array(
            "keyboard" => $data,
            "one_time_keyboard" => false,
            "resize_keyboard" => true
        );
        return json_encode($keyboard);
    }

    private function setFileLog($data)
    {
        $fh = fopen('log.txt', 'a') or die('can\'t open file');
        ((is_array($data)) || (is_object($data))) ? fwrite($fh, print_r($data, TRUE) . "\n") : fwrite($fh, $data . "\n");
        fclose($fh);
    }

    /**
     * Парсим что приходит преобразуем в массив
     * @param $data
     * @return mixed
     */
    private function getData($data)
    {
        return json_decode(file_get_contents($data), TRUE);
    }

    /** Отправляем запрос в Телеграмм
     * @param $data
     * @param string $type
     * @return mixed
     */
    private function requestToTelegram($data, $type)
    {
        $result = null;

        if (is_array($data)) {
            $ch = curl_init();
            curl_setopt($ch, CURLOPT_URL, $this->apiUrl . $this->botToken . '/' . $type);
            curl_setopt($ch, CURLOPT_POST, count($data));
            curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($data));
            $result = curl_exec($ch);
            curl_close($ch);
        }
        return $result;
    }
}
