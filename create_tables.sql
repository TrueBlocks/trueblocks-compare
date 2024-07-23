BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS download_status (
    address VARCHAR(42) PRIMARY KEY,
    provider TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS download_status_unique ON download_status (address, provider);

CREATE TABLE IF NOT EXISTS appearances (
	id INTEGER PRIMARY KEY,
	address VARCHAR(42) NOT NULL,
	block_number INT,
	transaction_index INT,
	provider TEXT
);
CREATE INDEX IF NOT EXISTS appearances_appearance ON appearances (address, block_number, transaction_index);
CREATE INDEX IF NOT EXISTS appearances_provider ON appearances (provider);

CREATE table if not EXISTS appearance_reasons (
	appearance_id INTEGER NOT NULL,
	provider TEXT,
	reason TEXT,
	comment TEXT,
	foreign key(appearance_id) references appearances(id)
);
create index if not exists appearance_reasons_id ON appearance_reasons (appearance_id);

CREATE table if not EXISTS appearance_balance_changes (
	appearance_id INTEGER NOT NULL,
	balance_change BOOLEAN,
	foreign key(appearance_id) references appearances(id)
);
create index if not exists appearance_balance_changes_id ON appearance_balance_changes (appearance_id);

CREATE TABLE IF NOT EXISTS incompatible_addresses (
	address VARCHAR(42) NOT NULL,
	appearances INT
);

CREATE VIEW IF NOT EXISTS view_appearances_with_providers AS SELECT
id,
address,
block_number,
transaction_index,
JSON_GROUP_ARRAY ( provider ) as providers
FROM (SELECT DISTINCT * FROM appearances)
GROUP BY address, block_number, transaction_index;

COMMIT;