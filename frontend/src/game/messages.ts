const SERVER_ERROR_MESSAGES: Record<string, string> = {
  "game is finished": "Партия уже завершена.",
  "invalid move": "Такой ход невозможен.",
  "piece not found": "На выбранной клетке нет вашей шашки.",
  "wrong piece color": "Сейчас ходит другой игрок.",
  "destination occupied": "Эта клетка занята.",
  "capture is required": "Есть обязательное взятие.",
  "must continue capture": "Нужно продолжить взятие той же шашкой.",
  "invalid position": "Клетка вне доски.",
  "invalid game state": "Состояние партии устарело. Обновите страницу.",
  "websocket connection error": "Соединение с партией потеряно.",
  "websocket is not connected": "Соединение с партией еще не установлено.",
  "invalid websocket message": "Сервер прислал неожиданный ответ.",
  "unknown websocket message": "Сервер прислал неизвестное событие.",
  "failed to send move": "Не удалось отправить ход.",
};

export function friendlyGameError(message: string): string {
  return SERVER_ERROR_MESSAGES[message] ?? message;
}
