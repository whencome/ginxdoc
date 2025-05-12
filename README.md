# ginxdoc

## 简介

ginxdoc是一个专门为ginx框架开发的一个文档工具，虽如此，其也可以用于其他框架。和其它文档工具不同的是，本项目是通过方法调用的方式添加文档信息，这不可避免的在编译后的程序中增加了相关文档信息，但优点也同样明显，实现简单，可以很好的解析项目中的各种结构体等，这样可以输出详细的文档信息。

本项目借鉴了[https://github.com/kwkwc/gin-docs](https://github.com/kwkwc/gin-docs)项目，对后端代码进行了部分重写以实现预期的效果。

## 在项目中使用ginxdoc

### 1. 初始化ginxdoc

**注意**：此步骤应当在注册接口路由之前调用。

```go
config := &ginxdoc.Config{
    Title:         "接口文档",
    Version:       "1.0",
    Description:   "接口文档",
    UrlPrefix:     "/ginxdocs",
    EnableDoc:     true,
    StaticResPath: "/path/to/ginxdoc/static/resources",
    // SHA256 encrypted authorization password, e.g. here is admin
    // echo -n admin | shasum -a 256
    // `8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918`
    PasswordSha2: "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918",
}
ginxdoc.Init(config)
```

### 2. 注册文档路由

**注意**：此步骤应当在注册接口路由之后调用。

```go
err := ginxdoc.Register(r)
if err != nil {
    log.Errorf("register ginxdoc fail： %s", err)
}
```

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

    ginxdoc.NewDoc(
		"@Summary", "获取当前账户信息",
		"@Description", "获取当前账户信息",
		"@Tags", "账户管理",
		"@Produce", "json",
		"@Return", "id string 账户id",
		"@Return", "account.balance  float32  账户余额",
		"@Return", "account.active   bool  是否激活",
		"@Router", "/account [post]",
	)
	g.POST("/account", ginx.NewApiHandler(nil, h.Account))
}
```

在上面的示例中，各注解的说明如下：

* **@Summary** 接口的简单说明，可以理解为接口的名称
* **@Description** 接口的文本说明，可以添加较为详细的介绍，此内容为纯文本信息
* **@Produce** 响应的内容类型，如json、xml等，最终将转换为MIME类型
* **@Param** 定义请求参数信息，格式为：“@Param 字段名 类型 是否必填 参数说明”，一个文档可以有多个@Param，如果@Param和@Request同时存在，则@Param优先级高
* **@Return** 定义返回字段信息，格式为：“@Return 字段名 类型 字段说明”，一个文档可以有多个@Return，如果@Return和@Response同时存在，则@Return的优先级高（此时忽略@Response）
* **@Response** 响应内容，这里是可以是对应结构体实例（空实例），也可以是响应结果文本说明，格式为：“@Response name type desc”，一个文档只支持一个@Response说明
* **@Request** 请求的结构体实例（空实例），一个文档只支持一个@Request说明
* **@Markdown** 此标签表明对应的值是markdown格式，此markdown内容将附加到文档末尾，可以有多个@Markdown内容，但markdown内容将按照添加的顺序依次添加到文档中
* **@WrapResponse** 是否包装响应结果，取值为on/off，默认为on，如果为on，则使用ResponseWrapFunc对返回内容进行包装，参考SetResponseWrapFunc方法
* **@Router** 注册的路由以及请求方式

**关于字段说明**

* ginxdoc会使用字段的“desc”tag中的内容作为字段说明；
* 显示字段暂时为写死内容，如果时请求参数，则解析“form”tag内容作为显示字段，其它则取"json" tag内容作为显示字段

## 统一返回值格式

通常，接口文档中只是定义了接口的实际返回值（ginx中），再返回到客户端前会做一层封装，包括是否成功、消息提示等。可以使用**SetResponseWrapFunc**方法来注册一个方法，用于对返回值进行封装。如：
```go
ginxdoc.SetResponseWrapFunc(func(v interface{}) interface{} {
		return map[string]interface{}{
			"code":    0,
			"message": "ok",
			"data":    v,
		}
	})
```
如果不需要对返回结果进行包装，则可以使用**ginxdoc.SetResponseWrapFunc(nil)**来实现。

## 添加全局文档说明

如果需要为每个文档加上相同的说明内容，则可以使用“SetGlobalDocMD”来实现，如：
```go
	ginxdoc.SetGlobalDocMD(`
### 返回值说明

- **code**: 响应码，0-成功，其它-失败
- **message**: 响应结果消息，如果code表示失败，这里是失败提示
- **data**: 响应数据，只有code表示成功时，才取data字段值

    `)
```
