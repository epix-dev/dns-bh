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

CREATE UNIQUE INDEX idx_whitelist_1 ON malware(domain) WHERE deleted_at IS NULL;
