# ginxdoc

## 简介

ginxdoc是一个专门为ginx框架开发的一个文档工具，虽如此，其也可以用于其他框架。和其它文档工具不同的是，本项目是通过方法调用的方式添加文档信息，这不可避免的在编译后的程序中增加了相关文档信息，但优点也同样明显，实现简单，可以很好的解析项目中的各种结构体等，这样可以输出详细的文档信息。

本项目借鉴了[https://github.com/kwkwc/gin-docs](https://github.com/kwkwc/gin-docs)项目，对后端代码进行了部分重写以实现预期的效果。

## 添加接口文档

* 此处的实例以ginx框架为例，其他框架也可以使用此工具，请自行研究和修改。

```go
// RegisterRoute 注册路由
func (h *DemoHandler) RegisterRoute(g *gin.RouterGroup) {
	ginxdoc.NewDoc(
		"@Summary", "查询列表",
		"@Description", "根据指定条件查询列表请求",
		"@Tags", "列表查询",
		"@Produce", "json",
		"@Request", requests.ListRequest{},
		"@Response", response.ListResponse{},
		"@Router", "/list [get]",
	)
	g.GET("/list", ginx.NewApiHandler(requests.ListRequest{}, h.List))

	ginxdoc.NewDoc(
		"@Summary", "查询详情",
		"@Description", "根据ID查询详情信息",
		"@Tags", "列表查询",
		"@Produce", "json",
		"@Request", requests.DetailRequest{},
		"@Response", response.DetailResponse{},
		"@Router", "/detail [get]",
	)
	g.GET("/detail", ginx.NewApiHandler(requests.DetailRequest{}, h.Detail))
}
```

在上面的示例中，各注解的说明如下：

* **@Summary** 接口的简单说明，可以理解为接口的名称
* **@Description** 接口的文本说明，可以添加较为详细的介绍，此内容为纯文本信息
* **@Produce** 响应的内容类型，如json、xml等，最终将转换为MIME类型
* **@Param** 定义请求参数信息，格式为：@Param 字段名 类型 是否必填 参数说明
* **@Response** 响应内容，这里是可以是对应结构体实例（空实例），也可以是注册的结构体名称。如果是结构体名称，应当在注册路由之前调用ginxdoc.AddStructs或ginxdoc.AddStruct进行注册，具体参考代码。需要说明的是，这里的结构体名称不是定义的名称，而是注册时指定的字符串名称
* **Request** 请求的结构体信息，可以是结构体实例也可以是注册的结构体名称，与@Response相同
* **@Markdown** 此标签表明对应的值是markdown格式，此markdown内容将附加到文档末尾，可以有多个@Markdown内容，但markdown内容将按照添加的顺序依次添加到文档中
* **@Router** 注册的路由以及请求方式

**关于字段说明**

* ginxdoc会使用字段的“desc”tag中的内容作为字段说明；
* 显示字段暂时为写死内容，如果时请求参数，则解析“form”tag内容作为显示字段，其它则取"json" tag内容作为显示字段