package ginxdoc

import (
    "bytes"
    "encoding/json"
    "fmt"
    "reflect"
    "strings"
)

// DocParser 文档解析器
type DocParser struct{}

func NewDocParser() *DocParser {
    return &DocParser{}
}

// ParseDocPairs 根据文档键值对解析文档信息
func (p *DocParser) ParseDocPairs(keyVals ...interface{}) *DocInfo {
    apiDoc := &DocInfo{}
    size := len(keyVals)
    var request, response interface{}
    for i := 0; i < size; i += 2 {
        if i+1 >= size {
            break
        }
        notation, ok := keyVals[i].(string)
        if !ok {
            continue
        }
        value := keyVals[i+1]
        switch notation {
        case "@Markdown":
            apiDoc.DocMD = value.(string)
        case "@Summary":
            apiDoc.Name = value.(string)
        case "@Description":
            apiDoc.Description = value.(string)
        case "@Router":
            router := value.(string)
            mStart := strings.Index(router, "[")
            mEnd := strings.Index(router, "]")
            var path, methods string
            if mStart > 0 {
                path = strings.TrimSpace(router[:mStart])
                methods = strings.TrimSpace(router[mStart+1 : mEnd])
            } else {
                path = strings.TrimSpace(router)
            }
            apiDoc.Path = path
            apiDoc.Method = methods
        case "@Tags":
            apiDoc.Group = value.(string)
        case "@Produce":
            produce := value.(string)
            apiDoc.MIME = GetMIMEType(produce)
        case "@Param":
            reqParam, ok := p.parseParam(value.(string))
            if ok {
                apiDoc.Params = append(apiDoc.Params, reqParam)
            }
        case "@Request":
            request = value
        case "@Response":
            response = value
        }
    }

    // 解析请求参数
    if len(apiDoc.Params) > 0 {
        apiDoc.ParamMD = p.buildParamMDByParams(apiDoc.Params)
    } else {
        p.parseRequestInfo(apiDoc, request)
    }

    // 解析响应结果
    if IsStruct(response) { // 如果直接传入的结构体，则直接解析
        apiDoc.RespMD = p.buildRespMD(response)
    } else { // 如果是结构体名称，则从注册的struct中获取，此方式需要提前注册结构体
        structName, ok := response.(string)
        if ok {
            if structVal, ok := registeredTypes[structName]; ok {
                apiDoc.RespMD = p.buildRespMD(structVal)
            }
        }
    }

    return apiDoc
}

// ParseDocString 根据文档字符串解析文档信息
func (p *DocParser) ParseDocString(doc string) *DocInfo {
    apiDoc := &DocInfo{}
    // 解析接口文档
    lines := strings.Split(doc, "\n")
    openMarkdown := false
    // 标识是否单独定义了参数，如果是，则不解析结构体
    definedParam := false
    // 响应结果
    respStructName := ""
    reqStructName := ""
    markdown := bytes.Buffer{}
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if openMarkdown {
            if strings.HasPrefix(line, "@Markdown") {
                openMarkdown = false
                continue
            } else {
                markdown.WriteString(line)
                markdown.WriteString("\n")
                continue
            }
        } else {
            if strings.HasPrefix(line, "@Markdown") {
                openMarkdown = !openMarkdown
                continue
            }
            if strings.HasPrefix(line, "@Summary") {
                apiDoc.Name = strings.TrimSpace(strings.TrimPrefix(line, "@Summary"))
                continue
            }
            if strings.HasPrefix(line, "@Description") {
                apiDoc.Description = strings.TrimSpace(strings.TrimPrefix(line, "@Description"))
                continue
            }
            if strings.HasPrefix(line, "@Router") {
                router := strings.TrimSpace(strings.TrimPrefix(line, "@Router"))
                mStart := strings.Index(router, "[")
                mEnd := strings.Index(router, "]")
                var path, methods string
                if mStart > 0 {
                    path = strings.TrimSpace(router[:mStart])
                    methods = strings.TrimSpace(router[mStart+1 : mEnd])
                } else {
                    path = strings.TrimSpace(router)
                }
                apiDoc.Path = path
                apiDoc.Method = methods
                continue
            }
            if strings.HasPrefix(line, "@Tags") {
                apiDoc.Group = strings.TrimSpace(strings.TrimPrefix(line, "@Tags"))
                continue
            }
            if strings.HasPrefix(line, "@Produce") {
                produce := strings.TrimSpace(strings.TrimPrefix(line, "@Produce"))
                apiDoc.MIME = GetMIMEType(produce)
                continue
            }
            if strings.HasPrefix(line, "@Param") {
                definedParam = true
                reqParam, ok := p.parseParam(strings.TrimSpace(strings.TrimPrefix(line, "@Param")))
                if ok {
                    apiDoc.Params = append(apiDoc.Params, reqParam)
                }
                continue
            }
            if strings.HasPrefix(line, "@Request") {
                reqStructName = strings.TrimSpace(strings.TrimPrefix(line, "@Request"))
                continue
            }
            if strings.HasPrefix(line, "@Response") {
                respStructName = strings.TrimSpace(strings.TrimPrefix(line, "@Response"))
                continue
            }
        }
    }

    if definedParam {
        apiDoc.ParamMD = p.buildParamMDByParams(apiDoc.Params)
    } else {
        p.parseRequestInfo(apiDoc, reqStructName)
    }

    if respStructName != "" {
        if structVal, ok := registeredTypes[respStructName]; ok {
            apiDoc.RespMD = p.buildRespMD(structVal)
        }
    }

    apiDoc.DocMD = markdown.String()

    return apiDoc
}

