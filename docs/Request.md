# Финансовая система API (Postman)

## Базовый URL
```
http://localhost:8082
```

## Аутентификация

### Регистрация
- **URL**: `/auth/register`
- **Метод**: `POST`
- **Тело**:
```json
{
  "username": "user1",
  "password": "secure_password",
  "email": "user@example.com",
  "full_name": "Иван Иванов"
}
```
- **Возможные ошибки**:
```json
{
  "error": "username already exists"
}
```

### Вход
- **URL**: `/auth/login`
- **Метод**: `POST`
- **Тело**:
```json
{
  "username": "user1",
  "password": "secure_password"
}
```
- **Ответ**: Содержит `access_token` для использования в последующих запросах
- **Возможные ошибки**:
```json
{
  "error": "invalid credentials"
}
```

### Обновление токена
- **URL**: `/auth/refresh`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {refresh_token}`
- **Возможные ошибки**:
```json
{
  "error": "refresh token is expired"
}
```

## Вклады

### Создание вклада
- **URL**: `/deposit/create`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "amount": 1000,
  "term": 12,
  "interest_rate": 5.5,
  "currency": "RUB"
}
```
- **Возможные ошибки**:
```json
{
  "error": "insufficient funds"
}
```

### Удаление вклада
- **URL**: `/deposit/delete`
- **Метод**: `DELETE`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "deposit_id": 123
}
```
- **Возможные ошибки**:
```json
{
  "error": "deposit not found"
}
```

### Перевод между счетами
- **URL**: `/deposit/transfer`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "from_deposit_id": 123,
  "to_deposit_id": 456,
  "amount": 500.00,
  "bank_name": "Example Bank"
}
```
- **Коды ответов**:
  - `200 OK`: Трансфер успешно выполнен
  - `400 Bad Request`: Ошибка в параметрах запроса
  - `401 Unauthorized`: Требуется аутентификация
  - `404 Not Found`: Один или оба депозита не найдены
  - `500 Internal Server Error`: Внутренняя ошибка сервера
  
- **Пример успешного ответа**:
```json
{
  "message": "transfer completed successfully",
  "transfer": {
    "from_account": 123,
    "to_account": 456,
    "amount": 500.00,
    "bank": "Example Bank",
    "timestamp": "2025-03-18T20:01:40Z"
  }
}
```

- **Возможные ошибки**:
```json
{
  "error": "valid account IDs are required"
}
```
> Эта ошибка возникает, если:
> - Значения `from_deposit_id` или `to_deposit_id` равны или меньше нуля
> - Не указано поле `bank_name`
> - Один или оба депозита не найдены в базе данных
> - Депозиты принадлежат разным пользователям (при отсутствии прав на межпользовательские переводы)
> 
> **Важно**: Убедитесь, что ваш JSON-запрос содержит ключи именно в таком формате:
> - `from_deposit_id` (а не fromAccount, from_account и т.д.)
> - `to_deposit_id` (а не toAccount, to_account и т.д.)
> - `bank_name` (обязательное поле)
> - `amount` (положительное число)

```json
{
  "error": "insufficient funds for transfer"
}
```

### Заморозка вклада
- **URL**: `/deposit/freeze`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "deposit_id": 123,
  "reason": "По запросу клиента"
}
```
- **Возможные ошибки**:
```json
{
  "error": "deposit already frozen"
}
```

### Блокировка вклада
- **URL**: `/deposit/block`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "deposit_id": 123,
  "reason": "Подозрительная активность"
}
```
- **Возможные ошибки**:
```json
{
  "error": "insufficient permissions"
}
```

### Разблокировка вклада
- **URL**: `/deposit/unblock`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "deposit_id": 123,
  "reason": "Проблема решена"
}
```
- **Возможные ошибки**:
```json
{
  "error": "deposit not blocked"
}
```

### Список вкладов
- **URL**: `/deposit/list`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Возможные ошибки**:
```json
{
  "error": "no deposits found"
}
```

## Кредиты

### Запрос кредита
- **URL**: `/loan/request`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "amount": 5000,
  "term": 36,
  "purpose": "Ремонт",
  "income": 3000
}
```
- **Возможные ошибки**:
```json
{
  "error": "insufficient income for requested amount"
}
```

