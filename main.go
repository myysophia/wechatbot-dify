package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Dify struct {
		ApiKey        string `yaml:"api_key"`
		ChatApi       string `yaml:"chat_api"`
		ChatflowAppId string `yaml:"chatflow_app_id"`
	} `yaml:"dify"`
	Wechat struct {
		WebhookUrl  string `yaml:"webhook_url"`
		RobotUserId string `yaml:"robot_user_id"`
	} `yaml:"wechat"`
}

var config Config

func init() {
	// 读取配置文件
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析配置文件
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
}

func main() {
	r := gin.Default()

	// 设置响应头编码
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Next()
	})

	// 接收企业微信消息
	r.POST("/wechat", func(c *gin.Context) {
		// 读取原始请求体
		body, err := io.ReadAll(transform.NewReader(c.Request.Body, simplifiedchinese.GBK.NewDecoder()))
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			c.JSON(400, gin.H{"error": "failed to read request body"})
			return
		}

		// 打印转换后的请求体用于调试
		log.Printf("转换后的请求体: %s", string(body))

		// 企业微信 webhook 消息结构体
		var payload struct {
			MsgType string `json:"msgtype"`
			Text    struct {
				Content       string   `json:"content"`
				MentionedList []string `json:"mentioned_list,omitempty"`
			} `json:"text"`
		}

		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("解析 JSON 失败: %v", err)
			c.JSON(400, gin.H{"error": "invalid json"})
			return
		}

		// 打印接收到的消息详情
		log.Printf("收到消息，提及列表: %v", payload.Text.MentionedList)
		log.Printf("配置的机器人ID: %s", config.Wechat.RobotUserId)

		// 判断是否被 @ 到
		mentioned := false
		for _, id := range payload.Text.MentionedList {
			if id == config.Wechat.RobotUserId || id == "@all" {
				mentioned = true
				break
			}
		}
		if !mentioned {
			log.Printf("未被@到，实际提及列表: %v, 期望机器人ID: %s", payload.Text.MentionedList, config.Wechat.RobotUserId)
			c.JSON(200, gin.H{"msg": "not triggered"})
			return
		}

		// 提取用户问题内容
		content := payload.Text.Content
		question := strings.TrimSpace(strings.Replace(content, "@智能运维机器人", "", 1))
		if question == "" {
			question = "你想问什么？"
		}
		log.Printf("用户提问内容: %s", question)

		// 调用 Dify Chatflow API
		client := resty.New()
		var result map[string]interface{}

		// 构建请求体
		requestBody := map[string]interface{}{
			"inputs":          map[string]string{},
			"query":           question,
			"response_mode":   "blocking",
			"conversation_id": "",
			"user":            "wecom-user-001",
		}

		// 打印请求详情
		requestJSON, _ := json.MarshalIndent(requestBody, "", "  ")
		log.Printf("Dify API 请求地址: %s", config.Dify.ChatApi)
		log.Printf("Dify API 请求头: Authorization: Bearer %s", config.Dify.ApiKey)
		log.Printf("Dify API 请求体:\n%s", string(requestJSON))

		resp, err := client.R().
			SetHeader("Authorization", "Bearer "+config.Dify.ApiKey).
			SetHeader("Content-Type", "application/json").
			SetBody(requestBody).
			SetResult(&result).
			Post(config.Dify.ChatApi)

		if err != nil {
			log.Printf("调用 Dify 失败: %v", err)
			c.JSON(500, gin.H{"error": "dify 调用失败"})
			return
		}

		// 打印响应详情
		log.Printf("Dify API 响应状态码: %d", resp.StatusCode())
		// log.Printf("Dify API 响应头: %v", resp.Header())
		// log.Printf("Dify API 原始响应体: %s", string(resp.Body()))
		// log.Printf("Dify API 解析后结果: %+v", result)

		// 提取 Dify 返回结果
		answer, ok := result["answer"].(string)
		if !ok {
			log.Printf("无法从响应中提取 answer 字段，result 类型: %T，值: %+v", result["answer"], result["answer"])
		}
		if answer == "" {
			log.Printf("answer 字段为空")
			answer = "抱歉，我暂时无法回答你的问题。"
		}

		// 组织回复格式为 markdown
		reply := fmt.Sprintf("## 问题\n%s\n\n## 回答\n%s", question, answer)
		log.Printf("回复内容: %s", reply)

		// 发送到企业微信群，使用 markdown 类型
		_, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]interface{}{
				"msgtype": "markdown",
				"markdown": map[string]string{
					"content": reply,
				},
			}).Post(config.Wechat.WebhookUrl)

		if err != nil {
			log.Println("回复企业微信群失败:", err)
		}

		c.JSON(200, gin.H{"msg": "replied"})
	})

	r.Run(":8080")
}
