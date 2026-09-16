UPDATE categories SET content_language = 'zh-Hant' WHERE content_language = 'en';
ALTER TABLE categories DROP CONSTRAINT categories_content_language_check;
ALTER TABLE categories
    ADD CONSTRAINT categories_content_language_check
    CHECK (content_language = 'zh-Hant');