// parseRequestInfo 解析请求信息
func (p *DocParser) parseRequestInfo(doc *DocInfo, request interface{}) {
    var req RequestInfo
    if IsStruct(request) { // 如果直接传入的结构体，则直接解析
        req = p.ParseRequest(request)
    } else { // 如果是结构体名称，则从注册的struct中获取，此方式需要提前注册结构体
        structName, ok := request.(string)
        if ok && structName != "" {
            if structVal, ok := registeredTypes[structName]; ok {
                req = p.ParseRequest(structVal)
            }
        }
    }
    // 请求的字段可能也是结构体（如嵌套结构体），需要再次解析
    req.Fields = p.paeseNestingFormFields(req.Fields)
    // 构造请求markdown内容
    if req.Name != "" {
        doc.ParamMD = p.buildParamMDByStruct(req)
    }
}

// paeseNestingFormFields 解析嵌套字段，目前主要是嵌套结构体
func (p *DocParser) paeseNestingFormFields(fields []FormField) []FormField {
    dstFields := make([]FormField, 0)
    if len(fields) == 0 {
        return dstFields
    }
    for _, field := range fields {
        // 暂时只处理嵌套结构体
        if field.IsStruct && strings.HasSuffix(field.Type, field.Name) {
            subDstFields := p.paeseNestingFormFields(field.Struct.Fields)
            if len(subDstFields) > 0 {
                dstFields = append(dstFields, subDstFields...)
            }
        } else {
            dstFields = append(dstFields, field)
        }
    }
    return dstFields
}

// ParseRequest 解析请求参数结构体信息
// 这是一个定制化的接口，用于gin通过Bind方式绑定参数的请求解析
func (p *DocParser) ParseRequest(v interface{}) RequestInfo {
    if IsNil(v) {
        return RequestInfo{}
    }
    var t reflect.Type
    if t1, ok := v.(reflect.Type); ok {
        t = t1
    } else {
        t = reflect.TypeOf(v)
    }
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    if t.Kind() != reflect.Struct {
        return RequestInfo{}
    }

    // 结构体信息
    structInf := RequestInfo{
        Name:   t.Name(),             // 结构体名称
        Desc:   "",                   // 结构体描述
        Fields: make([]FormField, 0), // 字段信息
    }

    // 解析字段
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        // 显示字段名
        showFieldName := field.Tag.Get("form")
        if strings.Contains(showFieldName, ",") {
            showFieldName = showFieldName[0:strings.Index(showFieldName, ",")]
        }
        if showFieldName == "" {
            showFieldName = field.Name
        }
        // 字段描述
        fieldDesc := field.Tag.Get("desc")
        if fieldDesc == "" {
            // 尝试使用label中的内容作为字段描述
            fieldDesc = field.Tag.Get("label")
        }
        // 是否必填
        required := false
        binding := field.Tag.Get("binding")
        if strings.Contains(binding, "required") {
            required = true
        }

        // 字段信息
        fieldInf := FormField{
            Name:     field.Name,
            IsStruct: false,
            Required: required,
            Type:     field.Type.String(),
            Tag:      showFieldName,
            Desc:     fieldDesc,
        }

        // 处理嵌套结构体
        fieldType := field.Type
        if fieldType.Kind() == reflect.Ptr {
            fieldType = fieldType.Elem()
        }
        if fieldType.Kind() == reflect.Struct {
            childStruct := p.ParseRequest(field.Type)
            childStruct.Name = field.Name
            childStruct.Desc = fieldDesc
            fieldInf.IsStruct = true
            fieldInf.Struct = childStruct
        }
        structInf.Fields = append(structInf.Fields, fieldInf)
    }

    return structInf
}

