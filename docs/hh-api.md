# hh.ru API — что нужно для проекта

## Аутентификация

- Заголовок: `Authorization: Bearer {ACCESS_TOKEN}`
- Заголовок `HH-User-Agent` обязателен для всех запросов: `MyApp/1.0 (feedback@example.com)`
- Без токена: работает, но после первого запроса может потребовать капчу (`403`)
- Токен получаем через OAuth 2.0, задаём через переменную окружения `HH_TOKEN`

---

## Нужные эндпоинты

### 1. Поиск вакансий

```
GET /vacancies
```

**Параметры поиска:**

| Параметр | Описание | Пример |
|---|---|---|
| `text` | Текст поиска | `"Go developer"` |
| `area` | ID региона | `1` = Москва, `2` = СПб |
| `professional_role` | ID профессиональной роли | из справочника `/professional_roles` |
| `experience` | Опыт работы | `noExperience`, `between1And3`, `between3And6`, `moreThan6` |
| `employment_form` | Тип занятости | из справочника |
| `work_format` | Формат работы | `REMOTE`, `ON_SITE`, `HYBRID` |
| `work_schedule_by_days` | График | из справочника |
| `salary` | Зарплата (нижняя граница) | `150000` |
| `currency` | Валюта | `RUR`, `USD`, `EUR` |
| `label` | Метки | `with_salary` — только с зарплатой |
| `period` | Дней с публикации | `7`, `30` |
| `order_by` | Сортировка | `relevance`, `salary_desc`, `publication_time` |
| `per_page` | Результатов на странице | макс. `100`, по умолчанию `20` |
| `page` | Номер страницы (0-based) | `0`, `1`, ... |

> **Лимит:** максимум 2000 результатов (`page * per_page ≤ 2000`)

**Ответ:**
```json
{
  "items": [ ...вакансии... ],
  "found": 1234,
  "pages": 50,
  "page": 0,
  "per_page": 20
}
```

---

### 2. Детали вакансии

```
GET /vacancies/{vacancy_id}
```

Возвращает полное описание одной вакансии. Используем после поиска, чтобы получить `description` и `key_skills`.

**Ключевые поля ответа:**

| Поле | Описание |
|---|---|
| `id` | ID вакансии |
| `name` | Название должности |
| `description` | Полное описание (HTML) |
| `key_skills` | Массив навыков `[{name}]` |
| `area` | Город `{id, name}` |
| `employer` | Компания `{id, name, trusted}` |
| `salary_range` | Зарплата `{from, to, currency, gross}` |
| `experience` | Требуемый опыт `{id, name}` |
| `employment_form` | Тип занятости `{id, name}` |
| `work_format` | Формат `[{id, name}]` — REMOTE/ON_SITE/HYBRID |
| `work_schedule_by_days` | График `[{id, name}]` |
| `professional_roles` | Профроли `[{id, name}]` |
| `published_at` | Дата публикации |
| `alternate_url` | Ссылка на вакансию |
| `contacts` | Контакты работодателя |
| `languages` | Языки `[{id, level}]` |

---

### 3. Справочники

```
GET /dictionaries
```

Возвращает все справочники одним запросом. Нужен для декодирования ID в читаемые названия.

Нужные словари:
- `experience` — уровни опыта
- `work_format` — форматы работы
- `work_schedule_by_days` — графики
- `currency` — валюты

---

### 4. Регионы

```
GET /areas
```

Дерево регионов. Нужен для получения ID нужного города.

Часто используемые:
- `1` — Москва
- `2` — Санкт-Петербург
- `113` — Россия (все регионы)

---

## Коды ответов

| Код | Ситуация | Что делать |
|---|---|---|
| `200` | OK | — |
| `400` | Неверные параметры | Логировать, пропустить |
| `403` | Нет токена / капча | Добавить токен, уменьшить частоту |
| `404` | Вакансия не найдена / удалена | Помечать в БД как `deleted` |
| `429` | Слишком много запросов | Добавить задержку, retry с backoff |
| `500/503` | Ошибка сервера | Retry с backoff |

---

## Rate limiting

- Рекомендуемая частота: **не чаще 3 req/sec** (~300 мс между запросами)
- Без токена риск капчи уже после первого запроса
- При `429` — exponential backoff

---

## Пайплайн получения данных

```
1. GET /dictionaries              → кешируем справочники
2. GET /vacancies?text=...        → собираем список (постранично, до 2000)
3. GET /vacancies/{id}            → детали каждой вакансии (description, key_skills)
4. Сохраняем в PostgreSQL
```

> Детали запрашиваем отдельно, потому что поиск возвращает краткую версию без `description` и `key_skills`.

---

## Поля для хранения в PostgreSQL

Минимальный набор для проекта:

```
id              TEXT PRIMARY KEY
name            TEXT
description     TEXT        -- полное описание из GET /vacancies/{id}
employer_name   TEXT
employer_id     TEXT
area_name       TEXT
salary_from     INT
salary_to       INT
salary_currency TEXT
experience      TEXT
work_format     TEXT[]      -- массив: REMOTE, ON_SITE, HYBRID
key_skills      TEXT[]      -- массив навыков
professional_roles TEXT[]
published_at    TIMESTAMPTZ
alternate_url   TEXT
raw_json        JSONB       -- сырой ответ API на случай если понадобятся доп. поля
created_at      TIMESTAMPTZ DEFAULT NOW()
```

---

## Поля для генерации резюме (передаём в MCP)

Это данные, которые Go передаёт в `resume_server.py`:

```json
{
  "vacancy_id": "...",
  "title": "...",
  "description": "...",
  "key_skills": ["Go", "PostgreSQL", "Docker"],
  "experience_required": "between1And3",
  "work_format": ["REMOTE"],
  "employer_name": "...",
  "salary_range": { "from": 150000, "to": 250000, "currency": "RUR" }
}
```
