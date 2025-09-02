package search

import (
	"context"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1) 初始化索引
	cfg := Config{
		IndexPath:           "idx.bleve",
		DefaultAnalyzer:     "standard",
		DefaultSearchFields: []string{"title", "body"},
		QueryTimeout:        2 * time.Second,
		BatchSize:           200,
	}
	m := BuildIndexMapping(cfg.DefaultAnalyzer)
	engine, err := New(cfg, mapping.IndexMapping(m))
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()

	// 2) Gin
	r := gin.Default()

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) { c.String(200, "ok") })

	// 写入单文档
	r.POST("/index", func(c *gin.Context) {
		var d Doc
		if err := c.ShouldBindJSON(&d); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if d.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
			return
		}
		if err := engine.Index(c.Request.Context(), d); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 批量写入
	r.POST("/index/batch", func(c *gin.Context) {
		var docs []Doc
		if err := c.ShouldBindJSON(&docs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := engine.IndexBatch(c.Request.Context(), docs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "count": len(docs)})
	})

	// 删除
	r.DELETE("/doc/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
			return
		}
		if err := engine.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 搜索
	r.POST("/search", func(c *gin.Context) {
		var req SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := engine.Search(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// 3) 优雅退出
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}
}
