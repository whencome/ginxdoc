package ginxdoc

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultConfig 生成一个默认配置
func DefaultConfig() *Config {
	docPath, err := os.Getwd()
	if err != nil {
		docPath = "./"
	}
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
		// 静态资源路径
		StaticResPath: docPath,

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
	if !config.EnableDoc || !IsStruct(v) {
		return
	}
	registeredTypes[typeName] = v
}

// AddStructs 批量注册结构体信息
func AddStructs(structMaps map[string]interface{}) {
	if !config.EnableDoc || len(structMaps) == 0 {
		return
	}
	for name, structV := range structMaps {
		AddStruct(name, structV)
	}
}

// SetResponseWrapFunc 设置响应包装函数
func SetResponseWrapFunc(f func(v interface{}) interface{}) {
	responseWrapperFunc = f
}

// SetGlobalDocMD 设置全局文档md内容，此内容将附加到每个文档末尾
func SetGlobalDocMD(md string) {
	globalDocMD = md
}

// NewDoc 添加api文档信息
func NewDoc(keyVals ...interface{}) {
	if !config.EnableDoc {
		return
	}
	size := len(keyVals)
	if size == 0 {
		return
	}
	var doc *DocInfo
	if size == 1 {
		doc = docParser.ParseDocString(keyVals[0].(string))
	} else {
		doc = docParser.ParseDocPairs(keyVals...)
	}
	doc.Hash = Md5Short(fmt.Sprintf("%s/%s/%s", doc.Group, doc.Name, doc.Path))
	addDoc(doc)
	return
}

// Register 注册文档路由
func Register(r *gin.Engine, middlewares ...gin.HandlerFunc) (err error) {
	if !config.EnableDoc {
		return nil
	}
	if err := initTemplates(); err != nil {
		return err
	}

	dataMap := apiDocs.ToApiData()

	g0 := r.Group("")
	g0.Use(middlewares...)

	g0.Static(config.UrlPrefix+"/static", filepath.Join(rootPath, "static"))

	g0.GET(config.UrlPrefix+"/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, renderHtml())
	})

	g0.GET(config.UrlPrefix+"/data",
		verifyPassword(config.PasswordSha2),
		func(c *gin.Context) {
			urlPrefix := config.UrlPrefix
			referer := c.Request.Header.Get("referer")
			if referer == "" {
				referer = "http://127.0.0.1"
			}
			host := strings.Split(referer, urlPrefix)[0]

			c.JSON(http.StatusOK, gin.H{
				"PROJECT_NAME":    PROJECT_NAME,
				"PROJECT_VERSION": PROJECT_VERSION,
				"host":            host,
				"title":           config.Title,
				"version":         config.Version,
				"description":     config.Description,
				"noDocText":       config.NoDocText,
				"data":            dataMap,
			})
		})

	return
}
