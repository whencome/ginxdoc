package ginxdoc

import (
	"embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 使用embed嵌入静态资源

//go:embed static/*
var staticRes embed.FS

//go:embed templates/*
var templatesRes embed.FS

type KVMap map[string]string
type KVMapSlice []KVMap

func (ks KVMapSlice) Len() int           { return len(ks) }
func (ks KVMapSlice) Less(i, j int) bool { return ks[i]["name"] < ks[j]["name"] }
func (ks KVMapSlice) Swap(i, j int)      { ks[i], ks[j] = ks[j], ks[i] }

type RouterMap map[string][]KVMap
type DataMap map[string]RouterMap

var rootPath string

var templateMap = KVMap{
	"index":              "",
	"css_template_cdn":   "",
	"css_template_local": "",
	"js_template_cdn":    "",
	"js_template_local":  "",
}

func initTemplates() error {
	rootPath = config.StaticResPath
	if err := readTemplate(rootPath); err != nil {
		return err
	}
	return nil
}

func verifyPassword(passwordSha2 string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authPasswordSha2 := c.Request.Header.Get("Auth-Password-SHA2")
		if passwordSha2 != "" && passwordSha2 != authPasswordSha2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
		}
	}
}

func readTemplate(rp string) error {
	templatesPath := filepath.Join(rp, "templates")
	for k := range templateMap {
		tByte, err := os.ReadFile(
			filepath.Join(templatesPath, k+".html"),
		)
		if err != nil {
			return err
		}
		templateMap[k] = string(tByte)
	}
	return nil
}

func renderHtml() string {
	htmlStr := templateMap["index"]
	return strings.Replace(
		strings.Replace(
			htmlStr, "<!-- ___CSS_TEMPLATE___ -->", templateMap["css_template_local"], -1,
		), "<!-- ___JS_TEMPLATE___ -->", templateMap["js_template_local"], -1,
	)
}
