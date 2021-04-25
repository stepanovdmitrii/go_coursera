package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type ApiGenSignature struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type MethodWrapperParams struct {
	ApiType          string
	ApiGenMethodName string
	ApiGenHttpMethod string
	ApiGenAuth       string
	ApiGenParamsType string
}

type MethodSetParams struct {
	FieldName    string
	ParamName    string
	IsRequired   bool
	HasEnum      bool
	EnumString   string
	HasDefault   bool
	DefaultValue string
	HasMin       bool
	Min          int
	HasMax       bool
	Max          int
}

type ServeHttpParams struct {
	Url     string
	Handler string
}

var wrapperTemplate = template.Must(template.New("wrapperTemplate").Parse(`
func (api *{{.ApiType}}) wrapperApiGen{{.ApiGenMethodName}}(w http.ResponseWriter, r *http.Request){
	if "{{.ApiGenHttpMethod}}" != "" && r.Method != "{{.ApiGenHttpMethod}}" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\" : \"bad method\"}"))
		return
	}

	if {{.ApiGenAuth}} && r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\" : \"unauthorized\"}"))
		return
	}

	var p {{.ApiGenParamsType}}
	setErr := p.ApiGenSet(r)
	if setErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + strings.ToLower(setErr.Error()) +"\"}"))
		return
	}
	
	res, err := api.{{.ApiGenMethodName}}(r.Context(), p)

	if err != nil {
		switch err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(apiErr.Err.Error()) +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + strings.ToLower(err.Error()) +"\"}"))
		}
		return
	}

	mapResponse := map[string]interface{}{"error": "", "response": res}

	data, _ := json.Marshal(mapResponse)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
`))

var strTpl = template.Must(template.New("strTpl").Parse(`
	// {{.FieldName}}


	{{.FieldName}}Raw := r.URL.Query().Get("{{.ParamName}}")
	if r.Method == "POST" {
		{{.FieldName}}Raw = r.FormValue("{{.ParamName}}")
	}

	if {{.IsRequired}} && {{.FieldName}}Raw == "" {
		return fmt.Errorf("{{.ParamName}} must me not empty")
	}

	if {{.HasDefault}} && {{.FieldName}}Raw == "" {
		{{.FieldName}}Raw = "{{.DefaultValue}}"
	}

	if {{.HasEnum}} {
		isValid := false
		for _, enumValue := range strings.Split("{{.EnumString}}", "|") {
			isValid = isValid || {{.FieldName}}Raw == enumValue
		}
		if !isValid {
			return fmt.Errorf("{{.ParamName}} must be one of [" + strings.Join(strings.Split("{{.EnumString}}", "|"), ", ") + "]")
		}
	}

	if {{.HasMin}} && len({{.FieldName}}Raw) < {{.Min}} {
		return fmt.Errorf("{{.ParamName}} len must be >= {{.Min}}")
	}
	in.{{.FieldName}} = {{.FieldName}}Raw
`))

var intTpl = template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	{{.FieldName}}Raw := r.URL.Query().Get("{{.ParamName}}")
	if r.Method == "POST" {
		{{.FieldName}}Raw = r.FormValue("{{.ParamName}}")
	}

	var {{.FieldName}}Value int

	if {{.IsRequired}} && {{.FieldName}}Raw == "" {
		return fmt.Errorf("{{.ParamName}} must me not empty")
	}

	if {{.FieldName}}Raw != "" {
		{{.FieldName}}ConvValue, {{.FieldName}}Err := strconv.Atoi({{.FieldName}}Raw)
		if {{.FieldName}}Err !=nil {
			return fmt.Errorf("{{.ParamName}} must be int")
		}
		{{.FieldName}}Value = {{.FieldName}}ConvValue
	}

	if {{.HasDefault}} && {{.FieldName}}Raw == "" {
		if "{{.DefaultValue}}" != "" {
			{{.FieldName}}Value, _ = strconv.Atoi("{{.DefaultValue}}")
		}
		
	}

	if {{.HasEnum}} {
		isValid := false

		for _, enumValue := range strings.Split("{{.EnumString}}", "|") {
			enumIntValue, _ := strconv.Atoi(enumValue)
			isValid = isValid || {{.FieldName}}Value == enumIntValue
		}

		if !isValid {
			return fmt.Errorf("{{.ParamName}} must be one of [" + strings.Join(strings.Split("{{.EnumString}}", "|"), ", ") + "]")
		}
	}

	if {{.HasMin}} && {{.FieldName}}Value < {{.Min}} {
		return fmt.Errorf("{{.ParamName}} must be >= {{.Min}}")
	}

	if {{.HasMax}} && {{.FieldName}}Value > {{.Max}} {
		return fmt.Errorf("{{.ParamName}} must be <= {{.Max}}")
	}

	in.{{.FieldName}} = {{.FieldName}}Value
