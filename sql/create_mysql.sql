-- ===============================================
-- 快组数据库初始化脚本 (MySQL)
-- ===============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS kuaizu DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE kuaizu;

-- ================= 基础字典表 =================

-- 学校表
CREATE TABLE `school` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `school_name` VARCHAR(100) NOT NULL COMMENT '学校名称',
    `school_code` VARCHAR(50) UNIQUE COMMENT '学校代码',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学校字典表';

-- 专业大类表
CREATE TABLE `major_class` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `class_name` VARCHAR(50) NOT NULL COMMENT '专业大类名称',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='专业大类表';

-- 专业表
CREATE TABLE `major` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `major_name` VARCHAR(100) NOT NULL COMMENT '专业名称',
    `class_id` INT NOT NULL COMMENT '所属大类ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT `fk_major_class` FOREIGN KEY (`class_id`) REFERENCES `major_class` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='专业表';

-- ================= 用户中心 =================

-- 用户表
CREATE TABLE `user` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `openid` VARCHAR(100) UNIQUE NOT NULL COMMENT '微信OpenID',
    `nickname` VARCHAR(50) COMMENT '昵称',
    `phone` VARCHAR(20) COMMENT '手机号',
    `email` VARCHAR(100) COMMENT '邮箱',
    `school_id` INT COMMENT '学校ID',
    `major_id` INT COMMENT '专业ID',
    `grade` INT COMMENT '年级',
    `olive_branch_count` INT DEFAULT 0 COMMENT '付费橄榄枝余额',
    `free_branch_used_today` INT DEFAULT 0 COMMENT '今日已用免费次数(每日重置)',
    `last_active_date` DATE COMMENT '最后活跃日期(用于重置免费次数)',
    `auth_status` INT DEFAULT 0 COMMENT '认证状态:0-未认证,1-已认证,2-认证失败',
    `auth_img_url` TEXT COMMENT '学生证认证图',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    CONSTRAINT `fk_user_school` FOREIGN KEY (`school_id`) REFERENCES `school` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_user_major` FOREIGN KEY (`major_id`) REFERENCES `major` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ================= 核心业务：人才库 =================

-- 人才档案表
CREATE TABLE `talent_profile` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `user_id` INT UNIQUE NOT NULL COMMENT '关联用户ID',
    `self_evaluation` TEXT COMMENT '自我评价',
    `skill_summary` TEXT COMMENT '技能标签',
    `project_experience` TEXT COMMENT '项目经历',
    `mbti` VARCHAR(10) COMMENT 'MBTI性格类型',
    `status` INT DEFAULT 1 COMMENT '状态:1-上架,0-下架',
    `is_public_contact` TINYINT(1) DEFAULT 0 COMMENT '是否公开联系方式',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT `fk_talent_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='人才档案表';

-- ================= 核心业务：项目组队 =================

