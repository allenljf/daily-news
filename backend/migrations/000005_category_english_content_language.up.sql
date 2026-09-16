ALTER TABLE categories DROP CONSTRAINT categories_content_language_check;
ALTER TABLE categories
    ADD CONSTRAINT categories_content_language_check
    CHECK (content_language IN ('zh-Hant', 'en'));
