package main

import (
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/unixpickle/essentials"
)

//go:generate go run codegen.go

func main() {
	GenerateTransforms()
}

func GenerateTransforms() {
	inPath := filepath.Join("model2d", "transform.go")
	outPath := filepath.Join("model3d", "transform.go")
	log.Println("Creating", outPath, "from", inPath, "...")
	inData, err := ioutil.ReadFile(inPath)
	essentials.Must(err)
	outStr := string(inData)
	outStr = strings.Replace(outStr, "model2d", "model3d", -1)
	outStr = strings.Replace(outStr, "Coord", "Coord3D", -1)
	outStr = strings.Replace(outStr, "Matrix2", "Matrix3", -1)
	outStr = InjectGeneratedComment(outStr, inPath)
	outStr = ApplyAddsAndReplaces(outStr)
	outStr = ReformatCode(outStr)
	essentials.Must(ioutil.WriteFile(outPath, []byte(outStr), 644))
}

func ApplyAddsAndReplaces(data string) string {
	addExpr := regexp.MustCompilePOSIX("[\t ]*// add-codegen: (.*)")
	replaceExpr := regexp.MustCompilePOSIX("[\t ]*// replace-codegen: (.*)")

	lines := strings.Split(data, "\n")
	newLines := []string{}
	for _, line := range lines {
		match := addExpr.FindStringSubmatch(line)
		if match != nil {
			newLines = append(newLines, match[len(match)-1])
			continue
		}
		match = replaceExpr.FindStringSubmatch(line)
		if match != nil {
			newLines[len(newLines)-1] = match[len(match)-1]
			continue
		}
		newLines = append(newLines, line)
	}
	return strings.Join(newLines, "\n")
}

func ReformatCode(code string) string {
	source, err := format.Source([]byte(code))
	essentials.Must(err)
	return string(source)
}

func InjectGeneratedComment(data, sourceFile string) string {
	return "// Generated from " + sourceFile + "\n\n" + data
}
