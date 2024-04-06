package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Message 定义了接收的JSON结构体
type Message struct {
	User    string    `json:"user"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// Messages 用于存储多个Message对象
type Messages struct {
	Messages []Message `json:"messages"`
}

var (
	messagesFile = "data.json"
	mutex        = &sync.Mutex{}
)

// saveMessage 将消息保存到文件
// saveMessage 将消息保存到文件
func saveMessage(msg Message) error {
	mutex.Lock()
	defer mutex.Unlock()

	var msgs Messages
	file, err := ioutil.ReadFile(messagesFile)
	if err == nil {
		// 尝试解析为 Messages 结构体
		err = json.Unmarshal(file, &msgs)
		if err != nil { // 如果解析失败，尝试解析为 Message 数组
			var msgArray []Message
			err = json.Unmarshal(file, &msgArray)
			if err != nil {
				return err // 如果仍然失败，返回错误
			}
			msgs.Messages = msgArray
		}
	} else if os.IsNotExist(err) {
		// 如果文件不存在，初始化 msgs
		msgs = Messages{Messages: []Message{}}
	} else {
		// 其他错误
		return err
	}

	// 添加新消息
	msgs.Messages = append(msgs.Messages, msg)

	// 写回文件
	newData, err := json.MarshalIndent(msgs, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(messagesFile, newData, 0644)
}

func main() {
	r := gin.Default()

	// 跨域处理
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
		} else {
			c.Next()
		}
	})

	r.POST("/", func(c *gin.Context) {
		var msg Message

		if err := c.BindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 保存消息到文件
		if err := saveMessage(msg); err != nil {
			log.Printf("Error saving message: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Message received and saved",
		})
	})

	// 监听并在 0.0.0.0:8080 上启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
