"""
MCP сервер для генерации резюме под вакансию.

Флоу:
  1. Принимает данные вакансии
  2. Отправляет в Groq LLM → получает адаптированный LaTeX код
  3. Компилирует LaTeX через tectonic → PDF
  4. Возвращает PDF в base64 + LaTeX исходник

Требует:
  GROQ_API_KEY — ключ Groq API
"""

import base64
import json
import os
import subprocess
import tempfile

from groq import Groq
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("resume-server")

# ── Базовое резюме (факты которые не меняются) ────────────────────────────────

BASE_RESUME = """
Антошин Иван Андреевич
Go developer / Backend Developer
Телефон: +79030634514
Город: Москва
Email: ivanantoshin176@gmail.com

ОБРАЗОВАНИЕ:
НИУ МЭИ Москва — Прикладная математика и информатика: математическое моделирование, 2022–2026

ЯЗЫКИ:
- Русский: Родной
- Английский: B2 (Upper-Intermediate)

ПЕТ-ПРОЕКТЫ:
1. Сервис на gRPC с базой данных (Go, gRPC, PostgreSQL, Docker)
   - Реализовал CRUD-операции с хранением данных в PostgreSQL
   - Настроил Docker-окружение для сервиса и базы данных

2. Сервис управления заказами (Go, Apache Kafka, PostgreSQL, Docker)
   - Разработал потребитель Kafka для чтения и обработки сообщений о заказах
   - Реализовал REST API эндпоинт для получения заказов по ID из кэша
   - Создал статический фронтенд на HTML/JavaScript
   - Контейнеризовал приложение с Docker Compose

3. Сервис REST API маркетплейса (Go, PostgreSQL, Docker)
   - Реализовал CRUD операции с авторизацией через JWT
   - Использовал Gorilla Mux для маршрутизации
   - Контейнеризовал приложение с Docker Compose с поддержкой миграций

БАЗОВЫЕ НАВЫКИ:
- Язык: Go
- Технологии: Docker, PostgreSQL, gRPC, REST API, Apache Kafka, Git
- Другие языки: Python, SQL

ЛИЧНЫЕ КАЧЕСТВА: Ответственный, усидчивый, адаптивный, внимательный.
"""

# ── LaTeX шаблон ──────────────────────────────────────────────────────────────

LATEX_TEMPLATE = r"""
\documentclass[11pt,a4paper]{article}
\usepackage{fontspec}
\usepackage{polyglossia}
\setdefaultlanguage{russian}
\setotherlanguage{english}
\setmainfont{PT Serif}
\newfontfamily\cyrillicfont{PT Serif}
\newfontfamily\cyrillicfontsf{PT Serif}
\newfontfamily\cyrillicfonttt{PT Serif}
\usepackage[margin=1.5cm]{geometry}
\usepackage{hyperref}
\usepackage{enumitem}
\usepackage{titlesec}
\usepackage{parskip}

\hypersetup{colorlinks=true, urlcolor=blue}
\titleformat{\section}{\large\bfseries}{}{0em}{}[\titlerule]
\setlist[itemize]{noitemsep, topsep=2pt, leftmargin=1.5em}
\pagestyle{empty}

\begin{document}

%%CONTENT%%

\end{document}
"""


@mcp.tool()
def generate_resume(data: str) -> str:
    """
    Генерирует адаптированное резюме под вакансию.

    Args:
        data: JSON строка с данными вакансии:
              - vacancy_id, title, description, key_skills,
                experience, work_formats, employer_name, salary

    Returns:
        JSON строка с полями:
        - vacancy_id: str
        - latex_src: str  (LaTeX исходник)
        - pdf_base64: str (PDF в base64)
    """
    vacancy = json.loads(data)

    latex_content = _generate_latex(vacancy)
    pdf_bytes = _compile_latex(latex_content)
    pdf_b64 = base64.b64encode(pdf_bytes).decode("utf-8")

    return json.dumps({
        "vacancy_id": vacancy.get("vacancy_id", ""),
        "latex_src": latex_content,
        "pdf_base64": pdf_b64,
    })


def _generate_latex(vacancy: dict) -> str:
    """Вызывает Groq LLM и получает LaTeX тело резюме."""
    client = Groq(api_key=os.environ["GROQ_API_KEY"])

    prompt = f"""Ты эксперт по написанию резюме. Тебе нужно адаптировать резюме кандидата под конкретную вакансию.

БАЗОВОЕ РЕЗЮМЕ КАНДИДАТА:
{BASE_RESUME}

ВАКАНСИЯ:
Название: {vacancy.get('title', '')}
Компания: {vacancy.get('employer_name', '')}
Требуемый опыт: {vacancy.get('experience', '')}
Формат работы: {', '.join(vacancy.get('work_formats') or [])}
Ключевые навыки из вакансии: {', '.join(vacancy.get('key_skills') or [])}
Описание вакансии:
{vacancy.get('description', '')[:2000]}

ЗАДАЧА:
Напиши ТОЛЬКО LaTeX код тела документа (без \\documentclass и преамбулы — только содержимое между \\begin{{document}} и \\end{{document}}).

Правила адаптации:
1. Имя, контакты, образование, языки — не меняй
2. Раздел "Обо мне" — перепиши акцент под требования вакансии
3. Ключевые навыки — выдели вперёд те навыки которые совпадают с вакансией
4. Пет-проекты — сделай акцент на проектах релевантных стеку вакансии
5. Личные качества — подбери формулировки под вакансию
6. Используй только \\section, \\textbf, itemize, \\href — простые LaTeX команды
7. Весь текст на русском языке
8. Не добавляй выдуманный опыт — только то что есть в базовом резюме

Верни ТОЛЬКО LaTeX код без объяснений и markdown блоков.
"""

    response = client.chat.completions.create(
        model="llama-3.3-70b-versatile",
        messages=[{"role": "user", "content": prompt}],
        temperature=0.3,
        max_tokens=3000,
    )

    content = response.choices[0].message.content.strip()

    # убираем markdown блоки если LLM обернул в ```latex ... ```
    if content.startswith("```"):
        lines = content.split("\n")
        content = "\n".join(lines[1:-1] if lines[-1].strip() == "```" else lines[1:])

    # убираем преамбульные команды которые LLM мог добавить в тело
    preamble_cmds = (
        r"\documentclass", r"\usepackage", r"\begin{document}",
        r"\end{document}", r"\setmainfont", r"\newfontfamily",
    )
    filtered = [
        line for line in content.split("\n")
        if not any(line.strip().startswith(cmd) for cmd in preamble_cmds)
    ]
    return "\n".join(filtered).strip()


def _compile_latex(latex_content: str) -> bytes:
    """Компилирует LaTeX в PDF через tectonic."""
    full_latex = LATEX_TEMPLATE.replace("%%CONTENT%%", latex_content)

    with tempfile.TemporaryDirectory() as tmpdir:
        tex_path = os.path.join(tmpdir, "resume.tex")
        pdf_path = os.path.join(tmpdir, "resume.pdf")

        with open(tex_path, "w", encoding="utf-8") as f:
            f.write(full_latex)

        result = subprocess.run(
            ["tectonic", "--chatter", "minimal", tex_path],
            capture_output=True,
            text=True,
            cwd=tmpdir,
        )

        if result.returncode != 0:
            raise RuntimeError(f"tectonic failed:\n{result.stderr}")

        with open(pdf_path, "rb") as f:
            return f.read()


if __name__ == "__main__":
    mcp.run(transport="stdio")
