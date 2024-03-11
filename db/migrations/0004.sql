-- +migrate Down
DROP TABLE IF EXISTS data_node.unresolved_batches CASCADE;

-- +migrate Up
CREATE TABLE data_node.unresolved_batches
(
    num         BIGINT NOT NULL,
    hash        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (num, hash)
);