### Список кредитов пользователя
- **URL**: `/loan/list`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`

### Детали кредита
- **URL**: `/loan/{id}`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Возможные ошибки**:
```json
{
  "error": "loan not found"
}
```

### Внести платеж по кредиту
- **URL**: `/loan/payment`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "loan_id": 234,
  "amount": 150.00
}
```
- **Возможные ошибки**:
```json
{
  "error": "payment amount below minimum required"
}
```

### Процентные ставки
- **URL**: `/loan/rates`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`

## Административная часть

### Ожидающие пользователи
- **URL**: `/admin/pending-users`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Возможные ошибки**:
```json
{
  "error": "insufficient permissions"
}
```

### Подтвердить пользователя
- **URL**: `/admin/approve-user`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "user_id": 5
}
```
- **Возможные ошибки**:
```json
{
  "error": "user already approved"
}
```

### Отклонить пользователя
- **URL**: `/admin/reject-user`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "user_id": 6,
  "reason": "Неполные данные"
}
```
- **Возможные ошибки**:
```json
{
  "error": "user not found"
}
```

### Журнал действий
- **URL**: `/admin/action-logs`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `start_date`, `end_date`, `user_id`, `action_type` (опционально)

### Отменить действия пользователя
- **URL**: `/admin/cancel-user-actions`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "user_id": 5,
  "reason": "Подозрительная активность"
}
```
- **Возможные ошибки**:
```json
{
  "error": "no actions available to cancel"
}
```

### Ожидающие кредиты
- **URL**: `/admin/loans/pending`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`

### Одобрить кредит
- **URL**: `/admin/loans/approve`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "loan_id": 10,
  "interest_rate": 8.5,
  "approved_amount": 5000
}
```
- **Возможные ошибки**:
```json
{
  "error": "loan already processed"
}
```

### Отклонить кредит
- **URL**: `/admin/loans/reject`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "loan_id": 11,
  "reason": "Низкий кредитный рейтинг"
}
```
- **Возможные ошибки**:
```json
{
  "error": "loan not in pending status"
}
```

## Операторская часть

### Статистика
- **URL**: `/operator/statistics`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `start_date`, `end_date` (опционально)
- **Возможные ошибки**:
```json
{
  "error": "invalid date format"
}
```

### Действия пользователя
- **URL**: `/operator/actions`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `user_id`
- **Возможные ошибки**:
```json
{
  "error": "user not found"
}
```

### Последние действия
- **URL**: `/operator/recent-actions`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `limit` (опционально)

### Список пользователей
- **URL**: `/operator/users`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`

### Последнее действие пользователя
- **URL**: `/operator/users/{id}/last-action`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Возможные ошибки**:
```json
{
  "error": "no actions found for user"
}
```

### Отмена операции
- **URL**: `/operator/cancel-action`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "action_id": 123,
  "reason": "По запросу клиента"
}
```
- **Возможные ошибки**:
```json
{
  "error": "action cannot be cancelled"
}
```

### Список транзакций
- **URL**: `/operator/transactions`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `start_date`, `end_date`, `user_id`, `transaction_type` (опционально)
- **Возможные ошибки**:
```json
{
  "error": "invalid transaction type"
}
```

## Менеджерская часть
Доступ к менеджерским API-эндпоинтам через базовый путь `/manager` с требуемой аутентификацией.

### Статистика транзакций (функционал оператора)
- **URL**: `/manager/statistics`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `period` (опционально, возможные значения: `day`, `week`, `month`, `year`, по умолчанию `month`)
- **Ответ**:
```json
{
  "statistics": {
    "total_deposits": 100500,
    "total_withdrawals": 50250,
    "average_transaction_amount": 750.25
  },
  "period": {
    "start": "2023-02-18",
    "end": "2023-03-18",
    "name": "month"
  },
  "by_type": {
    "deposit": 50,
    "withdrawal": 25,
    "transfer": 30
  }
}
```

### Просмотр транзакций (функционал оператора)
- **URL**: `/manager/transactions`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: 
  - `username` (опционально) - фильтр по имени пользователя
  - `type` (опционально) - фильтр по типу транзакции
  - `date` (опционально, формат: YYYY-MM-DD) - фильтр по дате
- **Ответ**:
```json
{
  "transactions": [
    {
      "id": 123,
      "user_id": 5,
      "username": "user1",
      "type": "deposit",
      "amount": 1000.00,
      "timestamp": "2023-03-18T10:30:00Z",
      "metadata": "Initial deposit"
    }
  ]
}
```

### Отмена операции (функционал оператора)
- **URL**: `/manager/transactions/cancel`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "user_id": 5,
  "action_id": 123
}
```
- **Ответ**:
```json
{
  "success": true,
  "message": "Операция успешно отменена"
}
```
- **Возможные ошибки**:
```json
{
  "error": "Операции удаления не могут быть отменены"
}
```

