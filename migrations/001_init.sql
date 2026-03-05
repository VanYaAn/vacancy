-- +goose Up
CREATE TABLE IF NOT EXISTS vacancies (
    id                 TEXT PRIMARY KEY,
    name               TEXT        NOT NULL,
    description        TEXT,
    employer_id        TEXT,
    employer_name      TEXT,
    employer_trusted   BOOLEAN     DEFAULT FALSE,
    area_id            TEXT,
    area_name          TEXT,
    salary_from        INT,
    salary_to          INT,
    salary_currency    TEXT,
    salary_gross       BOOLEAN,
    experience         TEXT,
    work_formats       TEXT[]      DEFAULT '{}',
    key_skills         TEXT[]      DEFAULT '{}',
    professional_roles TEXT[]      DEFAULT '{}',
    published_at       TIMESTAMPTZ,
    alternate_url      TEXT,
    deleted            BOOLEAN     DEFAULT FALSE,
    created_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS resumes (
    id          TEXT PRIMARY KEY,
    vacancy_id  TEXT        NOT NULL REFERENCES vacancies (id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vacancies_published_at ON vacancies (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_vacancies_deleted      ON vacancies (deleted);
CREATE INDEX IF NOT EXISTS idx_resumes_vacancy_id     ON resumes (vacancy_id);

-- +goose Down
DROP INDEX IF EXISTS idx_resumes_vacancy_id;
DROP INDEX IF EXISTS idx_vacancies_deleted;
DROP INDEX IF EXISTS idx_vacancies_published_at;
DROP TABLE IF EXISTS resumes;
DROP TABLE IF EXISTS vacancies;
