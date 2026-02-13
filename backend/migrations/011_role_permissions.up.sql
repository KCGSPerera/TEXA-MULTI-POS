CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

INSERT INTO permissions (code, description) VALUES
('sale.create', 'Create sales'),
('refund.create', 'Create refunds'),
('inventory.adjust', 'Adjust inventory'),
('grn.approve', 'Approve GRN'),
('transfer.create', 'Create branch transfers'),
('transfer.complete', 'Complete branch transfers'),
('report.view', 'View reports')
ON CONFLICT (code) DO NOTHING;
