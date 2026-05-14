-- Cleanup migration: remove legacy kaya_mp_route_matches rows that point to
-- non-boulder MP routes. These were created before the matcher's discipline
-- guard was tightened (see backend/cmd/match_kaya_mp/main.go isCompatibleMatch).
--
-- Kaya is bouldering-only, so any match where the MP route is not a boulder
-- (or is ice/mixed/snow/alpine) is incorrect and must be deleted. The matcher
-- uses ON CONFLICT DO UPDATE which never deletes stale rows, so this one-time
-- cleanup is required in addition to the code-level fix.
--
-- The migration runner wraps each migration in its own transaction, so an
-- explicit BEGIN/COMMIT is not used here.

DELETE FROM woulder.kaya_mp_route_matches m
USING woulder.mp_routes r
WHERE m.mp_route_id = r.mp_route_id
  AND (r.route_type !~* 'boulder'
       OR r.route_type ~* '(ice|mixed|snow|alpine)');
