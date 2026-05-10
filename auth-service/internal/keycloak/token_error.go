package keycloak

import (
	"encoding/json"
	"strings"
)

type tokenErrJSON struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// PublicTokenErrorMessage преобразует ответ токен-эндпоинта Keycloak в сообщение для клиента API
// (без деталей реализации и сырого тела ответа).
func PublicTokenErrorMessage(httpStatus int, body []byte) string {
	if httpStatus >= 500 {
		return "Сервис авторизации временно недоступен. Попробуйте позже."
	}
	if httpStatus == 0 || len(body) == 0 {
		return "Не удалось выполнить вход. Попробуйте позже."
	}
	var t tokenErrJSON
	if err := json.Unmarshal(body, &t); err != nil || t.Error == "" {
		return "Не удалось выполнить вход. Попробуйте позже."
	}
	desc := strings.ToLower(strings.TrimSpace(t.ErrorDescription))
	switch t.Error {
	case "invalid_grant":
		if strings.Contains(desc, "account is not fully set up") {
			return "Вход невозможен: учётная запись не готова к входу. Завершите обязательные действия (например, смену временного пароля или заполнение профиля) или обратитесь к администратору."
		}
		if strings.Contains(desc, "invalid user credentials") ||
			strings.Contains(desc, "invalid username or password") {
			return "Неверный логин или пароль."
		}
		if strings.Contains(desc, "not allowed") || strings.Contains(desc, "grant") {
			return "Этот способ входа для учётной записи недоступен. Обратитесь к администратору."
		}
		return "Не удалось войти. Проверьте логин и пароль."
	case "invalid_client", "unauthorized_client":
		return "Ошибка конфигурации входа. Обратитесь к администратору."
	case "invalid_request":
		return "Некорректный запрос авторизации."
	default:
		return "Не удалось выполнить вход. Попробуйте позже."
	}
}
