-- Rollback admin schema
DROP TABLE IF EXISTS email_templates CASCADE;
DROP TABLE IF EXISTS user_dismissed_announcements CASCADE;
DROP TABLE IF EXISTS announcement_banners CASCADE;
DROP TABLE IF EXISTS system_settings CASCADE;
DROP TABLE IF EXISTS user_feature_flags CASCADE;
DROP TABLE IF EXISTS feature_flags CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
