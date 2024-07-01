-- +migrate Down
ALTER TABLE data_node.offchain_data DROP COLUMN IF EXISTS batch_num;

DELETE FROM data_node.sync_tasks WHERE task = 'L1_BATCH_NUM';

-- +migrate Up
INSERT INTO data_node.sync_tasks (task, block) VALUES ('L1_BATCH_NUM', 0);

ALTER TABLE data_node.offchain_data
    ADD COLUMN IF NOT EXISTS batch_num BIGINT NOT NULL DEFAULT 0;