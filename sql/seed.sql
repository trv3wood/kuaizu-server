-- ===============================================
-- 快组测试数据填充脚本 (PostgreSQL) (Deprecated)
-- ===============================================

CREATE OR REPLACE FUNCTION seed_test_data(num_users INTEGER DEFAULT 100, num_projects INTEGER DEFAULT 20)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
    school_id_list INTEGER[];
    major_id_list INTEGER[];
    user_id_list INTEGER[];
    talent_user_id_list INTEGER[];
    project_id_list INTEGER[];
    temp_id INTEGER;
    temp_user_id INTEGER;
    temp_project_id INTEGER;
    temp_receiver_id INTEGER;
BEGIN
    -- 1. 清理旧数据 (可选，按需开启)
    TRUNCATE TABLE 
        project_application, 
        olive_branch_record, 
        "order", 
        project, 
        talent_profile, 
        "user", 
        major, 
        major_class, 
        school, 
        product,
        feedback,
        subscribe_config
    CASCADE;

    -- 2. 插入学校
    INSERT INTO school (school_name, school_code) VALUES 
    ('清华大学', 'THU'), ('北京大学', 'PKU'), ('浙江大学', 'ZJU'), 
    ('复旦大学', 'FUDAN'), ('上海交通大学', 'SJTU'), ('武汉大学', 'WHU'),
    ('南京大学', 'NJU'), ('中山大学', 'SYSU'), ('四川大学', 'SCU'), ('华中科技大学', 'HUST');
    
    SELECT array_agg(id) INTO school_id_list FROM school;

    -- 3. 插入专业分类和专业
    INSERT INTO major_class (class_name) VALUES ('工学'), ('理学'), ('医学'), ('管理学'), ('经济学'), ('文学');
    
    INSERT INTO major (major_name, class_id) 
    SELECT name, (SELECT id FROM major_class WHERE class_name = cname)
    FROM (VALUES 
        ('计算机科学与技术', '工学'), ('软件工程', '工学'), ('电子信息工程', '工学'),
        ('数学与应用数学', '理学'), ('物理学', '理学'),
        ('临床医学', '医学'), ('药学', '医学'),
        ('工商管理', '管理学'), ('市场营销', '管理学'),
        ('金融学', '经济学'), ('国际经济与贸易', '经济学'),
        ('汉语言文学', '文学'), ('英语', '文学')
    ) AS t(name, cname);
    
    SELECT array_agg(id) INTO major_id_list FROM major;

    -- 4. 插入商品
    INSERT INTO product (name, type, description, price, config_json) VALUES
    ('橄榄枝礼包-10个', 1, '购买10个橄榄枝，用于邀请人才加入项目', 9.90, '{"olive_branch_count": 10}'),
    ('橄榄枝礼包-30个', 1, '购买30个橄榄枝，用于邀请人才加入项目', 24.90, '{"olive_branch_count": 30}'),
    ('橄榄枝礼包-50个', 1, '购买50个橄榄枝，用于邀请人才加入项目', 39.90, '{"olive_branch_count": 50}'),
    ('项目推广服务-7天', 2, '项目置顶推广7天，提升曝光度', 19.90, '{"promotion_days": 7}'),
    ('项目推广服务-30天', 2, '项目置顶推广30天，提升曝光度', 59.90, '{"promotion_days": 30}');

    -- 5. 插入用户
    FOR i IN 1..num_users LOOP
        INSERT INTO "user" (
            openid, nickname, phone, email, 
            school_id, major_id, grade, 
            olive_branch_count, free_branch_used_today, 
            last_active_date, auth_status
        ) VALUES (
            'openid_' || i || '_' || floor(random() * 1000000)::text,
            '用户' || i,
            '138' || LPAD(i::text, 8, '0'),
            'user' || i || '@example.com',
            school_id_list[floor(random() * array_length(school_id_list, 1) + 1)],
            major_id_list[floor(random() * array_length(major_id_list, 1) + 1)],
            floor(random() * 4 + 1), -- 1-4年级
            10, -- 橄榄枝余额
            floor(random() * 3), -- 今日已用免费次数
            CURRENT_DATE - floor(random() * 30)::integer, -- 最后活跃日期
            (ARRAY[0, 1, 2])[floor(random() * 3 + 1)] -- 认证状态: 0-未认证, 1-已认证, 2-认证失败
        ) RETURNING id INTO temp_user_id;
        user_id_list := array_append(user_id_list, temp_user_id);

        -- 为其中的 40% 的用户创建人才档案
        IF random() < 0.4 THEN
            INSERT INTO talent_profile (
                user_id, self_evaluation, skill_summary, 
                project_experience, mbti, status, is_public_contact
            ) VALUES (
                temp_user_id,
                '我是一个非常积极向上的学生，热爱编程，喜欢挑战，有良好的团队协作能力。',
                'Go, Python, React, SQL, Docker',
                '参加过多个校园开发项目，包括校园二手交易平台、在线答题系统等。',
                (ARRAY['INTJ', 'ENFP', 'ISTP', 'ENTJ', 'INFJ', 'ENTP'])[floor(random() * 6 + 1)],
                1, -- 上架
                (random() > 0.5) -- 是否公开联系方式
            );
            talent_user_id_list := array_append(talent_user_id_list, temp_user_id);
        END IF;
    END LOOP;

    -- 6. 插入项目
    FOR i IN 1..num_projects LOOP
        temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
        INSERT INTO project (
            creator_id, name, description, 
            school_id, direction, member_count, 
            status, promotion_status, view_count
        ) VALUES (
            temp_user_id,
            '项目-' || i || ': ' || (ARRAY['校园社交App', '智能垃圾分类系统', '算法竞赛集训', '考研资料分享平台', '二次元社区', 'AI学习助手'])[floor(random() * 6 + 1)],
            '这是一个关于测试项目的详细描述，项目编号为 ' || i || '。我们正在寻找志同道合的小伙伴一起完成这个项目，欢迎感兴趣的同学申请加入！',
            (SELECT school_id FROM "user" WHERE id = temp_user_id),
            floor(random() * 3 + 1), -- 1-落地, 2-比赛, 3-学习
            floor(random() * 5 + 2), -- 2-6人
            (ARRAY[0, 1, 2])[floor(random() * 3 + 1)], -- 状态: 0-待审核, 1-已通过, 2-已驳回
            0, -- 推广状态: 0-无
            floor(random() * 200) -- 浏览量
        ) RETURNING id INTO temp_project_id;
        project_id_list := array_append(project_id_list, temp_project_id);
    END LOOP;

    -- 为部分项目添加推广状态
    FOR i IN 1..(num_projects / 4) LOOP
        temp_project_id := project_id_list[floor(random() * array_length(project_id_list, 1) + 1)];
        UPDATE project SET 
            promotion_status = 1,
            promotion_expire_time = CURRENT_TIMESTAMP + (floor(random() * 30 + 1) || ' days')::interval
        WHERE id = temp_project_id;
    END LOOP;

    -- 7. 插入一些申请 (Project Application)
    FOR i IN 1..(num_projects * 3) LOOP
        temp_project_id := project_id_list[floor(random() * array_length(project_id_list, 1) + 1)];
        temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
        
        -- 避免重复申请和自己申请自己的项目，且只申请已通过的项目
        IF NOT EXISTS (SELECT 1 FROM project_application WHERE project_id = temp_project_id AND user_id = temp_user_id)
           AND NOT EXISTS (SELECT 1 FROM project WHERE id = temp_project_id AND creator_id = temp_user_id)
           AND EXISTS (SELECT 1 FROM project WHERE id = temp_project_id AND status = 1) THEN
            INSERT INTO project_application (
                project_id, user_id, apply_reason, contact, status, reply_msg
            ) VALUES (
                temp_project_id, 
                temp_user_id, 
                '您好，我对这个项目非常感兴趣，我有相关的技术背景，希望能加入团队！',
                (SELECT phone FROM "user" WHERE id = temp_user_id),
                (ARRAY[0, 1, 2])[floor(random() * 3 + 1)], -- 状态: 0-待审核, 1-已通过, 2-已拒绝
                CASE 
                    WHEN random() > 0.7 THEN '欢迎加入！请加我微信详细沟通。'
                    WHEN random() > 0.5 THEN '抱歉，目前团队已满，感谢关注！'
                    ELSE NULL
                END
            );
        END IF;
    END LOOP;

    -- 8. 插入一些橄榄枝 (Olive Branch Record)
    IF array_length(talent_user_id_list, 1) > 0 THEN
        FOR i IN 1..(num_projects) LOOP
            temp_project_id := project_id_list[floor(random() * array_length(project_id_list, 1) + 1)];
            temp_receiver_id := talent_user_id_list[floor(random() * array_length(talent_user_id_list, 1) + 1)];
            
            -- 获取项目创建者
            SELECT creator_id INTO temp_user_id FROM project WHERE id = temp_project_id;
            
            -- 避免发送给自己，且确保项目已通过
            IF temp_user_id != temp_receiver_id 
               AND EXISTS (SELECT 1 FROM project WHERE id = temp_project_id AND status = 1) THEN
                -- 插入橄榄枝记录（类型1-人才互联，类型2-项目邀请）
                INSERT INTO olive_branch_record (
                    sender_id, receiver_id, related_project_id, 
                    type, cost_type, has_sms_notify, message, status
                ) VALUES (
                    temp_user_id,
                    temp_receiver_id,
                    temp_project_id,
                    2, -- 类型2: 项目邀请
                    (ARRAY[1, 2])[floor(random() * 2 + 1)], -- 消耗类型: 1-免费额度, 2-付费额度
                    (random() > 0.8), -- 是否购买短信通知
                    '同学你好，看了你的资料觉得很合适我们项目，欢迎加入！',
                    (ARRAY[0, 1, 2, 3])[floor(random() * 4 + 1)] -- 状态: 0-待处理, 1-已接受, 2-已拒绝, 3-已忽略
                );
            END IF;
        END LOOP;
    END IF;

    -- 插入一些人才互联的橄榄枝
    IF array_length(talent_user_id_list, 1) > 1 THEN
        FOR i IN 1..(num_users / 10) LOOP
            temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
            temp_receiver_id := talent_user_id_list[floor(random() * array_length(talent_user_id_list, 1) + 1)];
            
            IF temp_user_id != temp_receiver_id THEN
                INSERT INTO olive_branch_record (
                    sender_id, receiver_id, related_project_id,
                    type, cost_type, has_sms_notify, message, status
                ) VALUES (
                    temp_user_id,
                    temp_receiver_id,
                    NULL, -- 人才互联不关联项目
                    1, -- 类型1: 人才互联
                    (ARRAY[1, 2])[floor(random() * 2 + 1)],
                    FALSE,
                    '你好，看到你的技能很匹配我的需求，可以交个朋友吗？',
                    (ARRAY[0, 1, 2, 3])[floor(random() * 4 + 1)]
                );
            END IF;
        END LOOP;
    END IF;

    -- 9. 插入一些订单 (Order)
    DECLARE
        product_id_list INTEGER[];
        temp_product_id INTEGER;
    BEGIN
        SELECT array_agg(id) INTO product_id_list FROM product;

        FOR i IN 1..(num_users / 5) LOOP
            temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
            temp_product_id := product_id_list[floor(random() * array_length(product_id_list, 1) + 1)];

            INSERT INTO "order" (
                user_id, product_id, actual_paid, status, 
                wx_pay_no, pay_time
            )
            SELECT 
                temp_user_id, 
                temp_product_id,
                price,
                (ARRAY[0, 1, 2, 3])[floor(random() * 4 + 1)], -- 状态: 0-待支付, 1-已支付, 2-已取消, 3-已退款
                CASE 
                    WHEN random() > 0.3 THEN 'WX' || floor(random() * 1000000000000000)::text
                    ELSE NULL
                END,
                CASE 
                    WHEN random() > 0.3 THEN CURRENT_TIMESTAMP - (floor(random() * 30) || ' days')::interval
                    ELSE NULL
                END
            FROM product WHERE id = temp_product_id;
        END LOOP;
    END;

    -- 10. 插入一些意见反馈
    FOR i IN 1..(num_users / 20) LOOP
        temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
        INSERT INTO feedback (user_id, content, status, admin_reply) VALUES (
            temp_user_id,
            (ARRAY[
                '希望能增加更多的学校选择',
                '界面很美观，但是加载速度有点慢',
                '建议增加消息推送功能',
                '项目搜索功能可以优化一下'
            ])[floor(random() * 4 + 1)],
            (ARRAY[0, 1])[floor(random() * 2 + 1)], -- 0-待处理, 1-已处理
            CASE 
                WHEN random() > 0.5 THEN '感谢您的反馈，我们会继续改进！'
                ELSE NULL
            END
        );
    END LOOP;

    RAISE NOTICE 'Seed completed: % users (% with talent profiles), % projects created.', 
        num_users, array_length(talent_user_id_list, 1), num_projects;
END;
$$ LANGUAGE plpgsql;

-- 使用示例:
-- SELECT seed_test_data(50, 10);  -- 生成50个用户，10个项目
-- SELECT seed_test_data(100, 20); -- 生成100个用户，20个项目（默认值）
SELECT seed_test_data(100, 20);
