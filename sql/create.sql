-- ===============================================
-- 快组数据库初始化脚本 (PostgreSQL) (Deprecated)
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
    free_branch_used_today INTEGER DEFAULT 0,
    last_active_date DATE,
    auth_status INTEGER DEFAULT 0,
    auth_img_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 核心业务：人才库 =================

-- 人才档案表
CREATE TABLE talent_profile (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    self_evaluation TEXT,
    skill_summary TEXT,
    project_experience TEXT,
    mbti VARCHAR(10),
    status INTEGER DEFAULT 1,
    is_public_contact BOOLEAN DEFAULT FALSE,
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
    status INTEGER DEFAULT 0,
    promotion_status INTEGER DEFAULT 0,
    promotion_expire_time TIMESTAMP,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 项目申请表
CREATE TABLE project_application (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    apply_reason TEXT,
    contact TEXT,
    status INTEGER DEFAULT 0,
    reply_msg TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, user_id)
);

-- 橄榄枝表
CREATE TABLE olive_branch_record (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    receiver_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    related_project_id INTEGER REFERENCES project(id) ON DELETE CASCADE,
    type INTEGER NOT NULL,
    cost_type INTEGER NOT NULL,
    has_sms_notify BOOLEAN DEFAULT FALSE,
    message TEXT,
    status INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 增值服务 =================

-- 商品表
CREATE TABLE product (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type INTEGER NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    config_json TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 订单表
CREATE TABLE "order" (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES product(id) ON DELETE RESTRICT,
    actual_paid DECIMAL(10, 2) NOT NULL,
    status INTEGER DEFAULT 0,
    wx_pay_no VARCHAR(100),
    pay_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 意见反馈表
CREATE TABLE feedback (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    contact_image TEXT,
    status INTEGER DEFAULT 0,
    admin_reply TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 消息订阅配置表
CREATE TABLE subscribe_config (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    target_type VARCHAR(50) NOT NULL,
    filter_json TEXT,
    email VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ================= 索引优化 =================

-- 用户表索引
CREATE INDEX idx_user_openid ON "user"(openid);
CREATE INDEX idx_user_school ON "user"(school_id);
CREATE INDEX idx_user_major ON "user"(major_id);

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
CREATE INDEX idx_olive_sender ON olive_branch_record(sender_id);
CREATE INDEX idx_olive_receiver ON olive_branch_record(receiver_id);
CREATE INDEX idx_olive_project ON olive_branch_record(related_project_id);
CREATE INDEX idx_olive_status ON olive_branch_record(status);

-- 订单表索引
CREATE INDEX idx_order_user ON "order"(user_id);
CREATE INDEX idx_order_product ON "order"(product_id);
CREATE INDEX idx_order_wx_pay_no ON "order"(wx_pay_no);
CREATE INDEX idx_order_status ON "order"(status);

-- 反馈表索引
CREATE INDEX idx_feedback_user ON feedback(user_id);
CREATE INDEX idx_feedback_status ON feedback(status);

-- 订阅配置表索引
CREATE INDEX idx_subscribe_user ON subscribe_config(user_id);
CREATE INDEX idx_subscribe_active ON subscribe_config(is_active);

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
COMMENT ON COLUMN "user".grade IS '年级';
COMMENT ON COLUMN "user".olive_branch_count IS '付费橄榄枝余额';
COMMENT ON COLUMN "user".free_branch_used_today IS '今日已用免费次数(每日重置)';
COMMENT ON COLUMN "user".last_active_date IS '最后活跃日期(用于重置免费次数)';
COMMENT ON COLUMN "user".auth_status IS '认证状态:0-未认证,1-已认证,2-认证失败';
COMMENT ON COLUMN "user".auth_img_url IS '学生证认证图';

-- talent_profile表注释
COMMENT ON TABLE talent_profile IS '人才档案表';
COMMENT ON COLUMN talent_profile.user_id IS '关联用户ID';
COMMENT ON COLUMN talent_profile.self_evaluation IS '自我评价';
COMMENT ON COLUMN talent_profile.skill_summary IS '技能标签';
COMMENT ON COLUMN talent_profile.project_experience IS '项目经历';
COMMENT ON COLUMN talent_profile.mbti IS 'MBTI性格类型';
COMMENT ON COLUMN talent_profile.status IS '状态:1-上架,0-下架';
COMMENT ON COLUMN talent_profile.is_public_contact IS '是否公开联系方式';

-- project表注释
COMMENT ON TABLE project IS '项目表';
COMMENT ON COLUMN project.creator_id IS '队长(用户ID)';
COMMENT ON COLUMN project.name IS '项目名称';
COMMENT ON COLUMN project.description IS '项目详情';
COMMENT ON COLUMN project.school_id IS '所属学校';
COMMENT ON COLUMN project.direction IS '项目方向:1-落地,2-比赛,3-学习';
COMMENT ON COLUMN project.member_count IS '需求人数';
COMMENT ON COLUMN project.status IS '审核状态:0-待审核,1-已通过,2-已驳回';
COMMENT ON COLUMN project.promotion_status IS '推广状态:0-无,1-推广中,2-已结束';
COMMENT ON COLUMN project.promotion_expire_time IS '推广结束时间';
COMMENT ON COLUMN project.view_count IS '浏览量';

-- project_application表注释
COMMENT ON TABLE project_application IS '项目申请表';
COMMENT ON COLUMN project_application.project_id IS '项目ID';
COMMENT ON COLUMN project_application.user_id IS '申请人';
COMMENT ON COLUMN project_application.apply_reason IS '申请理由/留言';
COMMENT ON COLUMN project_application.status IS '状态:0-待审核,1-已通过,2-已拒绝';
COMMENT ON COLUMN project_application.reply_msg IS '队长回复';
COMMENT ON COLUMN project_application.applied_at IS '申请时间';

-- olive_branch_record表注释
COMMENT ON TABLE olive_branch_record IS '橄榄枝/联系记录表';
COMMENT ON COLUMN olive_branch_record.sender_id IS '发起人ID';
COMMENT ON COLUMN olive_branch_record.receiver_id IS '接收人ID(人才或队长)';
COMMENT ON COLUMN olive_branch_record.related_project_id IS '关联项目ID(若是项目邀请)';
COMMENT ON COLUMN olive_branch_record.type IS '类型:1-人才互联,2-项目邀请';
COMMENT ON COLUMN olive_branch_record.cost_type IS '消耗类型:1-免费额度,2-付费额度';
COMMENT ON COLUMN olive_branch_record.has_sms_notify IS '是否购买短信通知';
COMMENT ON COLUMN olive_branch_record.message IS '邀请留言';
COMMENT ON COLUMN olive_branch_record.status IS '状态:0-待处理,1-已接受,2-已拒绝,3-已忽略';

-- product表注释
COMMENT ON TABLE product IS '商品表';
COMMENT ON COLUMN product.name IS '商品名称';
COMMENT ON COLUMN product.type IS '类型:1-虚拟币,2-服务权益';
COMMENT ON COLUMN product.description IS '商品描述';
COMMENT ON COLUMN product.price IS '商品价格';
COMMENT ON COLUMN product.config_json IS '配置参数(如增加多少个橄榄枝)';

-- order表注释
COMMENT ON TABLE "order" IS '订单表';
COMMENT ON COLUMN "order".user_id IS '用户ID';
COMMENT ON COLUMN "order".product_id IS '商品ID';
COMMENT ON COLUMN "order".actual_paid IS '实付金额';
COMMENT ON COLUMN "order".status IS '支付状态:0-待支付,1-已支付,2-已取消,3-已退款';
COMMENT ON COLUMN "order".wx_pay_no IS '微信支付单号';
COMMENT ON COLUMN "order".pay_time IS '支付时间';

-- feedback表注释
COMMENT ON TABLE feedback IS '意见反馈表';
COMMENT ON COLUMN feedback.user_id IS '用户ID';
COMMENT ON COLUMN feedback.content IS '反馈内容';
COMMENT ON COLUMN feedback.contact_image IS '图片凭证';
COMMENT ON COLUMN feedback.status IS '处理状态:0-待处理,1-已处理';
COMMENT ON COLUMN feedback.admin_reply IS '管理员回复';

-- subscribe_config表注释
COMMENT ON TABLE subscribe_config IS '消息订阅配置表';
COMMENT ON COLUMN subscribe_config.user_id IS '用户ID';
COMMENT ON COLUMN subscribe_config.target_type IS '订阅类型(新项目/审核结果/投递进度)';
COMMENT ON COLUMN subscribe_config.filter_json IS '过滤条件(如:只看某学校项目)';
COMMENT ON COLUMN subscribe_config.email IS '接收邮箱';
COMMENT ON COLUMN subscribe_config.is_active IS '是否开启';
