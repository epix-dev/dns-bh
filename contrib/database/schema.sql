CREATE TABLE hazard (
    id SERIAL,
    entry_pos INT NOT NULL,
    entry_add TIMESTAMP NOT NULL,
    entry_del TIMESTAMP,
    domain VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    PRIMARY KEY(id)
);

CREATE UNIQUE INDEX idx_hazard_1 ON hazard(entry_pos) WHERE deleted_at IS NULL;

CREATE TABLE malware (
    id SERIAL,
    domain VARCHAR(255) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    source VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    PRIMARY KEY(id)
);

CREATE UNIQUE INDEX idx_malware_1 ON malware(domain,reason,source) WHERE deleted_at IS NULL;

CREATE TABLE whitelist (
    id SERIAL,
    domain VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    PRIMARY KEY(id)
);

CREATE UNIQUE INDEX idx_whitelist_1 ON whitelist(domain) WHERE deleted_at IS NULL;

CREATE TABLE cert_hole (
    id SERIAL,
    remote_id INTEGER NOT NULL,
    domain VARCHAR(255) NOT NULL,
    insert_time TIMESTAMP NOT NULL,
    delete_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    PRIMARY KEY(id)
);

CREATE UNIQUE INDEX idx_cert_hole_1 ON cert_hole(domain) WHERE deleted_at IS NULL;
