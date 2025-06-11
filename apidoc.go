package ginxdoc

import (
	"strings"
)

// 全局配置
var config *Config

// 文档解析器
var docParser *DocParser

// 保存全局文档信息
var apiDocs *DocGroup

// 文档映射关系
var docMaps map[string]*DocInfo

// 保存全局注册的结构体
var registeredTypes = make(map[string]interface{})

// 响应（示例）包装函数，用于组成完整的响应数据，此方法用于成功的响应数据
var responseWrapperFunc func(interface{}) interface{}

// 响应结果文档报装方法
var responseDocWrapperFunc func(respType string, respDesc string) string

// 全局文档markdown内容，此内容将附加到每个文档的末尾
var globalDocMD string

// 字段缩进
var fieldIndent string = "&nbsp;&nbsp;&nbsp;&nbsp;"

// 数据解析深度
var dataNestDepth int = 3

// 初始化全局对象
func init() {
	config = DefaultConfig()
	apiDocs = &DocGroup{
		Name:        "",                   // 分组名称
		Description: "",                   // 分组描述
		Sort:        100,                  // 用于控制文档排序
		Groups:      make([]*DocGroup, 0), // 子分组
		Docs:        make([]*DocInfo, 0),  // 文档列表
	}
	docMaps = make(map[string]*DocInfo)
	responseWrapperFunc = func(v interface{}) interface{} {
		return map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    v,
		}
	}
	responseDocWrapperFunc = func(respType string, respDesc string) string {
		return ""
	}
	Init(config)
}

// addDoc 添加文档
func addDoc(doc *DocInfo) {
	if doc.Name == "" {
		return
	}
	// 进行文档重复性检查
	if _, ok := docMaps[doc.Hash]; ok {
		return
	}
	docMaps[doc.Hash] = doc

	// 处理默认分组
	groupName := strings.TrimSpace(doc.Group)
	if groupName == "" {
		apiDocs.Docs = append(apiDocs.Docs, doc)
		return
	}
	// 将文档加入到分组
	found := false
	for _, g := range apiDocs.Groups {
		if g.Name == groupName {
			g.Docs = append(g.Docs, doc)
			found = true
			break
		}
	}
	if !found {
		g := &DocGroup{
			Name:   groupName,
			Sort:   100,
			Docs:   make([]*DocInfo, 0),
			Groups: make([]*DocGroup, 0),
		}
		g.Docs = append(g.Docs, doc)
		apiDocs.Groups = append(apiDocs.Groups, g)
	}
}
