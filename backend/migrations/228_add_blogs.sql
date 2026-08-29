-- 创建博客表（后台博客管理 + 前台博客展示）
-- 与 ent/schema/blog.go 保持一致
CREATE TABLE IF NOT EXISTS blogs (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    summary VARCHAR(1000) DEFAULT '',
    cover_image VARCHAR(1000) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    tags VARCHAR(255) DEFAULT '',
    published_at TIMESTAMPTZ DEFAULT NULL,
    created_by BIGINT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS blog_status ON blogs(status);
CREATE INDEX IF NOT EXISTS blog_created_at ON blogs(created_at);
CREATE INDEX IF NOT EXISTS blog_published_at ON blogs(published_at);

COMMENT ON TABLE blogs IS '博客文章';
COMMENT ON COLUMN blogs.title IS '博客标题';
COMMENT ON COLUMN blogs.content IS '博客内容（支持 Markdown）';
COMMENT ON COLUMN blogs.summary IS '博客摘要';
COMMENT ON COLUMN blogs.cover_image IS '封面图片 URL';
COMMENT ON COLUMN blogs.status IS '状态: draft, published';
COMMENT ON COLUMN blogs.tags IS '标签（逗号分隔）';
COMMENT ON COLUMN blogs.published_at IS '发布时间';
COMMENT ON COLUMN blogs.created_by IS '创建人用户ID';
COMMENT ON COLUMN blogs.deleted_at IS '软删除时间';