### Получение последнего действия пользователя (функционал оператора)
- **URL**: `/manager/users/{id}/last-action`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Ответ**:
```json
{
  "username": "user1",
  "user_id": 5,
  "action": {
    "id": 123,
    "type": "deposit",
    "amount": 1000.00,
    "timestamp": "2023-03-18T10:30:00Z",
    "can_cancel": true
  }
}
```

### Ожидающие кредиты для подтверждения
- **URL**: `/manager/loans/pending`
- **Метод**: `GET`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Параметры**: `status` (опционально, по умолчанию "pending")
- **Ответ**:
```json
{
  "loans": [
    {
      "id": 123,
      "user_id": 5,
      "username": "user1",
      "type": "standard",
      "amount": 5000.00,
      "term": 12,
      "interest_rate": 8.5,
      "total_payable": 5425.00,
      "monthly_payment": 452.08,
      "status": "pending",
      "created_at": "2023-03-18T10:30:00Z",
      "needs_review": true
    }
  ],
  "total_pending": 1
}
```

### Подтверждение кредита
- **URL**: `/manager/loans/approve`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "loan_id": 123,
  "comment": "Одобрено на стандартных условиях"
}
```
- **Ответ**:
```json
{
  "message": "loan approved successfully",
  "loan_id": 123
}
```
- **Возможные ошибки**:
```json
{
  "error": "loan not found"
}
```
или
```json
{
  "error": "only pending loans can be approved"
}
```

### Отклонение кредита
- **URL**: `/manager/loans/reject`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "loan_id": 123,
  "comment": "Недостаточный кредитный рейтинг"
}
```
- **Ответ**:
```json
{
  "message": "loan rejected successfully",
  "loan_id": 123
}
```
- **Возможные ошибки**:
```json
{
  "error": "comment is required when rejecting a loan"
}
```

### Создание кредита или рассрочки для клиента
- **URL**: `/manager/loans/create`
- **Метод**: `POST`
- **Заголовки**: `Authorization: Bearer {access_token}`
- **Тело**:
```json
{
  "user_id": 5,
  "type": "standard", // "standard" для кредита или "installment" для рассрочки
  "amount": 5000.00,
  "duration": 12, // месяцев
  "action": "approve", // всегда "approve", т.к. создаётся менеджером
  "interest_rate": 8.5 // только для кредитов
}
```
- **Ответ**:
```json
{
  "message": "standard approved successfully",
  "user_id": 5,
  "amount": 5000.00,
  "loan_id": 123
}
```
- **Возможные ошибки**:
```json
{
  "error": "invalid type, must be 'standard' or 'installment'"
}
```
или
```json
{
  "error": "failed to create standard"
}
```

## Проверка работоспособности
- **URL**: `/health`
- **Метод**: `GET`
- **Ответ**:
```json
{
  "status": "up",
  "version": "1.0.0",
  "database_connection": "up"
}
```

## Общие коды ошибок

- **401 Unauthorized**:
```json
{
  "error": "unauthorized access"
}
```

- **403 Forbidden**:
```json
{
  "error": "insufficient permissions"
}
```

- **404 Not Found**:
```json
{
  "error": "resource not found"
}
```

- **500 Internal Server Error**:
```json
{
  "error": "internal server error"
}
```
