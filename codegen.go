package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"

	"github.com/unixpickle/essentials"
)

//go:generate go run codegen.go

func main() {
	Generate2d3dTemplate("transform")
	Generate2d3dTemplate("bounder")
	Generate2d3dTemplate("solid")
}

func Generate2d3dTemplate(name string) {
	inPath := filepath.Join("templates", name+".template")
	template, err := template.ParseFiles(inPath)
	essentials.Must(err)
	for _, pkg := range []string{"model2d", "model3d"} {
		outPath := filepath.Join(pkg, name+".go")
		log.Println("Creating", outPath, "...")
		data := RenderTemplate(template, TemplateEnvironment(pkg))
		data = ReformatCode(data)
		data = InjectGeneratedComment(data, inPath)
		essentials.Must(ioutil.WriteFile(outPath, []byte(data), 644))
	}
}

func TemplateEnvironment(pkg string) map[string]interface{} {
	coordType := "Coord"
	matrixType := "Matrix2"
	if pkg == "model3d" {
		coordType = "Coord3D"
		matrixType = "Matrix3"
	}
	return map[string]interface{}{
		"package":    pkg,
		"model2d":    pkg == "model2d",
		"coordType":  coordType,
		"matrixType": matrixType,
	}
}

func RenderTemplate(template *template.Template, data interface{}) string {
	w := bytes.NewBuffer(nil)
	essentials.Must(template.Execute(w, data))
	return string(w.Bytes())
}

func ReformatCode(code string) string {
	source, err := format.Source([]byte(code))
	essentials.Must(err)
	return string(source)
}

func InjectGeneratedComment(data, sourceFile string) string {
	return "// Generated from " + sourceFile + "\n\n" + data
}
