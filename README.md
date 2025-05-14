# WeChat Bot with Dify Integration

这是一个基于 Go 语言开发的企业微信机器人，集成了 Dify API 来提供智能问答服务。当有人在企业微信群中@机器人时，机器人会调用 Dify API 获取回答并以 markdown 格式返回到群聊中。

## 功能特性

- 企业微信群聊消息监听
- 自动识别并处理 @ 消息
- 集成 Dify API 进行智能问答
- Markdown 格式回复
- 完整的日志记录
- 支持 GBK 编码转换

## 技术栈

- Go 1.23.3
- gin-gonic/gin (Web 框架)
- go-resty/resty (HTTP 客户端)
- yaml.v2 (配置文件解析)

## 快速开始

### 前置条件

- Go 1.23.3 或更高版本
- Dify API 账号和密钥
- 企业微信群机器人配置

### 安装

1. 克隆项目
```bash
git clone https://github.com/yourusername/wechatbot-dify.git
cd wechatbot-dify
```

2. 安装依赖
```bash
go mod download
```

### 配置

1. 复制配置文件模板
```bash
cp config-example.yaml config.yaml
```

2. 修改 config.yaml 配置：
```yaml
dify:
  api_key: "你的 Dify API 密钥"
  chat_api: "Dify 聊天 API 地址"
  chatflow_app_id: "你的 Dify 应用 ID"

wechat:
  webhook_url: "企业微信机器人的 webhook 地址"
  robot_user_id: "机器人的用户 ID"
```

### 运行

```bash
go run main.go
```

服务将在 8080 端口启动。

## 使用说明

1. 确保服务器能够被企业微信访问（需要公网地址或内网穿透）
2. 在企业微信群中 @机器人
3. 机器人会调用 Dify API 处理问题并返回 markdown 格式的回答

## 注意事项

- 请妥善保管 config.yaml，建议加入 .gitignore
- 确保服务器有足够的带宽和性能
- 建议在生产环境中使用 SSL 证书
- 定期检查日志确保服务正常运行

## 日志

程序会记录详细的运行日志，包括：
- 接收到的消息详情
- API 调用细节
- 错误信息