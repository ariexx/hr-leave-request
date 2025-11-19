-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees
ADD COLUMN role VARCHAR(50) NULL DEFAULT 'employee'
AFTER password;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees
DROP COLUMN role;
-- +goose StatementEnd
