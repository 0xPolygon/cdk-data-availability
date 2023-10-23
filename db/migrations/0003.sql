-- +migrate Down
DROP TABLE IF EXISTS data_node.sync_tasks CASCADE;

-- +migrate Up
CREATE TABLE data_node.sync_tasks
(
    task        VARCHAR PRIMARY KEY,
    block       BIGINT NOT NULL,
    processed   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- transfer data from old table to new
INSERT INTO data_node.sync_tasks (task, block)
    SELECT 'L1', MAX(block) FROM data_node.sync_info;

DROP TABLE IF EXISTS data_node.sync_info CASCADE;
