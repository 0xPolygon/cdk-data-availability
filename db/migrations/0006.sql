-- +migrate Down
-- Add the 'batch_num' column to 'offchain_data' table
ALTER TABLE data_node.offchain_data
    ADD COLUMN IF NOT EXISTS batch_num BIGINT NOT NULL DEFAULT 0;

-- Rename the 'missing_batches' table to 'unresolved_batches'
ALTER TABLE data_node.missing_batches RENAME TO data_node.unresolved_batches;

-- Create an index for the 'batch_num' column for better performance
CREATE INDEX IF NOT EXISTS idx_batch_num ON data_node.offchain_data(batch_num);

-- Reset the sync task for L1 to trigger resync
UPDATE data_node.sync_tasks SET block = 0 WHERE task = 'L1';

-- +migrate Up
-- Drop the 'batch_num' column from 'offchain_data' table
ALTER TABLE data_node.offchain_data DROP COLUMN batch_num;

-- Rename the 'unresolved_batches' table back to 'missing_batches'
ALTER TABLE data_node.unresolved_batches RENAME TO data_node.missing_batches;

-- Drop the index created on 'batch_num'
DROP INDEX IF EXISTS idx_batch_num;