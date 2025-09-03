package llm_test

import (
	"HibiscusIM/pkg/llm"
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// 模拟的 TTS 回调函数
func mockTTSCallback(segment string, playID string, autoHangup bool) error {
	// 模拟输出 TTS 内容
	return nil
}

func TestLLMHandler_QueryStream(t *testing.T) {
	// 准备测试所需的参数
	apiKey := "your-openai-api-key"
	endpoint := "https://api.openai.com/v1"
	systemPrompt := "You are a helpful assistant."

	// 创建 logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// 创建 LLMHandler 实例
	ctx := context.Background()
	llmHandler := llm.NewLLMHandler(ctx, apiKey, endpoint, systemPrompt, logger)

	// 模拟一个查询
	text := "你好，今天的天气怎么样？"
	model := "gpt-4"

	// 调用 QueryStream 测试函数
	_, err := llmHandler.QueryStream(model, text, mockTTSCallback)
	assert.NoError(t, err, "QueryStream 执行失败")

	// 检查返回结果是否符合预期
	// 这里可以添加更具体的断言来验证流式响应的内容
}

func TestLLMHandler_Query(t *testing.T) {
	// 准备测试所需的参数
	apiKey := "sk-2fd01e230c274cf79fa50fb03ffde1da"
	endpoint := "https://dashscope.aliyuncs.com/compatible-mode/v1"
	systemPrompt := "You are a helpful assistant."

	// 创建 logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// 创建 LLMHandler 实例
	ctx := context.Background()
	llmHandler := llm.NewLLMHandler(ctx, apiKey, endpoint, systemPrompt, logger)

	// 模拟一个查询
	text := "请给我一些编程建议。"
	model := "qwen-turbo"

	// 调用 Query 测试函数
	response, hangupTool, err := llmHandler.Query(model, text)
	assert.NoError(t, err, "Query 执行失败")
	assert.NotNil(t, response, "没有返回内容")
	if hangupTool != nil {
		t.Logf("会话被挂断，原因: %s", hangupTool.Reason)
	}

	// 断言返回的内容符合预期
	assert.Contains(t, response, "编程", "返回的建议应该包含编程内容")
}

func TestLLMHandler_Reset(t *testing.T) {
	// 准备测试所需的参数
	apiKey := "your-openai-api-key"
	endpoint := "https://api.openai.com/v1"
	systemPrompt := "You are a helpful assistant."

	// 创建 logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// 创建 LLMHandler 实例
	ctx := context.Background()
	llmHandler := llm.NewLLMHandler(ctx, apiKey, endpoint, systemPrompt, logger)

	// 执行一次查询，修改历史记录
	_, _, err := llmHandler.Query("gpt-4", "你好吗？")
	assert.NoError(t, err, "Query 执行失败")

	// 执行重置操作
	llmHandler.Reset()
}
