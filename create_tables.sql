BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS download_status (
    address VARCHAR(42) PRIMARY KEY,
    provider TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS download_status_unique ON download_status (address, provider);

CREATE TABLE IF NOT EXISTS appearances (
	address VARCHAR(42) NOT NULL,
	block_number INT,
	transaction_index INT,
	provider TEXT
);
CREATE INDEX IF NOT EXISTS appearances_appearance ON appearances (address, block_number, transaction_index);
CREATE INDEX IF NOT EXISTS appearances_provider ON appearances (provider);

CREATE VIEW view_appearances_with_providers AS SELECT
address,
block_number,
transaction_index,
JSON_GROUP_ARRAY ( provider ) as providers
FROM appearances
GROUP BY address, block_number, transaction_index;

COMMIT;