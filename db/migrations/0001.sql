-- +migrate Down
DROP SCHEMA IF EXISTS data_node CASCADE;

-- +migrate Up
CREATE SCHEMA data_node;

CREATE TABLE data_node.offchain_data
(
    key VARCHAR PRIMARY KEY,
    value VARCHAR
);