-- ===============================================
-- 快组数据库初始化脚本 (PostgreSQL)
-- ===============================================

-- 创建数据库
-- CREATE DATABASE kuaizu;

-- 连接到数据库 (需要手动执行 \c kuaizu)

-- ================= 基础字典表 =================

-- 学校表
CREATE TABLE school (
    id SERIAL PRIMARY KEY,
    school_name VARCHAR(100) NOT NULL,
    school_code VARCHAR(50) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 专业大类表
CREATE TABLE major_class (
    id SERIAL PRIMARY KEY,
    class_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 专业表
CREATE TABLE major (
    id SERIAL PRIMARY KEY,
    major_name VARCHAR(100) NOT NULL,
    class_id INTEGER NOT NULL REFERENCES major_class(id) ON DELETE RESTRICT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 用户中心 =================

-- 用户表
CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(100) UNIQUE NOT NULL,
    nickname VARCHAR(50),
    phone VARCHAR(20),
    email VARCHAR(100),
    school_id INTEGER REFERENCES school(id) ON DELETE SET NULL,
    major_id INTEGER REFERENCES major(id) ON DELETE SET NULL,
    grade INTEGER,
    olive_branch_count INTEGER DEFAULT 0,
    student_img_url TEXT,
    auth_status INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 简历表
CREATE TABLE resume (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    resume_name VARCHAR(100),
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 核心业务：人才库 =================

-- 人才档案表
-- 说明：用户需要单独创建人才档案才能进入人才库被搜索
CREATE TABLE talent_profile (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    self_evaluation TEXT,
    skill_summary TEXT,
    project_experience TEXT,
    mbti VARCHAR(10),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 核心业务：项目组队 =================

-- 项目表
CREATE TABLE project (
    id SERIAL PRIMARY KEY,
    creator_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    school_id INTEGER REFERENCES school(id) ON DELETE SET NULL,
    direction INTEGER,
    member_count INTEGER,
    education_req INTEGER,
    is_cross_school BOOLEAN DEFAULT FALSE,
    status INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 项目申请表
-- 流程一：用户申请加入项目
CREATE TABLE project_application (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    status INTEGER DEFAULT 0,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);

-- 橄榄枝表
-- 流程二：队长向人才抛出橄榄枝
CREATE TABLE olive_branch (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    talent_id INTEGER NOT NULL REFERENCES talent_profile(id) ON DELETE CASCADE,
    message TEXT,
    status INTEGER DEFAULT 0,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 增值服务 =================

-- 订单表
CREATE TABLE "order" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status INTEGER DEFAULT 0,
    trade_no VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 索引优化 =================

-- 用户表索引
CREATE INDEX idx_user_openid ON "user"(openid);
CREATE INDEX idx_user_school ON "user"(school_id);
CREATE INDEX idx_user_major ON "user"(major_id);

-- 简历表索引
CREATE INDEX idx_resume_user ON resume(user_id);

-- 人才档案表索引
CREATE INDEX idx_talent_user ON talent_profile(user_id);
CREATE INDEX idx_talent_status ON talent_profile(status);

-- 项目表索引
CREATE INDEX idx_project_creator ON project(creator_id);
CREATE INDEX idx_project_school ON project(school_id);
CREATE INDEX idx_project_status ON project(status);
CREATE INDEX idx_project_created ON project(created_at);

-- 项目申请表索引
CREATE INDEX idx_application_project ON project_application(project_id);
CREATE INDEX idx_application_user ON project_application(user_id);
CREATE INDEX idx_application_status ON project_application(status);

-- 橄榄枝表索引
CREATE INDEX idx_olive_project ON olive_branch(project_id);
CREATE INDEX idx_olive_talent ON olive_branch(talent_id);
CREATE INDEX idx_olive_status ON olive_branch(status);

-- 订单表索引
CREATE INDEX idx_order_user ON "order"(user_id);
CREATE INDEX idx_order_trade_no ON "order"(trade_no);
CREATE INDEX idx_order_status ON "order"(status);

-- ================= 注释说明 =================
-- 注意：PostgreSQL 不支持 COMMENT 关键字内联在列定义中
-- 以下使用 COMMENT ON 语句添加列注释

-- school表注释
COMMENT ON TABLE school IS '学校字典表';
COMMENT ON COLUMN school.school_name IS '学校名称';
COMMENT ON COLUMN school.school_code IS '学校代码';

-- major_class表注释
COMMENT ON TABLE major_class IS '专业大类表';
COMMENT ON COLUMN major_class.class_name IS '专业大类名称';

-- major表注释
COMMENT ON TABLE major IS '专业表';
COMMENT ON COLUMN major.major_name IS '专业名称';
COMMENT ON COLUMN major.class_id IS '所属大类ID';

-- user表注释
COMMENT ON TABLE "user" IS '用户表';
COMMENT ON COLUMN "user".openid IS '微信OpenID';
COMMENT ON COLUMN "user".nickname IS '昵称';
COMMENT ON COLUMN "user".phone IS '手机号';
COMMENT ON COLUMN "user".email IS '邮箱';
COMMENT ON COLUMN "user".school_id IS '学校ID';
COMMENT ON COLUMN "user".major_id IS '专业ID';
COMMENT ON COLUMN "user".grade IS '年级字典值';
COMMENT ON COLUMN "user".olive_branch_count IS '剩余橄榄枝数量';
COMMENT ON COLUMN "user".student_img_url IS '学生证照片';
COMMENT ON COLUMN "user".auth_status IS '认证状态:0-未认证,1-审核中,2-已认证,3-认证失败';

-- resume表注释
COMMENT ON TABLE resume IS '简历表';
COMMENT ON COLUMN resume.user_id IS '用户ID';
COMMENT ON COLUMN resume.resume_name IS '简历名称';
COMMENT ON COLUMN resume.content IS '简历内容/JSON';

-- talent_profile表注释
COMMENT ON TABLE talent_profile IS '人才档案表';
COMMENT ON COLUMN talent_profile.user_id IS '关联用户ID';
COMMENT ON COLUMN talent_profile.self_evaluation IS '自我评价';
COMMENT ON COLUMN talent_profile.skill_summary IS '技能标签';
COMMENT ON COLUMN talent_profile.project_experience IS '项目经历';
COMMENT ON COLUMN talent_profile.mbti IS 'MBTI性格类型';
COMMENT ON COLUMN talent_profile.status IS '状态:active-上架,inactive-下架';

-- project表注释
COMMENT ON TABLE project IS '项目表';
COMMENT ON COLUMN project.creator_id IS '队长(用户ID)';
COMMENT ON COLUMN project.name IS '项目名称';
COMMENT ON COLUMN project.description IS '项目详情';
COMMENT ON COLUMN project.school_id IS '所属学校';
COMMENT ON COLUMN project.direction IS '项目方向:1-落地,2-比赛,3-学习';
COMMENT ON COLUMN project.member_count IS '需求人数';
COMMENT ON COLUMN project.education_req IS '学历要求:1-不限,2-专科,3-本科,4-硕士,5-博士';
COMMENT ON COLUMN project.is_cross_school IS '是否跨校';
COMMENT ON COLUMN project.status IS '审核状态:0-待审核,1-已通过,2-已驳回';

-- project_application表注释
COMMENT ON TABLE project_application IS '项目申请表';
COMMENT ON COLUMN project_application.project_id IS '项目ID';
COMMENT ON COLUMN project_application.user_id IS '申请人';
COMMENT ON COLUMN project_application.status IS '状态:0-待审核,1-已通过,2-已拒绝';

-- olive_branch表注释
COMMENT ON TABLE olive_branch IS '橄榄枝邀请表';
COMMENT ON COLUMN olive_branch.project_id IS '项目ID';
COMMENT ON COLUMN olive_branch.talent_id IS '接收人(人才ID)';
COMMENT ON COLUMN olive_branch.message IS '邀请留言';
COMMENT ON COLUMN olive_branch.status IS '状态:0-待处理,1-已接受,2-已拒绝';

-- order表注释
COMMENT ON TABLE "order" IS '订单表';
COMMENT ON COLUMN "order".user_id IS '用户ID';
COMMENT ON COLUMN "order".amount IS '金额';
COMMENT ON COLUMN "order".type IS '类型:olive_branch-购买橄榄枝,email_promotion-邮件推广';
COMMENT ON COLUMN "order".status IS '支付状态:0-待支付,1-已支付,2-已取消,3-已退款';
COMMENT ON COLUMN "order".trade_no IS '微信支付单号';
