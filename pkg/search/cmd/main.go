package main

import (
	"HibiscusIM/pkg/search"
	"context"
	"fmt"
	"time"

	"github.com/blevesearch/bleve/v2/mapping"
)

func main() {
	// 1. 配置
	cfg := search.Config{
		IndexPath:           "idx.bleve",
		DefaultAnalyzer:     "standard",
		DefaultSearchFields: []string{"title", "body"},
		QueryTimeout:        2 * time.Second,
		BatchSize:           200,
	}

	// 2. 建立索引映射
	m := search.BuildIndexMapping(cfg.DefaultAnalyzer)

	// 3. 初始化 Engine
	engine, err := search.New(cfg, mapping.IndexMapping(m))
	if err != nil {
		panic(err)
	}
	defer engine.Close()

	// 4. 写入示例文档
	docs := []search.Doc{
		{
			ID:   "1",
			Type: "article",
			Fields: map[string]any{
				"title":     "你好，Bleve",
				"body":      "Go 语言的全文检索示例，支持高亮和结构化过滤。",
				"tags":      []string{"go", "search"},
				"author":    "alice",
				"createdAt": time.Now().Add(-1 * time.Hour),
				"views":     123.0,
			},
		},
		{
			ID:   "2",
			Type: "article",
			Fields: map[string]any{
				"title":     "中文检索小示例",
				"body":      "QueryString、前缀和模糊搜索。",
				"tags":      []string{"中文", "示例"},
				"author":    "bob",
				"createdAt": time.Now(),
				"views":     456.0,
			},
		},
	}
	if err := engine.IndexBatch(context.Background(), docs); err != nil {
		panic(err)
	}

	// 5. 执行搜索
	req := search.SearchRequest{
		QueryString: &search.ClauseQueryString{
			Query:  "Bleve OR 检索",
			Fields: []string{"title", "body"},
		},
		MustTerms: map[string][]string{"author": {"alice", "bob"}},
		NumericRanges: []search.NumericRangeFilter{
			{Field: "views", GTE: f(100)},
		},
		TimeRanges: []search.TimeRangeFilter{
			{Field: "createdAt", From: t(time.Now().Add(-24 * time.Hour)), To: t(time.Now()), IncFrom: true, IncTo: true},
		},
		SortBy:          []string{"-_score", "-createdAt"},
		From:            0,
		Size:            10,
		IncludeFields:   []string{"title", "author", "createdAt", "views"},
		Highlight:       true,
		HighlightFields: []string{"title", "body"},
		Facets: []search.FacetRequest{
			{Name: "by_author", Field: "author", Size: 5},
			{Name: "by_tag", Field: "tags", Size: 5},
		},
	}

	res, err := engine.Search(context.Background(), req)
	if err != nil {
		panic(err)
	}

	// 6. 打印结果
	fmt.Printf("total=%d took=%s\n", res.Total, res.Took)
	for _, h := range res.Hits {
		fmt.Printf("id=%s score=%.3f title=%v author=%v\n", h.ID, h.Score, h.Fields["title"], h.Fields["author"])
		if len(h.Fragments) > 0 {
			fmt.Printf("  highlight: %+v\n", h.Fragments)
		}
	}
	for name, facet := range res.Facets {
		fmt.Printf("facet[%s] total=%d\n", name, facet.Total)
		for _, t := range facet.Terms {
			fmt.Printf("  %s (%d)\n", t.Term, t.Count)
		}
	}
}

func f(v float64) *float64     { return &v }
func t(v time.Time) *time.Time { return &v }
