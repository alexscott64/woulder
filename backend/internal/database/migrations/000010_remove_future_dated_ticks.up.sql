-- Remove Mountain Project ticks with future dates (data quality issue)
-- Mountain Project sometimes has ticks with dates 7+ days in the future

-- Delete ticks that are more than 24 hours in the future
-- (Allow 24h buffer for clock skew)
DELETE FROM mp_ticks
WHERE climbed_at > NOW() + INTERVAL '24 hours';
