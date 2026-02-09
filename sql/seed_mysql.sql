-- ===============================================
-- 快组测试数据填充脚本 (MySQL)
-- ===============================================

DELIMITER //

DROP PROCEDURE IF EXISTS seed_test_data //

CREATE PROCEDURE seed_test_data(IN num_users INT, IN num_projects INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE temp_user_id INT;
    DECLARE temp_project_id INT;
    DECLARE temp_school_id INT;
    DECLARE temp_major_id INT;
    DECLARE temp_product_id INT;
    DECLARE temp_actual_paid DECIMAL(10, 2);
    DECLARE counter INT DEFAULT 0;

    -- 1. 清理旧数据
    SET FOREIGN_KEY_CHECKS = 0;
    TRUNCATE TABLE project_application;
    TRUNCATE TABLE olive_branch_record;
    TRUNCATE TABLE `order`;
    TRUNCATE TABLE project;
    TRUNCATE TABLE talent_profile;
    TRUNCATE TABLE `user`;
    TRUNCATE TABLE major;
    TRUNCATE TABLE major_class;
    TRUNCATE TABLE school;
    TRUNCATE TABLE product;
    TRUNCATE TABLE feedback;
    TRUNCATE TABLE subscribe_config;
    SET FOREIGN_KEY_CHECKS = 1;

    -- 2. 插入学校
    INSERT INTO school (school_name, school_code) VALUES 
    ('清华大学', 'THU'), ('北京大学', 'PKU'), ('浙江大学', 'ZJU'), 
    ('复旦大学', 'FUDAN'), ('上海交通大学', 'SJTU'), ('武汉大学', 'WHU'),
    ('南京大学', 'NJU'), ('中山大学', 'SYSU'), ('四川大学', 'SCU'), ('华中科技大学', 'HUST');
    
    -- 3. 插入专业分类和专业
    INSERT INTO major_class (class_name) VALUES ('工学'), ('理学'), ('医学'), ('管理学'), ('经济学'), ('文学');
    
    INSERT INTO major (major_name, class_id) 
    SELECT name, mc.id
    FROM (
        SELECT '计算机科学与技术' AS name, '工学' AS cname UNION ALL
        SELECT '软件工程', '工学' UNION ALL
        SELECT '电子信息工程', '工学' UNION ALL
        SELECT '数学与应用数学', '理学' UNION ALL
        SELECT '物理学', '理学' UNION ALL
        SELECT '临床医学', '医学' UNION ALL
        SELECT '药学', '医学' UNION ALL
        SELECT '工商管理', '管理学' UNION ALL
        SELECT '市场营销', '管理学' UNION ALL
        SELECT '金融学', '经济学' UNION ALL
        SELECT '国际经济与贸易', '经济学' UNION ALL
        SELECT '汉语言文学', '文学' UNION ALL
        SELECT '英语', '文学'
    ) AS t
    JOIN major_class mc ON t.cname = mc.class_name;

    -- 4. 插入商品
    INSERT INTO product (name, type, description, price, config_json) VALUES
    ('橄榄枝礼包-10个', 1, '购买10个橄榄枝，用于邀请人才加入项目', 9.90, '{"olive_branch_count": 10}'),
    ('橄榄枝礼包-30个', 1, '购买30个橄榄枝，用于邀请人才加入项目', 24.90, '{"olive_branch_count": 30}'),
    ('橄榄枝礼包-50个', 1, '购买50个橄榄枝，用于邀请人才加入项目', 39.90, '{"olive_branch_count": 50}'),
    ('项目推广服务-7天', 2, '项目置顶推广7天，提升曝光度', 19.90, '{"promotion_days": 7}'),
    ('项目推广服务-30天', 2, '项目置顶推广30天，提升曝光度', 59.90, '{"promotion_days": 30}');

    -- 5. 插入用户
    SET i = 1;
    WHILE i <= num_users DO
        SELECT id INTO temp_school_id FROM school ORDER BY RAND() LIMIT 1;
        SELECT id INTO temp_major_id FROM major ORDER BY RAND() LIMIT 1;
        
        INSERT INTO `user` (
            openid, nickname, phone, email, 
            school_id, major_id, grade, 
            olive_branch_count, free_branch_used_today, 
            last_active_date, auth_status
        ) VALUES (
            CONCAT('openid_', i, '_', FLOOR(RAND() * 1000000)),
            CONCAT('用户', i),
            CONCAT('138', LPAD(i, 8, '0')),
            CONCAT('user', i, '@example.com'),
            temp_school_id,
            temp_major_id,
            FLOOR(RAND() * 4 + 1),
            10,
            FLOOR(RAND() * 3),
            DATE_SUB(CURRENT_DATE, INTERVAL FLOOR(RAND() * 30) DAY),
            ELT(FLOOR(RAND() * 3) + 1, 0, 1, 2)
        );
        SET temp_user_id = LAST_INSERT_ID();

        -- 为其中的 40% 的用户创建人才档案
        IF RAND() < 0.4 THEN
            INSERT INTO talent_profile (
                user_id, self_evaluation, skill_summary, 
                project_experience, mbti, status, is_public_contact
            ) VALUES (
                temp_user_id,
                '我是一个非常积极向上的学生，热爱编程，喜欢挑战，有良好的团队协作能力。',
                'Go, Python, React, SQL, Docker',
                '参加过多个校园开发项目，包括校园二手交易平台、在线答题系统等。',
                ELT(FLOOR(RAND() * 6) + 1, 'INTJ', 'ENFP', 'ISTP', 'ENTJ', 'INFJ', 'ENTP'),
                1,
                (RAND() > 0.5)
            );
        END IF;
        
        SET i = i + 1;
    END WHILE;

    -- 6. 插入项目
    SET i = 1;
    WHILE i <= num_projects DO
        SELECT id, school_id INTO temp_user_id, temp_school_id FROM `user` ORDER BY RAND() LIMIT 1;
        
        INSERT INTO project (
            creator_id, name, description, 
            school_id, direction, member_count, 
            status, promotion_status, view_count,
            is_cross_school, education_requirement, skill_requirement
        ) VALUES (
            temp_user_id,
            CONCAT('项目-', i, ': ', ELT(FLOOR(RAND() * 6) + 1, '校园社交App', '智能垃圾分类系统', '算法竞赛集训', '考研资料分享平台', '二次元社区', 'AI学习助手')),
            CONCAT('这是一个关于测试项目的详细描述，项目编号为 ', i, '。我们正在寻找志同道合的小伙伴一起完成这个项目，欢迎感兴趣的同学申请加入！'),
            temp_school_id,
            FLOOR(RAND() * 3 + 1),
            FLOOR(RAND() * 5 + 2),
            ELT(FLOOR(RAND() * 3) + 1, 0, 1, 2),
            0,
            FLOOR(RAND() * 200),
            FLOOR(RAND() + 1),
            FLOOR(RAND() + 1),
            '技能要求'
        );
        SET i = i + 1;
    END WHILE;

    -- 为大约 25% 的项目添加推广状态
    UPDATE project SET 
        promotion_status = 1,
        promotion_expire_time = DATE_ADD(CURRENT_TIMESTAMP, INTERVAL FLOOR(RAND() * 30 + 1) DAY)
    WHERE RAND() < 0.25;

    -- 7. 插入一些申请 (Project Application)
    SET i = 1;
    WHILE i <= (num_projects * 2) DO
        SELECT id INTO temp_project_id FROM project WHERE status = 1 ORDER BY RAND() LIMIT 1;
        SELECT id INTO temp_user_id FROM `user` ORDER BY RAND() LIMIT 1;
        
        -- 简单检查，避免自己申请自己的项目（MySQL中复杂的EXISTS检查在循环中较慢，这里尽量简化）
        IF temp_project_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM project WHERE id = temp_project_id AND creator_id = temp_user_id) THEN
            INSERT IGNORE INTO project_application (
                project_id, user_id, apply_reason, contact, status, reply_msg
            ) VALUES (
                temp_project_id, 
                temp_user_id, 
                '您好，我对这个项目非常感兴趣，我有相关的技术背景，希望能加入团队！',
                (SELECT phone FROM `user` WHERE id = temp_user_id),
                ELT(FLOOR(RAND() * 3) + 1, 0, 1, 2),
                CASE 
                    WHEN RAND() > 0.7 THEN '欢迎加入！请加我微信详细沟通。'
                    WHEN RAND() > 0.5 THEN '抱歉，目前团队已满，感谢关注！'
                    ELSE NULL
                END
            );
        END IF;
        SET i = i + 1;
    END WHILE;

    -- 8. 插入一些橄榄枝 (Olive Branch Record)
    SET i = 1;
    WHILE i <= num_projects DO
        SELECT id, creator_id INTO temp_project_id, temp_user_id FROM project WHERE status = 1 ORDER BY RAND() LIMIT 1;
        SELECT user_id INTO temp_school_id FROM talent_profile ORDER BY RAND() LIMIT 1; -- 用这个存一下receiver_id
        
        IF temp_project_id IS NOT NULL AND temp_user_id != temp_school_id THEN
            INSERT INTO olive_branch_record (
                sender_id, receiver_id, related_project_id, 
                type, cost_type, has_sms_notify, message, status
            ) VALUES (
                temp_user_id,
                temp_school_id,
                temp_project_id,
                2,
                ELT(FLOOR(RAND() * 2) + 1, 1, 2),
                (RAND() > 0.8),
                '同学你好，看了你的资料觉得很合适我们项目，欢迎加入！',
                ELT(FLOOR(RAND() * 4) + 1, 0, 1, 2, 3)
            );
        END IF;
        SET i = i + 1;
    END WHILE;

    -- 9. 插入一些订单 (Order)
    SET i = 1;
    WHILE i <= (num_users / 5) DO
        SELECT id INTO temp_user_id FROM `user` ORDER BY RAND() LIMIT 1;
        SELECT id, price INTO temp_product_id, temp_actual_paid FROM product ORDER BY RAND() LIMIT 1;

        INSERT INTO `order` (
            user_id, product_id, actual_paid, status, 
            wx_pay_no, pay_time
        ) VALUES (
            temp_user_id, 
            temp_product_id,
            temp_actual_paid,
            ELT(FLOOR(RAND() * 4) + 1, 0, 1, 2, 3),
            CASE WHEN RAND() > 0.3 THEN CONCAT('WX', FLOOR(RAND() * 1000000000000000)) ELSE NULL END,
            CASE WHEN RAND() > 0.3 THEN DATE_SUB(CURRENT_TIMESTAMP, INTERVAL FLOOR(RAND() * 30) DAY) ELSE NULL END
        );
        SET i = i + 1;
    END WHILE;

    -- 10. 插入一些意见反馈
    SET i = 1;
    WHILE i <= (num_users / 20) DO
        SELECT id INTO temp_user_id FROM `user` ORDER BY RAND() LIMIT 1;
        INSERT INTO feedback (user_id, content, status, admin_reply) VALUES (
            temp_user_id,
            ELT(FLOOR(RAND() * 4) + 1, '希望能增加更多的学校选择', '界面很美观，但是加载速度有点慢', '建议增加消息推送功能', '项目搜索功能可以优化一下'),
            ELT(FLOOR(RAND() * 2) + 1, 0, 1),
            CASE WHEN RAND() > 0.5 THEN '感谢您的反馈，我们会继续改进！' ELSE NULL END
        );
        SET i = i + 1;
    END WHILE;

END //

DELIMITER ;

-- 调用示例
CALL seed_test_data(100, 20);
