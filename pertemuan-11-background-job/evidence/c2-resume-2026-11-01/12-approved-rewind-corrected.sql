BEGIN;
DO $guard$
DECLARE affected integer; c1_before jsonb; c1_after jsonb;
BEGIN
IF current_database() <> 'pertemuan11_test' OR current_user <> 'pbe11_baseline' THEN RAISE EXCEPTION 'Wrong target'; END IF;
IF NOT EXISTS(SELECT 1 FROM users WHERE id=61 AND username='mhsuji') THEN RAISE EXCEPTION 'Wrong user'; END IF;
SELECT jsonb_build_object('outbox',(SELECT jsonb_agg(to_jsonb(o) ORDER BY id) FROM outbox o WHERE id IN(21,23,24,25)), 'notifications',(SELECT jsonb_agg(to_jsonb(n) ORDER BY id) FROM notifications n WHERE outbox_id IN(21,23,24,25))) INTO c1_before;
IF (SELECT count(*) FROM outbox WHERE id IN(26,27,28) AND (payload->>'user_id')::int=61 AND (payload->>'course_id')::int IN(5,6,7))<>3 THEN RAISE EXCEPTION 'C2 event guard'; END IF;
DELETE FROM notifications WHERE outbox_id IN(26,27,28) AND user_id=61;
GET DIAGNOSTICS affected=ROW_COUNT; IF affected<>3 THEN RAISE EXCEPTION 'Notification guard'; END IF;
DELETE FROM outbox WHERE id IN(26,27,28) AND (payload->>'user_id')::int=61;
GET DIAGNOSTICS affected=ROW_COUNT; IF affected<>3 THEN RAISE EXCEPTION 'Outbox guard'; END IF;
DELETE FROM enrollments WHERE user_id=61 AND course_id IN(5,6,7);
GET DIAGNOSTICS affected=ROW_COUNT; IF affected<>3 THEN RAISE EXCEPTION 'Enrollment guard'; END IF;
UPDATE courses c SET enrolled_count=(SELECT count(*) FROM enrollments e WHERE e.course_id=c.id) WHERE c.id IN(5,6,7);
GET DIAGNOSTICS affected=ROW_COUNT; IF affected<>3 THEN RAISE EXCEPTION 'Course guard'; END IF;
SELECT jsonb_build_object('outbox',(SELECT jsonb_agg(to_jsonb(o) ORDER BY id) FROM outbox o WHERE id IN(21,23,24,25)), 'notifications',(SELECT jsonb_agg(to_jsonb(n) ORDER BY id) FROM notifications n WHERE outbox_id IN(21,23,24,25))) INTO c1_after;
IF c1_before IS DISTINCT FROM c1_after THEN RAISE EXCEPTION 'C1 changed'; END IF;
IF (SELECT count(*) FROM outbox WHERE id IN(21,23,24,25))<>4 OR (SELECT count(*) FROM notifications WHERE outbox_id IN(21,23,24,25))<>4 THEN RAISE EXCEPTION 'C1 missing'; END IF;
END $guard$;
COMMIT;

