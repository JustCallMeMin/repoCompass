-- Rollback: 000003 create analysis persistence tables

DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS metric_snapshots;
DROP TABLE IF EXISTS assessments;
DROP TABLE IF EXISTS recommendations;
DROP TABLE IF EXISTS finding_evidences;
DROP TABLE IF EXISTS findings;
DROP TABLE IF EXISTS analyzer_results;
DROP TABLE IF EXISTS rule_set_rules;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS rule_sets;
