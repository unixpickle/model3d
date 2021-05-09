package main

import (
	"bytes"
	"errors"
	"flag"
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"

	"github.com/unixpickle/essentials"
)

//go:generate go run codegen.go

func main() {
	var checkNoChange bool
	flag.BoolVar(&checkNoChange, "check", false, "if true, assert that nothing changes")
	flag.Parse()

	Generate2d3dTemplate("transform", checkNoChange)
	Generate2d3dTemplate("bounder", checkNoChange)
	Generate2d3dTemplate("solid", checkNoChange)
	Generate2d3dTemplate("mesh", checkNoChange)
	Generate2d3dTemplate("mesh_hierarchy", checkNoChange)
	Generate2d3dTemplate("mesh_hierarchy_test", checkNoChange)
	Generate2d3dTemplate("bvh", checkNoChange)
	Generate2d3dTemplate("polytope", checkNoChange)
	Generate2d3dTemplate("polytope_test", checkNoChange)
	Generate2d3dTemplate("util_test", checkNoChange)
	Generate2d3dTemplate("sdf", checkNoChange)
	Generate2d3dTemplate("coord_tree", checkNoChange)
	Generate2d3dTemplate("coord_tree_test", checkNoChange)
	Generate2d3dTemplate("fast_maps", checkNoChange)
	Generate2d3dTemplate("fast_maps_test", checkNoChange)
}

func Generate2d3dTemplate(name string, checkNoChange bool) {
	inPath := filepath.Join("templates", name+".template")
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"mkargs": func(base map[string]interface{}, vs ...string) (map[string]interface{}, error) {
			if len(vs)%2 != 0 {
				return nil, errors.New("mismatched keys and values")
			}
			res := map[string]interface{}{}
			for k, v := range base {
				res[k] = v
			}
			for i := 0; i < len(vs); i += 2 {
				res[vs[i]] = vs[i+1]
			}
			return res, nil
		},
		"cat": func(strs ...string) string {
			res := ""
			for _, s := range strs {
				res += s
			}
			return res
		},
	})
	data, err := ioutil.ReadFile(inPath)
	essentials.Must(err)
	_, err = tmpl.Parse(string(data))
	essentials.Must(err)
	for _, pkg := range []string{"model2d", "model3d"} {
		outPath := filepath.Join(pkg, name+".go")
		log.Println("Creating", outPath, "...")
		data := RenderTemplate(tmpl, TemplateEnvironment(pkg))
		data = ReformatCode(data)
		data = InjectGeneratedComment(data, inPath)
		if checkNoChange {
			oldData, err := ioutil.ReadFile(outPath)
			essentials.Must(err)
			if !bytes.Equal(oldData, []byte(data)) {
				essentials.Die("File changed, check failed!")
			}
		} else {
			essentials.Must(ioutil.WriteFile(outPath, []byte(data), 0644))
		}
	}
}

func TemplateEnvironment(pkg string) map[string]interface{} {
	coordType := "Coord"
	matrixType := "Matrix2"
	faceType := "Segment"
	faceName := "segment"
	numDims := 2
	if pkg == "model3d" {
		coordType = "Coord3D"
		matrixType = "Matrix3"
		faceType = "Triangle"
		faceName = "triangle"
		numDims = 3
	}
	return map[string]interface{}{
		"package":    pkg,
		"model2d":    pkg == "model2d",
		"coordType":  coordType,
		"matrixType": matrixType,
		"faceType":   faceType,
		"faceName":   faceName,
		"numDims":    numDims,
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
