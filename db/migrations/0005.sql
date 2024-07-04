-- +migrate Down
DROP VIEW IF EXISTS vw_batch_num_gaps;
ALTER TABLE data_node.offchain_data DROP COLUMN IF EXISTS batch_num;

-- +migrate Up
ALTER TABLE data_node.offchain_data
    ADD COLUMN IF NOT EXISTS batch_num BIGINT NOT NULL DEFAULT 0;

-- Ensure batch_num is indexed for optimal performance
CREATE INDEX idx_batch_num ON data_node.offchain_data(batch_num);

-- Create a view to detect gaps in batch_num
CREATE VIEW vw_batch_num_gaps AS
WITH numbered_batches AS (
    SELECT
        batch_num,
        ROW_NUMBER() OVER (ORDER BY batch_num) AS row_number
    FROM data_node.offchain_data
)
SELECT
    nb1.batch_num AS current_batch_num,
    nb2.batch_num AS next_batch_num
FROM
    numbered_batches nb1
        LEFT JOIN numbered_batches nb2 ON nb1.row_number = nb2.row_number - 1
WHERE
    nb1.batch_num IS NOT NULL
  AND nb2.batch_num IS NOT NULL
  AND nb1.batch_num + 1 <> nb2.batch_num;

-- It triggers resync with an updated logic of all batches
UPDATE data_node.sync_tasks SET block = 0 WHERE task = 'L1';