ALTER TABLE categories
    ADD COLUMN content_language TEXT NOT NULL DEFAULT 'zh-Hant'
    CHECK (content_language = 'zh-Hant');
