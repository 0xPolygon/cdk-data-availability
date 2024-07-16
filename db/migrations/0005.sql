-- +migrate Down
ALTER TABLE data_node.offchain_data DROP COLUMN IF EXISTS batch_num;

-- +migrate Up
ALTER TABLE data_node.offchain_data
    ADD COLUMN IF NOT EXISTS batch_num BIGINT NOT NULL DEFAULT 0;

-- Ensure batch_num is indexed for optimal performance
CREATE INDEX idx_batch_num ON data_node.offchain_data(batch_num);

-- It triggers resync with an updated logic of all batches
UPDATE data_node.sync_tasks SET block = 0 WHERE task = 'L1';