-- Hanya rancangan kueri pembacaan; belum dijalankan.
SELECT o.id, o.event_type, o.published_at IS NOT NULL AS di_publish,
       n.user_id, n.subject, n.sent_at
FROM outbox o LEFT JOIN notifications n ON n.outbox_id=o.id
ORDER BY o.id DESC LIMIT 10;

SELECT c.id,c.code,c.quota,c.enrolled_count,COUNT(e.id) AS actual_enrollments
FROM courses c LEFT JOIN enrollments e ON e.course_id=c.id
GROUP BY c.id ORDER BY c.id;
