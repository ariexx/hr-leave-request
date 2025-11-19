ALTER TABLE employees
ADD COLUMN role VARCHAR(50) NULL DEFAULT 'employee'
AFTER password;