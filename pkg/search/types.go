package search

import "time"

type Config struct {
	IndexPath           string
	DefaultAnalyzer     string
	DefaultSearchFields []string
	OpenTimeout         time.Duration
	QueryTimeout        time.Duration
	BatchSize           int
}

type Doc struct {
	ID     string
	Type   string
	Fields map[string]any
}

// -------- 过滤器 --------
type NumericRangeFilter struct {
	Field   string
	GTE, GT *float64
	LTE, LT *float64
}

type TimeRangeFilter struct {
	Field   string
	From    *time.Time
	To      *time.Time
	IncFrom bool
	IncTo   bool
}

// -------- 高级搜索子句（新增） --------
type ClauseMatch struct { // 单字段 Match / 可带权重
	Field    string
	Query    string
	Boost    *float64
	Operator string // "and"/"or"，默认 or
}

type ClausePhrase struct {
	Field  string
	Phrase string
	Slop   int // 词间隔
	Boost  *float64
}

type ClausePrefix struct {
	Field  string
	Prefix string
	Boost  *float64
}

type ClauseWildcard struct {
	Field   string
	Pattern string
	Boost   *float64
}

type ClauseRegex struct {
	Field   string
	Pattern string
	Boost   *float64
}

type ClauseFuzzy struct {
	Field     string
	Term      string
	Fuzziness int // 0,1,2…
	Prefix    int // 前缀长
	Boost     *float64
}

type ClauseQueryString struct {
	Query  string   // 直接使用 QueryString 语法
	Fields []string // 如果不为空，会转换成 field:(q) OR ...
	Boost  *float64
}

// Facet 聚合
type FacetRequest struct {
	Name  string // 返回名
	Field string // 字段
	Size  int    // Top N
	// 对时间范围/数值区间，如果需要，可以扩展成 DateRange/NumericRangeFacet
}

type SearchRequest struct {
	// 关键字（保留老接口）
	Keyword      string
	SearchFields []string

	// 结构化 Term
	MustTerms    map[string][]string
	MustNotTerms map[string][]string
	ShouldTerms  map[string][]string

	// 数值/时间过滤
	NumericRanges []NumericRangeFilter
	TimeRanges    []TimeRangeFilter

	// 高级查询子句（新增）
	QueryString *ClauseQueryString
	Matches     []ClauseMatch
	Phrases     []ClausePhrase
	Prefixes    []ClausePrefix
	Wildcards   []ClauseWildcard
	Regexps     []ClauseRegex
	Fuzzies     []ClauseFuzzy

	// 布尔控制
	MinShould int // 至少满足多少个 should（对 ShouldTerms + 高级 should 子句生效）

	// Facet 聚合
	Facets []FacetRequest

	// 排序与分页
	SortBy []string
	From   int
	Size   int

	// 字段返回与高亮
	IncludeFields   []string
	Highlight       bool
	HighlightFields []string // 指定需要高亮的字段，默认全部 text 字段
	FragmentSize    int      // 片段长度
	MaxFragments    int      // 每字段片段数
}
type Hit struct {
	ID        string
	Score     float64
	Fields    map[string]any
	Fragments map[string][]string
}
type FacetTerm struct {
	Term  string
	Count int
}
type FacetResult struct {
	Total int
	Terms []FacetTerm
}
type SearchResult struct {
	Total  uint64
	Took   time.Duration
	Hits   []Hit
	Facets map[string]FacetResult
}
