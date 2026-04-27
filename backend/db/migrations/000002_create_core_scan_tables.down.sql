-- Migration: 000002 rollback
-- Drops core scan tables in reverse dependency order.

DROP TABLE IF EXISTS scans;
DROP TABLE IF EXISTS repository_snapshots;
DROP TABLE IF EXISTS repositories;
