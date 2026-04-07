--Дата, на которую назначена задача
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS scheduled_for DATE NOT NULL DEFAULT CURRENT_DATE;

--Ссылка на шаблон
ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS template_id BIGINT NULL;

--Шапка периодической задачи
CREATE TABLE IF NOT EXISTS task_templates (
                                              id BIGSERIAL PRIMARY KEY,
                                              title TEXT NOT NULL,
                                              description TEXT NOT NULL DEFAULT '',
                                              status TEXT NOT NULL,
                                              active BOOLEAN NOT NULL DEFAULT TRUE,
                                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

--Настройка периодичности
CREATE TABLE IF NOT EXISTS task_recurrences (
                                                template_id BIGINT PRIMARY KEY REFERENCES task_templates(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,

    every_n_days INT NULL,
    day_of_month INT NULL,
    parity TEXT NULL,
    specific_dates JSONB NULL,

    CHECK (type IN ('daily', 'monthly', 'specific_dates', 'day_parity')),
    CHECK (every_n_days IS NULL OR every_n_days > 0),
    CHECK (day_of_month IS NULL OR day_of_month BETWEEN 1 AND 30),
    CHECK (parity IS NULL OR parity IN ('even', 'odd'))
    );

ALTER TABLE tasks
    ADD CONSTRAINT fk_tasks_template
        FOREIGN KEY (template_id) REFERENCES task_templates(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_template_id ON tasks (template_id);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_for ON tasks (scheduled_for);
CREATE INDEX IF NOT EXISTS idx_task_templates_active ON task_templates (active);

CREATE UNIQUE INDEX IF NOT EXISTS ux_tasks_template_date
    ON tasks (template_id, scheduled_for)
    WHERE template_id IS NOT NULL;