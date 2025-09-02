package websocket

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub(nil)
	assert.NotNil(t, hub)
	assert.Equal(t, int64(100000), hub.config.MaxConnections)
	assert.Equal(t, 30*time.Second, hub.config.HeartbeatInterval)

	hub.Close()
}

func TestHubConnectionManagement(t *testing.T) {
	hub := NewHub(nil)
	defer hub.Close()

	// 测试连接注册
	conn := &Connection{
		ID:       "test_conn_1",
		UserID:   "test_user_1",
		IsAlive:  true,
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	hub.register <- conn
	time.Sleep(100 * time.Millisecond) // 等待处理

	assert.Equal(t, int64(1), hub.GetConnectionCount())
	assert.Equal(t, 1, hub.GetUserConnections("test_user_1"))

	// 测试连接注销
	hub.unregister <- conn
	time.Sleep(100 * time.Millisecond) // 等待处理

	assert.Equal(t, int64(0), hub.GetConnectionCount())
	assert.Equal(t, 0, hub.GetUserConnections("test_user_1"))
}

func TestHubGroupManagement(t *testing.T) {
	hub := NewHub(nil)
	defer hub.Close()

	// 创建测试连接
	conn1 := &Connection{
		ID:       "test_conn_1",
		UserID:   "test_user_1",
		IsAlive:  true,
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	conn2 := &Connection{
		ID:       "test_conn_2",
		UserID:   "test_user_2",
		IsAlive:  true,
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	// 注册连接
	hub.register <- conn1
	hub.register <- conn2
	time.Sleep(100 * time.Millisecond)

	// 加入组
	conn1.JoinGroup("test_group")
	conn2.JoinGroup("test_group")
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 2, hub.GetGroupConnections("test_group"))

	// 离开组
	conn1.LeaveGroup("test_group")
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, hub.GetGroupConnections("test_group"))

	// 清理
	hub.unregister <- conn1
	hub.unregister <- conn2
	time.Sleep(100 * time.Millisecond)
}

func TestHubMessageBroadcasting(t *testing.T) {
	hub := NewHub(nil)
	defer hub.Close()

	// 创建测试连接
	conn := &Connection{
		ID:       "test_conn_1",
		UserID:   "test_user_1",
		IsAlive:  true,
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
	}

	hub.register <- conn
	time.Sleep(100 * time.Millisecond)

	// 测试广播消息
	message := &Message{
		Type: "test",
		Data: "test_data",
	}

	hub.broadcast <- message
	time.Sleep(100 * time.Millisecond)

	// 清理
	hub.unregister <- conn
	time.Sleep(100 * time.Millisecond)
}

func TestConnectionMessageHandling(t *testing.T) {
	hub := NewHub(nil)
	defer hub.Close()

	conn := &Connection{
		ID:       "test_conn_1",
		UserID:   "test_user_1",
		IsAlive:  true,
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
		Hub:      hub,
	}

	conn.handlePing()

	// 测试加入组消息
	joinMsg := Message{Type: "join_group", Data: "test_group"}
	conn.handleJoinGroup(joinMsg)

	assert.True(t, conn.IsInGroup("test_group"))

	// 测试离开组消息
	leaveMsg := Message{Type: "leave_group", Data: "test_group"}
	conn.handleLeaveGroup(leaveMsg)

	assert.False(t, conn.IsInGroup("test_group"))
}

func TestWebSocketHandler(t *testing.T) {
	hub := NewHub(nil)
	defer hub.Close()

	handler := NewHandler(hub)

	// 测试获取统计信息
	req := httptest.NewRequest("GET", "/ws/stats", nil)
	w := httptest.NewRecorder()

	// 创建Gin上下文
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "total_connections")
}

func TestConfigValidation(t *testing.T) {
	// 测试有效配置
	validConfig := &Config{
		MaxConnections:     1000,
		HeartbeatInterval:  30 * time.Second,
		ConnectionTimeout:  60 * time.Second,
		MessageBufferSize:  256,
		MessageQueueSize:   1000,
		EnableCompression:  true,
		EnableMessageQueue: true,
		EnableCluster:      false,
	}

	err := ValidateConfig(validConfig)
	assert.NoError(t, err)

	// 测试无效配置
	invalidConfig := &Config{
		MaxConnections:     0,
		HeartbeatInterval:  60 * time.Second,
		ConnectionTimeout:  30 * time.Second,
		MessageBufferSize:  0,
		MessageQueueSize:   0,
		EnableCompression:  true,
		EnableMessageQueue: true,
		EnableCluster:      false,
	}

	err = ValidateConfig(invalidConfig)
	assert.Error(t, err)
}

func TestConfigLoading(t *testing.T) {
	// 测试默认配置
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, int64(100000), config.MaxConnections)

	// 测试配置克隆
	clonedConfig := CloneConfig(config)
	assert.NotNil(t, clonedConfig)
	assert.Equal(t, config.MaxConnections, clonedConfig.MaxConnections)

	// 测试配置合并
	config1 := &Config{MaxConnections: 1000}
	config2 := &Config{HeartbeatInterval: 60 * time.Second}

	mergedConfig := MergeConfig(config1, config2)
	assert.Equal(t, int64(1000), mergedConfig.MaxConnections)
	assert.Equal(t, 60*time.Second, mergedConfig.HeartbeatInterval)
}

func TestConnectionGroupOperations(t *testing.T) {
	hub := NewHub(nil)
	defer hub.Close()

	conn := &Connection{
		ID:       "test_conn_1",
		UserID:   "test_user_1",
		IsAlive:  true,
		Groups:   make(map[string]bool),
		Metadata: make(map[string]interface{}),
		Hub:      hub,
	}

	// 测试加入组
	conn.JoinGroup("group1")
	conn.JoinGroup("group2")

	groups := conn.GetGroups()
	assert.Len(t, groups, 2)
	assert.Contains(t, groups, "group1")
	assert.Contains(t, groups, "group2")

	// 测试检查组成员身份
	assert.True(t, conn.IsInGroup("group1"))
	assert.False(t, conn.IsInGroup("group3"))

	// 测试离开组
	conn.LeaveGroup("group1")
	assert.False(t, conn.IsInGroup("group1"))
	assert.True(t, conn.IsInGroup("group2"))

	groups = conn.GetGroups()
	assert.Len(t, groups, 1)
	assert.Contains(t, groups, "group2")
}

func TestMessageSerialization(t *testing.T) {
	message := &Message{
		Type:      "test",
		Data:      "test_data",
		Timestamp: 1234567890,
		From:      "user1",
		To:        "user2",
		Group:     "test_group",
	}

	// 测试序列化
	data, err := json.Marshal(message)
	require.NoError(t, err)

	// 测试反序列化
	var decodedMessage Message
	err = json.Unmarshal(data, &decodedMessage)
	require.NoError(t, err)

	assert.Equal(t, message.Type, decodedMessage.Type)
	assert.Equal(t, message.Data, decodedMessage.Data)
	assert.Equal(t, message.Timestamp, decodedMessage.Timestamp)
	assert.Equal(t, message.From, decodedMessage.From)
	assert.Equal(t, message.To, decodedMessage.To)
	assert.Equal(t, message.Group, decodedMessage.Group)
}
