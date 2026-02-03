-- Migration 000015 Down: Remove Canadian provinces and territories

DELETE FROM woulder.mp_state_configs
WHERE region = 'Canada';
