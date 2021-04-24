package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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

var wrapperTemplate = template.Must(template.New("wrapperTemplate").Parse(`
func (api *{{.ApiType}}) wrapperApiGen{{.ApiGenMethodName}}(w http.ResponseWriter, r *http.Request){
	if r.Method != "{{.ApiGenHttpMethod}}" {
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
	p.fillApiGen(r)
	validateErr := p.validateApiGen()
	if validateErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\" : \"" + validateErr.Error() +"\"}"))
		return
	}
	
	res, err := h.{{.ApiGenMethodName}}(r.Context(), p)

	if err != nil {
		switch v := err.(type){
		case ApiError:
			apiErr := err.(ApiError)
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte("{\"error\" : \"" + apiErr.Err.Error() +"\"}"))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"error\" : \"" + err.Error() +"\"}"))
		}
		return
	}

	data, _ := json.Marshal(res)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
`))

func main() {
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

	for _, dec := range file.Decls {
		fun, ok := dec.(*ast.FuncDecl)
		if !ok {
			continue
		} else {
			generateFunc(fun, out)
		}
	}
}

func generateFunc(funcDecl *ast.FuncDecl, out *os.File) {
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

	if tplParams.ApiGenMethodName == "" {
		tplParams.ApiGenMethodName = "GET"
	}

	if signature.Auth {
		tplParams.ApiGenAuth = "true"
	} else {
		tplParams.ApiGenAuth = "false"
	}

	wrapperTemplate.Execute(out, tplParams)
}