`))

func main() {
	aggr := map[string][]ServeHttpParams{}
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(out, `package `+file.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out, `import "strings"`)
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out) // empty line

	for _, dec := range file.Decls {
		fun, ok := dec.(*ast.FuncDecl)
		if ok {
			generateFunc(fun, out, aggr)
		}

		g, ok := dec.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if currStruct.Fields == nil {
				continue
			}

			if len(currStruct.Fields.List) == 0 || currStruct.Fields.List[0].Tag == nil {
				continue
			}

			needCodegen := strings.Contains(currStruct.Fields.List[0].Tag.Value, "apivalidator:")

			if needCodegen {
				generateStruct(currType, currStruct, out)
			}
		}
	}
	generateServeHttp(aggr, out)
}

func generateServeHttp(aggr map[string][]ServeHttpParams, out *os.File) {
	for k, v := range aggr {
		fmt.Fprintln(out, "func (api *"+k+") ServeHTTP(w http.ResponseWriter, r *http.Request) {")
		fmt.Fprintln(out, "    switch r.URL.Path {")
		for _, m := range v {
			fmt.Fprintln(out, "    case \""+m.Url+"\":")
			fmt.Fprintln(out, "        api."+m.Handler+"(w,r)")
		}
		fmt.Fprintln(out, "    default:")
		fmt.Fprintln(out, "        w.WriteHeader(http.StatusNotFound)")
		fmt.Fprintln(out, "        w.Write([]byte(\"{\\\"error\\\": \\\"unknown method\\\"}\"))")
		fmt.Fprintln(out, "    }")
		fmt.Fprintln(out, "}")
	}
}

func generateStruct(typeSpec *ast.TypeSpec, structType *ast.StructType, out *os.File) {
	fmt.Fprintln(out, "func (in *"+typeSpec.Name.Name+") ApiGenSet(r *http.Request) error {")
	for _, field := range structType.Fields.List {
		if field.Tag == nil {
			continue
		}

		tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
		tagValue := tag.Get("apivalidator")
		if tagValue == "" {
			continue
		}

		fieldName := field.Names[0].Name
		fieldType := field.Type.(*ast.Ident).Name
		rules := extractRules(&tagValue)
		rules.FieldName = fieldName
		if rules.ParamName == "" {
			rules.ParamName = strings.ToLower(fieldName)
		}
		switch fieldType {
		case "string":
			strTpl.Execute(out, rules)
		case "int":
			intTpl.Execute(out, rules)
		default:
			panic("unsupported")
		}
	}
	fmt.Fprintln(out, "    return nil")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out, "")
}

func extractRules(tag *string) MethodSetParams {
	var rules MethodSetParams

	for _, keyValuePair := range strings.Split(*tag, ",") {
		kV := strings.Split(keyValuePair, "=")
		if kV[0] == "required" {
			rules.IsRequired = true
		}

		if kV[0] == "paramname" {
			rules.ParamName = kV[1]
		}

		if kV[0] == "enum" {
			rules.HasEnum = true
			rules.EnumString = kV[1]
		}

		if kV[0] == "default" {
			rules.HasDefault = true
			rules.DefaultValue = kV[1]
		}

		if kV[0] == "min" {
			rules.HasMin = true
			rules.Min, _ = strconv.Atoi(kV[1])
		}

		if kV[0] == "max" {
			rules.HasMax = true
			rules.Max, _ = strconv.Atoi(kV[1])
		}
	}

	return rules
}

func generateFunc(funcDecl *ast.FuncDecl, out *os.File, aggr map[string][]ServeHttpParams) {
	if funcDecl.Doc == nil {
		return
	}

	needCodegen := false
	var jsonStr string
	for _, comment := range funcDecl.Doc.List {
		if strings.HasPrefix(comment.Text, "// apigen:api") {
			jsonStr = strings.TrimPrefix(comment.Text, "// apigen:api")
			needCodegen = true
			break
		}
	}

	if !needCodegen {
		return
	}
	var signature ApiGenSignature
	json.Unmarshal([]byte(jsonStr), &signature)

	var tplParams MethodWrapperParams
	tplParams.ApiType = funcDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	tplParams.ApiGenMethodName = funcDecl.Name.Name
	tplParams.ApiGenHttpMethod = signature.Method
	tplParams.ApiGenParamsType = funcDecl.Type.Params.List[1].Type.(*ast.Ident).Name

	if signature.Auth {
		tplParams.ApiGenAuth = "true"
	} else {
		tplParams.ApiGenAuth = "false"
	}

	wrapperTemplate.Execute(out, tplParams)
	var serveHttp ServeHttpParams
	serveHttp.Url = signature.Url
	serveHttp.Handler = "wrapperApiGen" + tplParams.ApiGenMethodName
	existed, _ := aggr[tplParams.ApiType]
	aggr[tplParams.ApiType] = append(existed, serveHttp)
}
