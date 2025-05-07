package ginxdoc

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"reflect"
)

// MIMEMaps 定义mime映射关系
var MIMEMaps = map[string]string{
	"text": "text/plain",                                                              // 纯文本（.txt）
	"bin":  "application/octet-stream",                                                // 二进制数据
	"html": "text/html",                                                               // HTML文档（.html）
	"css":  "text/css",                                                                // 层叠样式表（.css）
	"csv":  "text/csv",                                                                // 逗号分隔值（.csv）
	"json": "application/json",                                                        // JSON数据（.json）
	"xml":  "application/xml",                                                         // XML文档（.xml）
	"jpeg": "image/jpeg",                                                              // JPEG图片（.jpg）
	"png":  "image/png",                                                               // PNG图片（.png）
	"webp": "image/webp",                                                              // WebP图片（.webp）
	"svg":  "image/svg+xml",                                                           // 矢量图（.svg）
	"mpeg": "audio/mpeg",                                                              // MP3音频（.mp3）
	"mp4":  "video/mp4",                                                               // MP4视频（.mp4）
	"ogg":  "application/ogg",                                                         // OGG媒体（.ogv/.oga）
	"pdf":  "application/pdf",                                                         // PDF文档（.pdf）
	"doc":  "application/msword",                                                      // Word旧格式（.doc）
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document", // DOCX（.docx）
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",       // （.xlsx）
	"ppt":  "application/vnd.ms-powerpoint",                                           // .ppt
	"zip":  "application/zip",                                                         // ZIP压缩（.zip）
	"rar":  "application/x-rar-compressed",                                            // RAR压缩（.rar）
	"7z":   "application/x-7z-compressed",                                             // 7-Zip压缩（.7z）
}

// GetMIMEType 获取mime类型
func GetMIMEType(contentType string) string {
	if mimetype, ok := MIMEMaps[contentType]; ok {
		return mimetype
	}
	return "text/plain"
}

// Md5 生成md5 hash
func Md5(str string) string {
	h := md5.New()
	_, _ = io.WriteString(h, str)
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Md5Short 16位MD5
func Md5Short(str string) string {
	hexStr := Md5(str)
	return string([]byte(hexStr)[8:24])
}

// IsNil 判断给定的值是否为nil
func IsNil(i interface{}) bool {
	ret := i == nil
	// 需要进一步做判断
	if !ret {
		vi := reflect.ValueOf(i)
		kind := reflect.ValueOf(i).Kind()
		if kind == reflect.Slice ||
			kind == reflect.Map ||
			kind == reflect.Chan ||
			kind == reflect.Interface ||
			kind == reflect.Func ||
			kind == reflect.Ptr {
			return vi.IsNil()
		}
	}
	return ret
}

// IsStruct 判断给定对象是否是一个结构体
func IsStruct(v interface{}) bool {
	if IsNil(v) {
		return false
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return true
	}
	return false
}

func IsFunc(param interface{}) bool {
	t := reflect.TypeOf(param)
	return t != nil && t.Kind() == reflect.Func
}

// IsList 判断给定的对象是否是数组或者切片
func IsList(i interface{}) bool {
	if i == nil {
		return false
	}
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		return true
	}
	return false
}

// GetListStructItem 获取列表结构体元素
func GetListStructItem(i interface{}) interface{} {
	val := reflect.ValueOf(i)

	// 检查是否是切片或数组
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}

	// 检查元素类型
	elemType := val.Type().Elem()

	// 如果元素是结构体
	if elemType.Kind() == reflect.Struct {
		// 创建该结构体的新实例
		return reflect.New(elemType).Interface()
	}

	// 元素不是结构体
	return nil
}

// 创建类型的默认值
func createDefaultValue(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(0) // 整数默认值
	case reflect.String:
		return reflect.ValueOf("") // 字符串默认值
	case reflect.Struct:
		return reflect.New(t).Elem() // 结构体默认值（各字段为零值）
	default:
		return reflect.Zero(t) // 其他类型返回零值
	}
}

// CreateDefaultInstance 创建结构体的默认实例
func CreateDefaultInstance(typ reflect.Type) reflect.Value {
	// 处理指针类型
	if typ.Kind() == reflect.Ptr {
		elem := CreateDefaultInstance(typ.Elem())
		ptr := reflect.New(typ.Elem())
		ptr.Elem().Set(elem)
		return ptr
	}
	// 创建对应类型的零值
	val := reflect.New(typ).Elem()
	// 根据不同类型处理
	switch typ.Kind() {
	case reflect.Struct:
		// 递归处理结构体字段
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.IsExported() { // 只处理可导出字段
				fieldVal := CreateDefaultInstance(field.Type)
				val.Field(i).Set(fieldVal)
			}
		}

	case reflect.Slice:
		// 创建包含一个元素的切片
		elemType := typ.Elem()
		elem := CreateDefaultInstance(elemType)
		slice := reflect.MakeSlice(typ, 1, 1)
		slice.Index(0).Set(elem)
		val.Set(slice)

	case reflect.Array:
		// 创建数组并初始化第一个元素
		elemType := typ.Elem()
		elem := CreateDefaultInstance(elemType)
		arr := reflect.New(typ).Elem()
		if arr.Len() > 0 {
			arr.Index(0).Set(elem)
		}
		val.Set(arr)

	// 基础类型会自动初始化为零值
	default:
		// 不需要额外处理，reflect.New 已经创建了零值
	}

	return val
}
