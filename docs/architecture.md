# Architecture

## Обзор

Go-приложение оркестрирует весь пайплайн: получает вакансии с hh.ru, сохраняет в PostgreSQL,
генерирует резюме через LLM и экспортирует данные в Google Sheets.
Python MCP-серверы берут на себя задачи, где Python удобнее.

## Структура проекта

```
mcpProp/
├── cmd/
│   └── main.go                        # точка входа, инициализация, DI
│
├── internal/
│   ├── domain/                        # модели и интерфейсы (нет внешних зависимостей)
│   │   ├── vacancy.go                 # Vacancy, VacancyDetail, Salary, SearchParams
│   │   ├── resume.go                  # Resume, ResumeTemplate
│   │   └── repository.go             # интерфейсы: VacancyRepository, ResumeRepository
│   │
│   ├── service/                       # бизнес-логика (Use Cases)
│   │   ├── vacancy_service.go         # поиск и сохранение вакансий
│   │   ├── resume_service.go          # оркестрация генерации резюме
│   │   └── export_service.go          # экспорт в Google Sheets
│   │
│   └── infrastructure/                # реализации внешних зависимостей
│       ├── hh/
│       │   └── client.go              # HTTP клиент hh.ru API
│       ├── postgres/
│       │   └── vacancy_repo.go        # репозиторий PostgreSQL
│       └── mcp/
│           ├── resume_client.go       # Go клиент → resume_server.py
│           └── sheets_client.go       # Go клиент → sheets_server.py
│
└── server/                            # Python MCP серверы
    ├── resume_server.py               # MCP сервер #1: генерация резюме через LLM
    └── sheets_server.py               # MCP сервер #2: экспорт данных в Google Sheets
```

## Слои и зависимости

```
cmd/main.go
  ↓ создаёт
infrastructure/*        (знает о domain, внешних SDK)
  ↓ передаёт через интерфейсы в
service/*               (знает только о domain)
  ↓ работает с
domain/*                (ничего не знает о внешнем мире)
```

Зависимости направлены только внутрь. `domain` — чистые модели без импортов внешних пакетов.

## Компоненты

### Go-приложение (оркестратор)
- Главная точка управления всем пайплайном
- Запускает Python MCP серверы как subprocess
- Общается с ними через stdin/stdout по MCP протоколу

### hh.ru API client (`infrastructure/hh`)
- Поиск вакансий по параметрам
- Получение детальной информации по вакансии
- Rate limiting (~3 req/sec)
- Опциональная авторизация через Bearer-токен

### PostgreSQL (`infrastructure/postgres`)
- Хранение всех найденных вакансий
- Хранение сгенерированных резюме
- Связь вакансия → резюме

### MCP сервер #1 — Resume (`server/resume_server.py`)
- **Задача:** генерация резюме под конкретную вакансию
- **Вход:** данные вакансии + шаблон резюме
- **Логика:** LLM анализирует требования вакансии и адаптирует блоки шаблона
- **Выход:** готовый текст резюме

### MCP сервер #2 — Sheets (`server/sheets_server.py`)
- **Задача:** экспорт данных в Google Sheets
- **Вход:** список вакансий (и/или резюме)
- **Логика:** запись в таблицу через gspread
- **Выход:** ссылка на таблицу

## Пайплайн (поток данных)

```
hh.ru API
  ↓ SearchVacancies / GetVacancy
PostgreSQL
  ↓ сохранены вакансии
MCP resume_server.py  ←  vacancy data + template
  ↓ LLM генерирует резюме
PostgreSQL
  ↓ сохранено резюме, привязано к вакансии
MCP sheets_server.py  ←  vacancies + resumes
  ↓ gspread
Google Sheets          →  финальный результат для анализа
```

## Решения и обоснования

| Решение | Обоснование |
|---|---|
| Go как оркестратор | Надёжный, быстрый, удобен для системной логики и конкурентности |
| Python для резюме | Удобные LLM-библиотеки, гибкая работа со строками |
| Python для Sheets | `gspread` значительно проще Google API клиента на Go |
| Два отдельных MCP сервера | Single Responsibility — каждый сервер решает одну задачу, легче дебажить и заменять |
| Clean Architecture | Возможность заменить любой компонент (MCP → прямой LLM API, Postgres → другая БД) без переписывания бизнес-логики |

## Ещё не определено

- Критерии поиска вакансий (фильтры, ключевые слова)
- Структура шаблона резюме (блоки, из которых LLM собирает итоговый вариант)
- Схема таблицы PostgreSQL
- Структура Google Sheets (колонки, листы)