// ParseStruct 解析通用的结构体信息
func (p *DocParser) ParseStruct(v interface{}) StructInfo {
    // 如果对象为nil，则不处理
    if IsNil(v) {
        return StructInfo{}
    }

    // 获取结构体反射类型
    var t reflect.Type
    if t1, ok := v.(reflect.Type); ok {
        t = t1
    } else {
        t = reflect.TypeOf(v)
    }
    if t.Kind() == reflect.Ptr {
        t = t.Elem()
    }
    if t.Kind() != reflect.Struct {
        return StructInfo{}
    }

    // 结构体信息
    structInf := StructInfo{
        Name:   t.Name(),             // 结构体名称
        Desc:   "",                   // 结构体描述
        Fields: make([]FieldInfo, 0), // 字段信息
    }
    // 解析字段信息
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        // 解析显示字段
        showFieldName := field.Name
        jsonTag := field.Tag.Get("json")
        if jsonTag != "" && jsonTag != "-" {
            if strings.Contains(jsonTag, ",") {
                jsonTag = jsonTag[:strings.Index(jsonTag, ",")]
            }
            showFieldName = jsonTag
        }
        // 解析注释说明，应当放在desc标签中
        descTag := field.Tag.Get("desc")

        // 字段信息
        fieldInf := FieldInfo{
            Name:     field.Name,
            Tag:      showFieldName,
            IsStruct: false,
            Type:     field.Type.String(),
            Desc:     descTag,
        }

        // 处理嵌套结构体
        fieldType := field.Type
        if fieldType.Kind() == reflect.Ptr {
            fieldType = fieldType.Elem()
        }
        if fieldType.Kind() == reflect.Struct {
            childStruct := p.ParseStruct(fieldType)
            childStruct.Name = field.Name
            childStruct.Desc = descTag
            fieldInf.IsStruct = true
            fieldInf.Struct = childStruct
        } else if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
            elemType := fieldType.Elem()
            item := createDefaultValue(elemType)
            if IsStruct(item) {
                childStruct := p.ParseStruct(elemType)
                childStruct.Name = field.Name
                childStruct.Desc = descTag
                fieldInf.IsStruct = true
                fieldInf.Struct = childStruct
            }
        }
        structInf.Fields = append(structInf.Fields, fieldInf)
    }
    return structInf
}

// parseParam 解析参数
func (p *DocParser) parseParam(param string) (ApiReqParam, bool) {
    reqParam := ApiReqParam{}
    if strings.ToLower(param) == "none" {
        return reqParam, false
    }
    chars := []rune(param)
    pos := 0
    openQuote := false
    data := make([]rune, 0)
    writeData := false
    for _, char := range chars {
        if char == '"' {
            openQuote = !openQuote
            writeData = true
        } else if char == ' ' && !openQuote {
            writeData = true
        } else {
            data = append(data, char)
        }
        if writeData {
            writeData = false
            switch pos {
            case 0:
                reqParam.Name = string(data)
            case 1:
                reqParam.Type = string(data)
            case 2:
                reqParam.Required = strings.ToLower(string(data)) == "true"
            case 3:
                reqParam.Description = string(data)
            }
            data = make([]rune, 0)
            pos++
        }
    }
    if len(data) > 0 {
        switch pos {
        case 0:
            reqParam.Name = string(data)
        case 1:
            reqParam.Type = string(data)
        case 2:
            reqParam.Required = strings.ToLower(string(data)) == "true"
        case 3:
            reqParam.Description = string(data)
        }
    }
    return reqParam, true
}

// buildParamMDByParams 根据@Param定义的参数解析请求参数markdown内容
func (p *DocParser) buildParamMDByParams(params []ApiReqParam) string {
    reqParamMD := `
|参数名|必选|类型|说明|
|:----|:----|:----|----|
`
    for _, param := range params {
        reqParamMD += fmt.Sprintf("|%s|%v|%s|%s|\n", param.Name, param.Required, param.Type, param.Description)
    }
    return reqParamMD
}

// buildParamMDByStruct 根据注册路由时使用的结构体或者通过@Request定义的结构体解析markdown内容
func (p *DocParser) buildParamMDByStruct(req RequestInfo) string {
    reqParamMD := ""
    if req.Name != "" {
        reqParamMD += `
|参数名|必选|类型|说明|
|:----|:----|:----|----|
`
        for _, field := range req.Fields {
            reqParamMD += fmt.Sprintf("|%s|%v|%s|%s|\n", field.Tag, field.Required, field.Type, field.Desc)
        }
    }
    return reqParamMD
}

// buildRespMD 解析响应结果
func (p *DocParser) buildRespMD(resp interface{}) string {
    if IsNil(resp) {
        return ""
    }
    obj := p.ParseStruct(resp)
    if obj.Name == "" {
        return ""
    }
    // 解析为markdown内容
    respMD := `
|参数名|类型|说明|
|:----|:----|----|
`
    respMD += p.buildStructMDBody(obj, "")

    // 添加相应结果示例
    respDemoValue := CreateDefaultInstance(reflect.TypeOf(resp))
    jsonDemo, err := json.MarshalIndent(respDemoValue.Interface(), "", "    ")
    if err == nil {
        respMD += fmt.Sprintf("\n\n**示例**\n\n```json\n%s\n```\n", string(jsonDemo))
    }

    return respMD
}

// buildStructMDBody 构造结构体markdown内容，主要用于响应结果解析
func (p *DocParser) buildStructMDBody(obj StructInfo, fieldPrefix string) string {
    reqParamMD := ""
    for _, field := range obj.Fields {
        reqParamMD += fmt.Sprintf("|%s|%s|%s|\n", fieldPrefix+field.Tag, field.Type, field.Desc)
        if field.IsStruct {
            reqParamMD += p.buildStructMDBody(field.Struct, fieldPrefix+field.Tag+".")
            continue
        }
    }
    return reqParamMD
}
