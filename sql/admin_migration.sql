-- 管理员用户表
CREATE TABLE IF NOT EXISTS `admin_user` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    `username` VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    `password_hash` VARCHAR(255) NOT NULL COMMENT 'bcrypt密码哈希',
    `nickname` VARCHAR(50) COMMENT '显示名称',
    `status` TINYINT DEFAULT 1 COMMENT '状态:1-启用,0-禁用',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员用户表';

CREATE INDEX idx_admin_user_username ON `admin_user`(`username`);

-- 默认管理员 (密码: admin123, 生产环境务必修改)
INSERT INTO `admin_user` (`username`, `password_hash`, `nickname`)
VALUES ('admin', '$2a$10$ctRtDR6rF7T0YIxzAuHAn.99OtkEPmgqDwgfzA.uB099UMg3ezLxu', '超级管理员');
