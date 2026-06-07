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
  "no draw offer": "Предложение ничьей уже недоступно.",
  "draw already offered": "Предложение ничьей уже отправлено.",
  "cannot answer own draw offer": "Нельзя ответить на собственное предложение ничьей.",
  "game is not active": "Партия уже завершена.",
};

export function friendlyGameError(message: string): string {
  return SERVER_ERROR_MESSAGES[message] ?? message;
}

const INPUT_ERROR_MESSAGES = new Set([
  "invalid move",
  "piece not found",
  "wrong piece color",
  "destination occupied",
  "capture is required",
  "must continue capture",
  "invalid position",
  "not player's turn",
]);

export function isGameInputError(message: string): boolean {
  return INPUT_ERROR_MESSAGES.has(message);
}
