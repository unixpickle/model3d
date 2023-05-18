package fileformats

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"strconv"

	"github.com/unixpickle/essentials"
)

type ThreeMFUnit string

const (
	ThreeMFUnitMicron     ThreeMFUnit = "micron"
	ThreeMFUnitMillimeter             = "millimeter"
	ThreeMFUnitCentimeter             = "centimeter"
	ThreeMFUnitInch                   = "inch"
	ThreeMFUnitFoot                   = "foot"
	ThreeMFUnitMeter                  = "meter"
)

// Write3MFMesh encodes a mesh as a 3MF file.
//
// The mesh is passed as a collection of vertices, and
// triangles which index into this vertex list starting at
// 0 for the first vertex.
func Write3MFMesh(w io.Writer, unit ThreeMFUnit, vertices [][3]float64, triangles [][3]int) (err error) {
	defer essentials.AddCtxTo("write 3MF file", &err)
	vertexElems := make([]threeMFVertex, len(vertices))
	triangleElems := make([]threeMFTriangle, len(triangles))
	for i, v := range vertices {
		vertexElems[i].X = strconv.FormatFloat(v[0], 'f', 32, 64)
		vertexElems[i].Y = strconv.FormatFloat(v[1], 'f', 32, 64)
		vertexElems[i].Z = strconv.FormatFloat(v[2], 'f', 32, 64)
	}
	for i, t := range triangles {
		triangleElems[i].V1 = strconv.Itoa(t[0])
		triangleElems[i].V2 = strconv.Itoa(t[1])
		triangleElems[i].V3 = strconv.Itoa(t[2])
	}
	zipWriter := zip.NewWriter(w)
	modelWriter, err := zipWriter.Create("3D/3dmodel.model")
	if err != nil {
		return err
	}
	if _, err := io.WriteString(modelWriter, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(modelWriter)
	enc.Indent("", "  ")
	err = enc.Encode(threeMFModel{
		Unit:    string(unit),
		XmlLang: "en-US",
		Xmlns:   "http://schemas.microsoft.com/3dmanufacturing/core/2015/02",
		Resources: []threeMFObject{
			{
				ID:   "1",
				Type: "model",
				Mesh: threeMFMesh{
					Vertices:  vertexElems,
					Triangles: triangleElems,
				},
			},
		},
		Build: threeMFBuild{
			Item: []threeMFItem{
				{ObjectID: "1"},
			},
		},
	})
	if err != nil {
		return err
	}
	relsWriter, err := zipWriter.Create("_rels/.rels")
	if err != nil {
		return err
	}
	if _, err := io.WriteString(relsWriter, xml.Header); err != nil {
		return err
	}
	err = xml.NewEncoder(relsWriter).Encode(threeMFRelationships{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
		Relationship: []threeMFRelationship{
			{
				Target: "/3D/3dmodel.model",
				Id:     "rel-1",
				Type:   "http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel",
			},
		},
	})
	if err != nil {
		return err
	}
	typesWriter, err := zipWriter.Create("[Content_Types].xml")
	if err != nil {
		return err
	}
	if _, err := io.WriteString(typesWriter, xml.Header); err != nil {
		return err
	}
	err = xml.NewEncoder(typesWriter).Encode(threeMFTypes{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/content-types",
		Default: []threeMFDefault{
			{Extension: "jpeg", ContentType: "image/jpeg"},
			{Extension: "jpg", ContentType: "image/jpeg"},
			{Extension: "model", ContentType: "application/vnd.ms-package.3dmanufacturing-3dmodel+xml"},
			{Extension: "png", ContentType: "image/png"},
			{Extension: "rels", ContentType: "application/vnd.openxmlformats-package.relationships+xml"},
			{Extension: "texture", ContentType: "application/vnd.ms-package.3dmanufacturing-3dmodeltexture"},
		},
	})
	if err != nil {
		return err
	}
	return zipWriter.Close()
}

type threeMFModel struct {
	XMLName   xml.Name        `xml:"model"`
	Unit      string          `xml:"unit,attr"`
	XmlLang   string          `xml:"xml:lang,attr"`
	Xmlns     string          `xml:"xmlns,attr"`
	Resources []threeMFObject `xml:"resources>object"`
	Build     threeMFBuild    `xml:"build"`
}

type threeMFObject struct {
	ID   string      `xml:"id,attr"`
	Type string      `xml:"type,attr"`
	Mesh threeMFMesh `xml:"mesh"`
}

type threeMFMesh struct {
	Vertices  []threeMFVertex   `xml:"vertices>vertex"`
	Triangles []threeMFTriangle `xml:"triangles>triangle"`
}

type threeMFVertex struct {
	X string `xml:"x,attr"`
	Y string `xml:"y,attr"`
	Z string `xml:"z,attr"`
}

type threeMFTriangle struct {
	V1 string `xml:"v1,attr"`
	V2 string `xml:"v2,attr"`
	V3 string `xml:"v3,attr"`
}

type threeMFBuild struct {
	Item []threeMFItem `xml:"item"`
}

type threeMFItem struct {
	ObjectID string `xml:"objectid,attr"`
}

type threeMFRelationships struct {
	XMLName      xml.Name              `xml:"Relationships"`
	Xmlns        string                `xml:"xmlns,attr"`
	Relationship []threeMFRelationship `xml:"Relationship"`
}

type threeMFRelationship struct {
	Target string `xml:"Target,attr"`
	Id     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
}

type threeMFTypes struct {
	XMLName xml.Name         `xml:"Types"`
	Xmlns   string           `xml:"xmlns,attr"`
	Default []threeMFDefault `xml:"Default"`
}

type threeMFDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}
