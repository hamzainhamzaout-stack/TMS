-- +goose Up
-- +goose StatementBegin

-- User and role management is restricted to the "admin" role, but it was
-- never seeded: the seeded admin account was given "fleet_manager".
INSERT INTO role (name) VALUES ('admin') ON CONFLICT (name) DO NOTHING;

-- Move the seeded admin account onto the admin role.
UPDATE users
SET role_id = (SELECT id FROM role WHERE name = 'admin'),
    updated_at = NOW()
WHERE user_name = 'admin' OR phone_number = '09876567890';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- users.role_id cascades on delete, so move admins off the role before
-- dropping it instead of deleting their accounts.
UPDATE users
SET role_id = (SELECT id FROM role WHERE name = 'fleet_manager'),
    updated_at = NOW()
WHERE role_id = (SELECT id FROM role WHERE name = 'admin');

DELETE FROM role WHERE name = 'admin';

-- +goose StatementEnd
