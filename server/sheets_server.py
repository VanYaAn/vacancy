"""
MCP сервер для экспорта вакансий и резюме в Google Sheets.

Требует:
  - GOOGLE_CREDENTIALS_FILE: путь к JSON файлу service account
  - GOOGLE_SHEET_ID: ID таблицы (из URL: /spreadsheets/d/{ID}/edit)

Таблица должна быть расшарена на email service account с правами редактора.
"""

import json
import os
import sys
from datetime import datetime

import gspread
from google.oauth2.service_account import Credentials
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("sheets-server")

SCOPES = [
    "https://www.googleapis.com/auth/spreadsheets",
    "https://www.googleapis.com/auth/drive",
]


def get_sheet_client():
    creds_file = os.environ.get("GOOGLE_CREDENTIALS_FILE", "credentials.json")
    creds = Credentials.from_service_account_file(creds_file, scopes=SCOPES)
    return gspread.authorize(creds)


@mcp.tool()
def export_to_sheets(data: str) -> str:
    """
    Экспортирует вакансии и резюме в Google Sheets.

    Args:
        data: JSON строка с полями:
              - vacancies: список вакансий
              - resumes: список резюме

    Returns:
        URL таблицы Google Sheets.
    """
    payload = json.loads(data)
    vacancies = payload.get("vacancies") or []
    resumes = payload.get("resumes") or []

    # индекс резюме по vacancy_id для быстрого поиска
    resume_by_vacancy = {r["vacancy_id"]: r["content"] for r in resumes}

    sheet_id = os.environ.get("GOOGLE_SHEET_ID")
    if not sheet_id:
        raise ValueError("GOOGLE_SHEET_ID environment variable is not set")

    gc = get_sheet_client()
    spreadsheet = gc.open_by_key(sheet_id)

    _export_vacancies(spreadsheet, vacancies, resume_by_vacancy)

    url = f"https://docs.google.com/spreadsheets/d/{sheet_id}/edit"
    return url


def _export_vacancies(spreadsheet, vacancies: list, resume_by_vacancy: dict):
    """Записывает вакансии на лист 'Вакансии'."""
    try:
        ws = spreadsheet.worksheet("Вакансии")
        ws.clear()
    except gspread.WorksheetNotFound:
        ws = spreadsheet.add_worksheet(title="Вакансии", rows=1000, cols=20)

    headers = [
        "ID",
        "Название",
        "Компания",
        "Город",
        "Опыт",
        "Форматы работы",
        "Зарплата от",
        "Зарплата до",
        "Валюта",
        "Навыки",
        "Ссылка",
        "Опубликовано",
        "Резюме готово",
    ]

    rows = [headers]
    for v in vacancies:
        salary = v.get("salary") or {}
        has_resume = "да" if v["id"] in resume_by_vacancy else "нет"

        published = v.get("published_at", "")
        if published:
            try:
                dt = datetime.fromisoformat(published.replace("Z", "+00:00"))
                published = dt.strftime("%d.%m.%Y")
            except Exception:
                pass

        employer = v.get("employer") or {}
        area = v.get("area") or {}

        rows.append([
            v.get("id") or "",
            v.get("name") or "",
            employer.get("name") if isinstance(employer, dict) else v.get("employer_name") or "",
            area.get("name") if isinstance(area, dict) else v.get("area_name") or "",
            v.get("experience") or "",
            ", ".join(v.get("work_formats") or []),
            salary.get("from") or "",
            salary.get("to") or "",
            salary.get("currency") or "",
            ", ".join(v.get("key_skills") or []),
            v.get("alternate_url") or "",
            published,
            has_resume,
        ])

    ws.update(range_name="A1", values=rows, value_input_option="USER_ENTERED")

    # форматирование заголовка
    ws.format("A1:M1", {
        "backgroundColor": {"red": 0.2, "green": 0.5, "blue": 0.9},
        "textFormat": {"bold": True, "foregroundColor": {"red": 1, "green": 1, "blue": 1}},
    })

    # фиксируем первую строку
    spreadsheet.batch_update({
        "requests": [{
            "updateSheetProperties": {
                "properties": {
                    "sheetId": ws.id,
                    "gridProperties": {"frozenRowCount": 1},
                },
                "fields": "gridProperties.frozenRowCount",
            }
        }]
    })


if __name__ == "__main__":
    mcp.run(transport="stdio")
