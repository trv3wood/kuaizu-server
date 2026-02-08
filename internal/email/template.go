package email

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/trv3wood/kuaizu-server/internal/models"
)

// ProjectPromotionData 项目推广邮件数据
type ProjectPromotionData struct {
	Nickname       string
	ProjectName    string
	ProjectDesc    string
	SchoolName     string
	MemberCount    int
	ProjectURL     string
	UnsubscribeURL string
}

// TemplateRenderer 邮件模板渲染器
type TemplateRenderer struct {
	baseURL string
}

// NewTemplateRenderer 创建模板渲染器
func NewTemplateRenderer(baseURL string) *TemplateRenderer {
	return &TemplateRenderer{baseURL: baseURL}
}

// RenderProjectPromotion 渲染项目推广邮件
func (r *TemplateRenderer) RenderProjectPromotion(project *models.Project, nickname *string, unsubscribeToken string) (string, string, error) {
	// 邮件主题
	subject := fmt.Sprintf("【快组】有一个项目可能适合你：%s", project.Name)

	// 准备数据
	data := ProjectPromotionData{
		Nickname:       "同学",
		ProjectName:    project.Name,
		ProjectURL:     fmt.Sprintf("%s/project/%d", r.baseURL, project.ID),
		UnsubscribeURL: fmt.Sprintf("%s/api/email/unsubscribe?token=%s", r.baseURL, unsubscribeToken),
	}

	if nickname != nil && *nickname != "" {
		data.Nickname = *nickname
	}

	if project.Description != nil {
		data.ProjectDesc = *project.Description
	}

	if project.SchoolName != nil {
		data.SchoolName = *project.SchoolName
	}

	if project.MemberCount != nil {
		data.MemberCount = *project.MemberCount
	}

	// 渲染模板
	body, err := r.renderTemplate(projectPromotionTemplate, data)
	if err != nil {
		return "", "", err
	}

	return subject, body, nil
}

func (r *TemplateRenderer) renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// 项目推广邮件模板
const projectPromotionTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'PingFang SC', 'Microsoft YaHei', Arial, sans-serif;
            background: #f5f5f5;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 2px 12px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 30px;
            color: white;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
        }
        .content {
            padding: 30px;
        }
        .greeting {
            font-size: 16px;
            color: #333;
            margin-bottom: 15px;
        }
        .intro {
            font-size: 14px;
            color: #666;
            margin-bottom: 20px;
        }
        .project-card {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            border-left: 4px solid #667eea;
        }
        .project-card h2 {
            margin: 0 0 10px 0;
            color: #333;
            font-size: 18px;
        }
        .project-card p {
            color: #666;
            margin: 8px 0;
            font-size: 14px;
            line-height: 1.6;
        }
        .meta {
            display: flex;
            gap: 20px;
            margin-top: 15px;
            font-size: 14px;
            color: #888;
        }
        .meta span {
            display: inline-block;
        }
        .btn {
            display: inline-block;
            background: #667eea;
            color: white !important;
            padding: 12px 30px;
            border-radius: 6px;
            text-decoration: none;
            margin-top: 20px;
            font-size: 14px;
        }
        .btn:hover {
            background: #5a6fd6;
        }
        .footer {
            padding: 20px 30px;
            background: #f8f9fa;
            font-size: 12px;
            color: #999;
            text-align: center;
        }
        .footer a {
            color: #667eea;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 有一个项目可能适合你！</h1>
        </div>
        
        <div class="content">
            <p class="greeting">Hi {{.Nickname}}，</p>
            <p class="intro">平台上有一个项目正在招募队员，快来看看是否适合你：</p>
            
            <div class="project-card">
                <h2>{{.ProjectName}}</h2>
                {{if .ProjectDesc}}<p>{{.ProjectDesc}}</p>{{end}}
                <div class="meta">
                    {{if .SchoolName}}<span>📍 {{.SchoolName}}</span>{{end}}
                    {{if .MemberCount}}<span>👥 需要 {{.MemberCount}} 人</span>{{end}}
                </div>
            </div>
            
            <a href="{{.ProjectURL}}" class="btn">查看详情 →</a>
        </div>
        
        <div class="footer">
            <p>此邮件由快组平台发送</p>
            <p>如不想收到此类邮件，请 <a href="{{.UnsubscribeURL}}">点击退订</a></p>
        </div>
    </div>
</body>
</html>`
