package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"

	"github.com/unixpickle/essentials"
)

//go:generate go run codegen.go

func main() {
	GenerateTransforms()
}

func GenerateTransforms() {
	inPath := filepath.Join("templates", "transform.template")
	template, err := template.ParseFiles(inPath)
	essentials.Must(err)
	for _, pkg := range []string{"model2d", "model3d"} {
		outPath := filepath.Join(pkg, "transform.go")
		log.Println("Creating", outPath, "...")
		coordType := "Coord"
		matrixType := "Matrix2"
		if pkg == "model3d" {
			coordType = "Coord3D"
			matrixType = "Matrix3"
		}
		data := RenderTemplate(template, map[string]interface{}{
			"model2d":    pkg == "model2d",
			"coordType":  coordType,
			"matrixType": matrixType,
		})
		if pkg == "model3d" {
			fmt.Println(data)
		}
		data = ReformatCode(data)
		data = InjectGeneratedComment(data, inPath)
		essentials.Must(ioutil.WriteFile(outPath, []byte(data), 644))
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
