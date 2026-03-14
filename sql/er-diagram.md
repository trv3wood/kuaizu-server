erDiagram
    %% ================= 关系定义 =================
    
    SCHOOL ||--o{ USER : "属于"
    SCHOOL ||--o{ PROJECT : "归属"
    
    MAJOR ||--o{ USER : "专业"
    MAJOR_CLASS ||--o{ MAJOR : "包含"

    USER ||--o| TALENT_PROFILE : "拥有"
    USER ||--o{ PROJECT : "创建"
    USER ||--o{ ORDER : "支付"
    USER ||--o{ FEEDBACK : "提交"
    USER ||--o{ SUBSCRIBE_CONFIG : "订阅"

    PROJECT ||--o{ PROJECT_APPLICATION : "收到申请"
    USER ||--o{ PROJECT_APPLICATION : "发起申请"

    %% 橄榄枝关系的复杂性
    USER ||--o{ OLIVE_BRANCH_RECORD : "发送橄榄枝"
    USER ||--o{ OLIVE_BRANCH_RECORD : "接收橄榄枝"
    PROJECT ||--o{ OLIVE_BRANCH_RECORD : "作为邀请背景"

    PRODUCT ||--o{ ORDER : "包含"

    USER ||--o{ EMAIL_PROMOTION : "发起"
    ORDER ||--o| EMAIL_PROMOTION : "触发"
    PROJECT ||--o{ EMAIL_PROMOTION : "被推广"