-- ===============================================
-- 快组测试数据填充脚本 (PostgreSQL)
-- ===============================================

CREATE OR REPLACE FUNCTION seed_test_data(num_users INTEGER DEFAULT 100, num_projects INTEGER DEFAULT 20)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
    school_id_list INTEGER[];
    major_id_list INTEGER[];
    user_id_list INTEGER[];
    talent_id_list INTEGER[];
    project_id_list INTEGER[];
    temp_id INTEGER;
    temp_user_id INTEGER;
    temp_talent_id INTEGER;
    temp_project_id INTEGER;
BEGIN
    -- 1. 清理旧数据 (可选，按需开启)
    TRUNCATE TABLE project_application, olive_branch, "order", project, talent_profile, resume, "user", major, major_class, school, product CASCADE;

    -- 2. 插入学校
    IF NOT EXISTS (SELECT 1 FROM school LIMIT 1) THEN
        INSERT INTO school (school_name, school_code) VALUES 
        ('清华大学', 'THU'), ('北京大学', 'PKU'), ('浙江大学', 'ZJU'), 
        ('复旦大学', 'FUDAN'), ('上海交通大学', 'SJTU'), ('武汉大学', 'WHU'),
        ('南京大学', 'NJU'), ('中山大学', 'SYSU'), ('四川大学', 'SCU'), ('华中科技大学', 'HUST');
    END IF;
    SELECT array_agg(id) INTO school_id_list FROM school;

    -- 3. 插入专业分类和专业
    IF NOT EXISTS (SELECT 1 FROM major_class LIMIT 1) THEN
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
    END IF;
    SELECT array_agg(id) INTO major_id_list FROM major;

    -- 4. 插入商品
    IF NOT EXISTS (SELECT 1 FROM product LIMIT 1) THEN
        INSERT INTO product (name, description, price, available_amount) VALUES
        ('橄榄枝礼包-10个', '购买10个橄榄枝，用于邀请人才加入项目', 9.90, 10),
        ('橄榄枝礼包-30个', '购买30个橄榄枝，用于邀请人才加入项目', 24.90, 30),
        ('橄榄枝礼包-50个', '购买50个橄榄枝，用于邀请人才加入项目', 39.90, 50),
        ('邮件推广服务', '通过邮件向目标人才推广项目', 19.90, 1);
    END IF;

    -- 5. 插入用户
    FOR i IN 1..num_users LOOP
        INSERT INTO "user" (
            openid, nickname, phone, email, 
            school_id, major_id, grade, 
            olive_branch_count, auth_status
        ) VALUES (
            'openid_' || i || '_' || floor(random() * 1000000)::text,
            '用户' || i,
            '138' || LPAD(i::text, 8, '0'),
            'user' || i || '@example.com',
            school_id_list[floor(random() * array_length(school_id_list, 1) + 1)],
            major_id_list[floor(random() * array_length(major_id_list, 1) + 1)],
            floor(random() * 4 + 1), -- 1-4年级
            10,
            2 -- 已认证
        ) RETURNING id INTO temp_user_id;
        user_id_list := array_append(user_id_list, temp_user_id);

        -- 为其中的 40% 的用户创建简历和人才档案
        IF random() < 0.4 THEN
            INSERT INTO resume (user_id, resume_name, content) 
            VALUES (temp_user_id, '我的简历', '{"summary": "这是用户' || i || '的个人简历内容"}');

            INSERT INTO talent_profile (
                user_id, self_evaluation, skill_summary, 
                project_experience, mbti, status
            ) VALUES (
                temp_user_id,
                '我是一个非常积极向上的学生',
                'Go, Python, React, SQL',
                '参加过多个校园开发项目',
                (ARRAY['INTJ', 'ENFP', 'ISTP', 'ENTJ'])[floor(random() * 4 + 1)],
                'active'
            ) RETURNING id INTO temp_talent_id;
            talent_id_list := array_append(talent_id_list, temp_talent_id);
        END IF;
    END LOOP;

    -- 6. 插入项目
    FOR i IN 1..num_projects LOOP
        temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
        INSERT INTO project (
            creator_id, name, description, 
            school_id, direction, member_count, 
            education_req, is_cross_school, status
        ) VALUES (
            temp_user_id,
            '项目-' || i || ': ' || (ARRAY['校园社交App', '智能垃圾分类', '算法竞赛集训', '考研资料分享', '二次元社区'])[floor(random() * 5 + 1)],
            '这是一个关于测试项目的详细描述，编号为 ' || i,
            (SELECT school_id FROM "user" WHERE id = temp_user_id),
            floor(random() * 3 + 1), -- 1-落地, 2-比赛, 3-学习
            floor(random() * 5 + 2), -- 2-6人
            floor(random() * 5 + 1), -- 学历要求
            (random() > 0.5),
            1 -- 已通过
        ) RETURNING id INTO temp_project_id;
        project_id_list := array_append(project_id_list, temp_project_id);
    END LOOP;

    -- 7. 插入一些申请 (Project Application)
    FOR i IN 1..(num_projects * 2) LOOP
        temp_project_id := project_id_list[floor(random() * array_length(project_id_list, 1) + 1)];
        temp_user_id := user_id_list[floor(random() * array_length(user_id_list, 1) + 1)];
        
        -- 避免重复申请和自己申请自己的项目
        IF NOT EXISTS (SELECT 1 FROM project_application WHERE project_id = temp_project_id AND user_id = temp_user_id)
           AND NOT EXISTS (SELECT 1 FROM project WHERE id = temp_project_id AND creator_id = temp_user_id) THEN
            INSERT INTO project_application (project_id, user_id, status)
            VALUES (temp_project_id, temp_user_id, floor(random() * 3)::integer);
        END IF;
    END LOOP;

    -- 8. 插入一些橄榄枝 (Olive Branch)
    IF array_length(talent_id_list, 1) > 0 THEN
        FOR i IN 1..(num_projects) LOOP
            temp_project_id := project_id_list[floor(random() * array_length(project_id_list, 1) + 1)];
            temp_talent_id := talent_id_list[floor(random() * array_length(talent_id_list, 1) + 1)];
            
            -- 避免重复发送和发送给自己
            IF NOT EXISTS (SELECT 1 FROM olive_branch WHERE project_id = temp_project_id AND talent_id = temp_talent_id)
               AND NOT EXISTS (SELECT 1 FROM talent_profile tp JOIN project p ON tp.user_id = p.creator_id WHERE tp.id = temp_talent_id AND p.id = temp_project_id) THEN
                INSERT INTO olive_branch (project_id, talent_id, message, status)
                VALUES (temp_project_id, temp_talent_id, '同学你好，看了你的资料觉得很合适，欢迎加入！', floor(random() * 3)::integer);
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

            INSERT INTO "order" (user_id, amount, product_id, status, trade_no)
            SELECT temp_user_id, price, temp_product_id,
                   floor(random() * 4)::integer,
                   'WX' || floor(random() * 1000000000000000)::text
            FROM product WHERE id = temp_product_id;
        END LOOP;
    END;

    RAISE NOTICE 'Seed completed: % users, % projects created.', num_users, num_projects;
END;
$$ LANGUAGE plpgsql;

-- 使用示例:
SELECT seed_test_data(50, 10);
