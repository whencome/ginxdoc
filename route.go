package apidoc

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	PROJECT_NAME    = "Ginx Docs"
	PROJECT_VERSION = Version
)

type KVMap map[string]string
type KVMapSlice []KVMap

func (ks KVMapSlice) Len() int           { return len(ks) }
func (ks KVMapSlice) Less(i, j int) bool { return ks[i]["name"] < ks[j]["name"] }
func (ks KVMapSlice) Swap(i, j int)      { ks[i], ks[j] = ks[j], ks[i] }

type RouterMap map[string][]KVMap
type DataMap map[string]RouterMap

var docMap = make(map[string]KVMap)

var pkgMap = make(map[string][]string)

var rootPath string

var templateMap = KVMap{
	"index":              "",
	"css_template_cdn":   "",
	"css_template_local": "",
	"js_template_cdn":    "",
	"js_template_local":  "",
}

func initTemplates() error {
	rootPath = getRootPath()
	if err := readTemplate(rootPath); err != nil {
		return err
	}
	return nil
}

func rootPathFunc() {}
func getRootPath() string {
	funcValue := reflect.ValueOf(rootPathFunc)
	fn := runtime.FuncForPC(funcValue.Pointer())
	filePath, _ := fn.FileLine(0)
	rp := filepath.Dir(filePath)

	return rp
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

// Register 注册文档路由
func Register(r *gin.Engine, middlewares ...gin.HandlerFunc) (err error) {
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
