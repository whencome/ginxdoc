package apidoc

import (
	"fmt"
	"strings"
)

// 全局配置
var config *Config

// 文档解析器
var docParser *DocParser

// 保存全局文档信息
var apiDocs *DocGroup

// 文档映射关系
var docMaps map[string]*ApiDocInfo

// 保存全局注册的结构体
var registeredTypes = make(map[string]interface{})

// 初始化全局对象
func init() {
	config = DefaultConfig()
	apiDocs = &DocGroup{
		Name:        "",                     // 分组名称
		Description: "",                     // 分组描述
		Sort:        100,                    // 用于控制文档排序
		Groups:      make([]*DocGroup, 0),   // 子分组
		Docs:        make([]*ApiDocInfo, 0), // 文档列表
	}
	docMaps = make(map[string]*ApiDocInfo)
	Init(config)
}

// DefaultConfig 生成一个默认配置
func DefaultConfig() *Config {
	return &Config{
		// Title, default `API Doc`
		Title: "Ginx文档",
		// Version, default `1.0.0`
		Version: "1.0",
		// Description
		Description: "",
		// Custom url prefix, default `/docs/api`
		UrlPrefix: "/ginx/docs",
		// No document text, default `No documentation found for this API`
		NoDocText: "<no documents>",

		// 是否启用文档
		EnableDoc: true,

		// SHA256 encrypted authorization password, e.g. here is admin
		// echo -n admin | shasum -a 256
		// `8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918`
		PasswordSha2: "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918",
	}
}

// Init 初始化配置，此方法应当在注册业务路由之前调用，否则无法解析接口文档
func Init(c *Config) {
	if c == nil {
		c = DefaultConfig()
	}
	// 初始化全局对象
	if !c.EnableDoc { // 不启用文档
		return
	}
	// 初始化全局对象
	config = c
	docParser = NewDocParser()
}

// AddStruct 注册结构体
func AddStruct(typeName string, v interface{}) {
	if !IsStruct(v) {
		return
	}
	registeredTypes[typeName] = v
}

// AddStructs 批量注册结构体信息
func AddStructs(strcutMaps map[string]interface{}) {
	if len(strcutMaps) == 0 {
		return
	}
	for name, structV := range strcutMaps {
		AddStruct(name, structV)
	}
}

// NewDoc 添加api文档信息
func NewDoc(keyvals ...interface{}) {
	size := len(keyvals)
	if size == 0 {
		return
	}
	var doc *ApiDocInfo
	if size == 1 {
		doc = docParser.ParseDocString(keyvals[0].(string))
	} else {
		doc = docParser.ParseDocPairs(keyvals...)
	}
	doc.Hash = Md5Short(fmt.Sprintf("%s/%s/%s", doc.Group, doc.Name, doc.Path))
	addDoc(doc)
	return
}

// addDoc 添加文档
func addDoc(doc *ApiDocInfo) {
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
			Docs:   make([]*ApiDocInfo, 0),
			Groups: make([]*DocGroup, 0),
		}
		g.Docs = append(g.Docs, doc)
		apiDocs.Groups = append(apiDocs.Groups, g)
	}
}