-- 项目表
CREATE TABLE `project` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `creator_id` INT NOT NULL COMMENT '队长(用户ID)',
    `name` VARCHAR(200) NOT NULL COMMENT '项目名称',
    `description` TEXT COMMENT '项目详情',
    `school_id` INT COMMENT '所属学校',
    `direction` INT COMMENT '项目方向:1-落地,2-比赛,3-学习',
    `member_count` INT COMMENT '需求人数',
    `status` INT DEFAULT 0 COMMENT '审核状态:0-待审核,1-已通过,2-已驳回',
    `promotion_status` INT DEFAULT 0 COMMENT '推广状态:0-无,1-推广中,2-已结束',
    `promotion_expire_time` TIMESTAMP NULL COMMENT '推广结束时间',
    `view_count` INT DEFAULT 0 COMMENT '浏览量',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `is_cross_school` TINYINT DEFAULT 1 COMMENT "是否跨校: 1-可以,2-不可以",
    `education_requirement` TINYINT DEFAULT 1 COMMENT "学历要求1-大专2-本科",
    `skill_requirement` TEXT COMMENT "技能要求",
    CONSTRAINT `fk_project_creator` FOREIGN KEY (`creator_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_project_school` FOREIGN KEY (`school_id`) REFERENCES `school` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目表';

-- 项目申请表
CREATE TABLE `project_application` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `project_id` INT NOT NULL COMMENT '项目ID',
    `user_id` INT NOT NULL COMMENT '申请人',
    `apply_reason` TEXT COMMENT '申请理由/留言',
    `contact` TEXT COMMENT '联系方式',
    `status` INT DEFAULT 0 COMMENT '状态:0-待审核,1-已通过,2-已拒绝',
    `reply_msg` TEXT COMMENT '队长回复',
    `applied_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_project_user` (`project_id`, `user_id`),
    CONSTRAINT `fk_app_project` FOREIGN KEY (`project_id`) REFERENCES `project` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_app_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='项目申请表';

-- 橄榄枝表
CREATE TABLE `olive_branch_record` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `sender_id` INT NOT NULL COMMENT '发起人ID',
    `receiver_id` INT NOT NULL COMMENT '接收人ID(人才或队长)',
    `related_project_id` INT COMMENT '关联项目ID(若是项目邀请)',
    `type` INT NOT NULL COMMENT '类型:1-人才互联,2-项目邀请',
    `cost_type` INT NOT NULL COMMENT '消耗类型:1-免费额度,2-付费额度',
    `has_sms_notify` TINYINT(1) DEFAULT 0 COMMENT '是否购买短信通知',
    `message` TEXT COMMENT '邀请留言',
    `status` INT DEFAULT 0 COMMENT '状态:0-待处理,1-已接受,2-已拒绝,3-已忽略',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT `fk_olive_sender` FOREIGN KEY (`sender_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_olive_receiver` FOREIGN KEY (`receiver_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_olive_project` FOREIGN KEY (`related_project_id`) REFERENCES `project` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='橄榄枝/联系记录表';

-- ================= 增值服务 =================

-- 商品表
CREATE TABLE `product` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `name` VARCHAR(100) NOT NULL COMMENT '商品名称',
    `type` INT NOT NULL COMMENT '类型:1-虚拟币,2-服务权益',
    `description` TEXT COMMENT '商品描述',
    `price` DECIMAL(10, 2) NOT NULL COMMENT '商品价格',
    `config_json` TEXT COMMENT '配置参数(如增加多少个橄榄枝)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';

-- 订单表
CREATE TABLE `order` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `user_id` INT NOT NULL COMMENT '用户ID',
    `product_id` INT NOT NULL COMMENT '商品ID',
    `actual_paid` DECIMAL(10, 2) NOT NULL COMMENT '实付金额',
    `status` INT DEFAULT 0 COMMENT '支付状态:0-待支付,1-已支付,2-已取消,3-已退款',
    `wx_pay_no` VARCHAR(100) COMMENT '微信支付单号',
    `pay_time` TIMESTAMP NULL COMMENT '支付时间',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT `fk_order_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_order_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';

-- 意见反馈表
CREATE TABLE `feedback` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `user_id` INT NOT NULL COMMENT '用户ID',
    `content` TEXT NOT NULL COMMENT '反馈内容',
    `contact_image` TEXT COMMENT '图片凭证',
    `status` INT DEFAULT 0 COMMENT '处理状态:0-待处理,1-已处理',
    `admin_reply` TEXT COMMENT '管理员回复',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT `fk_feedback_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='意见反馈表';

-- 消息订阅配置表
CREATE TABLE `subscribe_config` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `user_id` INT NOT NULL COMMENT '用户ID',
    `target_type` VARCHAR(50) NOT NULL COMMENT '订阅类型(新项目/审核结果/投递进度)',
    `filter_json` TEXT COMMENT '过滤条件(如:只看某学校项目)',
    `email` VARCHAR(100) COMMENT '接收邮箱',
    `is_active` TINYINT(1) DEFAULT 1 COMMENT '是否开启',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    CONSTRAINT `fk_sub_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息订阅配置表';

-- ================= 索引优化 =================

-- 相关外键已自动创建索引，以下为其他常用查询索引
CREATE INDEX idx_user_openid ON `user`(`openid`);
CREATE INDEX idx_user_school ON `user`(`school_id`);
CREATE INDEX idx_user_major ON `user`(`major_id`);

CREATE INDEX idx_talent_user ON `talent_profile`(`user_id`);
CREATE INDEX idx_talent_status ON `talent_profile`(`status`);

CREATE INDEX idx_project_creator ON `project`(`creator_id`);
CREATE INDEX idx_project_school ON `project`(`school_id`);
CREATE INDEX idx_project_status ON `project`(`status`);
CREATE INDEX idx_project_created ON `project`(`created_at`);

CREATE INDEX idx_application_project ON `project_application`(`project_id`);
CREATE INDEX idx_application_user ON `project_application`(`user_id`);
CREATE INDEX idx_application_status ON `project_application`(`status`);

CREATE INDEX idx_olive_sender ON `olive_branch_record`(`sender_id`);
CREATE INDEX idx_olive_receiver ON `olive_branch_record`(`receiver_id`);
CREATE INDEX idx_olive_project ON `olive_branch_record`(`related_project_id`);
CREATE INDEX idx_olive_status ON `olive_branch_record`(`status`);

CREATE INDEX idx_order_user ON `order`(`user_id`);
CREATE INDEX idx_order_product ON `order`(`product_id`);
CREATE INDEX idx_order_wx_pay_no ON `order`(`wx_pay_no`);
CREATE INDEX idx_order_status ON `order`(`status`);

CREATE INDEX idx_feedback_user ON `feedback`(`user_id`);
CREATE INDEX idx_feedback_status ON `feedback`(`status`);

CREATE INDEX idx_subscribe_user ON `subscribe_config`(`user_id`);
CREATE INDEX idx_subscribe_active ON `subscribe_config`(`is_active`);
