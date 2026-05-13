CREATE TABLE IF NOT EXISTS records (
    id         INTEGER  NOT NULL,
    version    INTEGER  NOT NULL,
    data       JSON     NOT NULL,
    created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, version)
);

CREATE VIEW IF NOT EXISTS records_latest AS
SELECT id, version, data, created_on
FROM records r
WHERE version = (
    SELECT MAX(version) FROM records WHERE id = r.id
);
