-- +migrate Down
DROP TABLE IF EXISTS data_node.sync_info CASCADE;

-- +migrate Up
CREATE TABLE data_node.sync_info
(
    block       BIGINT PRIMARY KEY,
    processed   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- init with the genesis block
INSERT INTO data_node.sync_info (block) VALUES (0);
