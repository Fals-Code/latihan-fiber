-- D.2 ONLY. Apply after 008 and only with the d2 variant, never automatically.
BEGIN;
ALTER TABLE notifications DROP CONSTRAINT notifications_outbox_unique;
ALTER TABLE notifications ADD CONSTRAINT notifications_event_recipient_channel_unique
    UNIQUE (outbox_id, user_id, channel);
COMMIT;
