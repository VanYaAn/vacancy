-- +goose Up
CREATE TABLE IF NOT EXISTS resume_pdfs (
    id          TEXT PRIMARY KEY,
    vacancy_id  TEXT        NOT NULL REFERENCES vacancies (id) ON DELETE CASCADE,
    latex_src   TEXT        NOT NULL,
    pdf_data    BYTEA       NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resume_pdfs_vacancy_id ON resume_pdfs (vacancy_id);

-- +goose Down
DROP INDEX IF EXISTS idx_resume_pdfs_vacancy_id;
DROP TABLE IF EXISTS resume_pdfs;
