import { ApiError } from "../api/http";
import type { Color, GameStatus, SearchStatus } from "../types/domain";
import type { SocketStatus } from "../ws/gameSocket";

const API_ERROR_BY_CODE: Record<string, string> = {
  invalid_request: "Проверьте введенные данные.",
  invalid_username: "Имя пользователя указано неверно.",
  invalid_password: "Пароль должен быть не короче 8 символов.",
  username_taken: "Такое имя пользователя уже занято.",
  email_taken: "Такая почта уже используется.",
  invalid_credentials: "Неверное имя пользователя или пароль.",
  unauthorized: "Нужно войти в аккаунт.",
  invalid_token: "Сессия устарела. Войдите заново.",
  forbidden: "У вас нет доступа к этой партии.",
  not_found: "Данные не найдены.",
  game_not_found: "Партия не найдена.",
  profile_not_found: "Профиль не найден.",
  invalid_display_name: "Отображаемое имя не должно быть длиннее 40 символов.",
  invalid_country_code: "Укажите двухбуквенный код страны.",
  invalid_avatar_url: "Укажите корректную ссылку на изображение.",
  invalid_bio: "Описание не должно быть длиннее 300 символов.",
  invalid_cursor: "Не удалось продолжить загрузку истории.",
  too_many_profiles: "Запрошено слишком много профилей.",
  already_matched: "Матч уже создан.",
  internal_error: "Сервис временно недоступен.",
  not_ready: "Сервис еще запускается.",
  request_failed: "Запрос не выполнен.",
};

const API_ERROR_BY_MESSAGE: Record<string, string> = {
  "invalid request body": "Проверьте введенные данные.",
  "invalid credentials provided": "Неверное имя пользователя или пароль.",
  "failed to search match": "Не удалось начать поиск игры.",
  "failed to get matchmaking status": "Не удалось обновить статус поиска.",
  "failed to cancel matchmaking": "Не удалось отменить поиск.",
  "failed to load ratings": "Не удалось загрузить рейтинг.",
  "failed to load game": "Не удалось загрузить партию.",
  "failed to send move": "Не удалось отправить ход.",
  "database can't be reached": "База данных временно недоступна.",
};

export function friendlyApiError(error: unknown, fallback = "Что-то пошло не так."): string {
  if (error instanceof ApiError) {
    return API_ERROR_BY_CODE[error.code] ?? API_ERROR_BY_MESSAGE[error.message] ?? fallback;
  }

  if (error instanceof Error) {
    return API_ERROR_BY_MESSAGE[error.message] ?? error.message;
  }

  return fallback;
}

export function colorLabel(color?: Color): string {
  switch (color) {
    case "white":
      return "Белые";
    case "black":
      return "Черные";
    default:
      return "-";
  }
}

export function gameStatusLabel(status?: GameStatus): string {
  switch (status) {
    case "active":
      return "Идет игра";
    case "finished":
      return "Игра завершена";
    default:
      return "-";
  }
}

export function matchmakingStatusLabel(status?: SearchStatus | "idle"): string {
  switch (status) {
    case "waiting":
      return "Ожидание соперника";
    case "matching":
      return "Создаем партию";
    case "matched":
      return "Матч найден";
    case "idle":
    case undefined:
      return "";
    default:
      return "Неизвестно";
  }
}

export function socketStatusLabel(status: SocketStatus): string {
  switch (status) {
    case "connecting":
      return "Подключаемся";
    case "open":
      return "Онлайн";
    case "closed":
      return "Нет связи";
  }
}
