package fileformats

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type PLYFormat int

const (
	PLYFormatASCII PLYFormat = iota
	PLYFormatBinaryLittle
	PLYFormatBinaryBig
)

type PLYPropertyType string

const (
	PLYPropertyTypeNone    PLYPropertyType = ""
	PLYPropertyTypeChar                    = "char"
	PLYPropertyTypeUchar                   = "uchar"
	PLYPropertyTypeShort                   = "short"
	PLYPropertyTypeUshort                  = "ushort"
	PLYPropertyTypeInt                     = "int"
	PLYPropertyTypeUint                    = "uint"
	PLYPropertyTypeFloat                   = "float"
	PLYPropertyTypeDouble                  = "double"
	PLYPropertyTypeInt8                    = "int8"
	PLYPropertyTypeUint8                   = "uint8"
	PLYPropertyTypeInt16                   = "int16"
	PLYPropertyTypeUint16                  = "uint16"
	PLYPropertyTypeInt32                   = "int32"
	PLYPropertyTypeUint32                  = "uint32"
	PLYPropertyTypeFloat32                 = "float32"
	PLYPropertyTypeFloat64                 = "float64"
)

func (p PLYPropertyType) Validate() error {
	switch p {
	case PLYPropertyTypeChar, PLYPropertyTypeInt8, PLYPropertyTypeUchar, PLYPropertyTypeUint8,
		PLYPropertyTypeShort, PLYPropertyTypeInt16, PLYPropertyTypeUshort, PLYPropertyTypeUint16,
		PLYPropertyTypeInt, PLYPropertyTypeInt32, PLYPropertyTypeUint, PLYPropertyTypeUint32,
		PLYPropertyTypeFloat, PLYPropertyTypeFloat32, PLYPropertyTypeDouble, PLYPropertyTypeFloat64:
		return nil
	default:
		return errors.New("unknown PLY type name: " + string(p))
	}
}

func (p PLYPropertyType) Size() int {
	switch p {
	case PLYPropertyTypeChar, PLYPropertyTypeInt8, PLYPropertyTypeUchar, PLYPropertyTypeUint8:
		return 1
	case PLYPropertyTypeShort, PLYPropertyTypeInt16, PLYPropertyTypeUshort, PLYPropertyTypeUint16:
		return 2
	case PLYPropertyTypeInt, PLYPropertyTypeInt32, PLYPropertyTypeUint, PLYPropertyTypeUint32:
		return 4
	case PLYPropertyTypeFloat, PLYPropertyTypeFloat32:
		return 4
	case PLYPropertyTypeDouble, PLYPropertyTypeFloat64:
		return 8
	default:
		panic("unknown property type: " + p)
	}
}

func (p PLYPropertyType) Format(value interface{}) string {
	switch p {
	case PLYPropertyTypeChar, PLYPropertyTypeInt8:
		return strconv.FormatInt(int64(value.(int8)), 10)
	case PLYPropertyTypeUchar, PLYPropertyTypeUint8:
		return strconv.FormatUint(uint64(value.(uint8)), 10)
	case PLYPropertyTypeShort, PLYPropertyTypeInt16:
		return strconv.FormatInt(int64(value.(int16)), 10)
	case PLYPropertyTypeUshort, PLYPropertyTypeUint16:
		return strconv.FormatUint(uint64(value.(uint16)), 10)
	case PLYPropertyTypeInt, PLYPropertyTypeInt32:
		return strconv.FormatInt(int64(value.(int32)), 10)
	case PLYPropertyTypeUint, PLYPropertyTypeUint32:
		return strconv.FormatUint(uint64(value.(uint32)), 10)
	case PLYPropertyTypeFloat, PLYPropertyTypeFloat32:
		return strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
	case PLYPropertyTypeDouble, PLYPropertyTypeFloat64:
		return strconv.FormatFloat(value.(float64), 'f', -1, 64)
	default:
		panic("unknown property type: " + p)
	}
}

type PLYProperty struct {
	// May be PLYPropertyTypeNone for non-lists.
	LenType PLYPropertyType

	ElemType PLYPropertyType
	Name     string
}

func NewPLYPropertyString(lineData string) (*PLYProperty, error) {
	parts := strings.Fields(strings.TrimSpace(lineData))
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid property: %#v", lineData)
	}
	prop := &PLYProperty{}
	if parts[1] == "list" {
		if len(parts) != 5 {
			return nil, fmt.Errorf("invalid list property: %#v", lineData)
		}
		prop.LenType = PLYPropertyType(parts[2])
		prop.ElemType = PLYPropertyType(parts[3])
		err := prop.LenType.Validate()
		if err == nil {
			err = prop.ElemType.Validate()
		}
		if err != nil {
			return nil, err
		}
		prop.Name = parts[4]
	} else {
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid property: %#v", lineData)
		}
		prop.ElemType = PLYPropertyType(parts[1])
		err := prop.ElemType.Validate()
		if err != nil {
			return nil, err
		}
		prop.Name = parts[2]
	}
	return prop, nil
}

type PLYElement struct {
	Name       string
	Count      int64
	Properties []*PLYProperty
}

func NewPLYElementColoredVertex(count int64) *PLYElement {
	return &PLYElement{
		Name:  "vertex",
		Count: count,
		Properties: []*PLYProperty{
			{Name: "x", ElemType: PLYPropertyTypeFloat},
			{Name: "y", ElemType: PLYPropertyTypeFloat},
			{Name: "z", ElemType: PLYPropertyTypeFloat},
			{Name: "red", ElemType: PLYPropertyTypeUchar},
			{Name: "green", ElemType: PLYPropertyTypeUchar},
			{Name: "blue", ElemType: PLYPropertyTypeUchar},
		},
	}
}

func NewPLYElementFace(count int64) *PLYElement {
	return &PLYElement{
		Name:  "face",
		Count: count,
		Properties: []*PLYProperty{
			{Name: "vertex_index", LenType: PLYPropertyTypeUchar, ElemType: PLYPropertyTypeInt},
		},
	}
}

func (p *PLYElement) Encode() string {
	var header strings.Builder
	header.WriteString(fmt.Sprintf("element %s %d\n", p.Name, p.Count))
	for _, prop := range p.Properties {
		if prop.LenType == PLYPropertyTypeNone {
			header.WriteString(fmt.Sprintf("property %s %s\n", prop.ElemType, prop.Name))
		} else {
			header.WriteString(fmt.Sprintf("property list %s %s %s\n", prop.LenType, prop.ElemType, prop.Name))
		}
	}
	return header.String()
}

type PLYHeader struct {
	Format   PLYFormat
	Elements []*PLYElement
}

func NewPLYHeaderRead(r io.Reader) (*PLYHeader, error) {
	var data []byte
	for {
		var next [1]byte
		if n, err := r.Read(next[:]); n == 0 {
			if errors.Is(err, io.EOF) {
				return nil, errors.Wrap(io.ErrUnexpectedEOF, "read PLY header")
			}
			return nil, errors.Wrap(err, "read PLY header")
		}
		data = append(data, next[0])
		if strings.HasSuffix(string(data), "end_header\n") {
			return NewPLYHeaderDecode(string(data))
		}
	}
}

func NewPLYHeaderDecode(data string) (*PLYHeader, error) {
	lines := strings.Split(data, "\n")
	if len(lines) < 4 {
		return nil, errors.New("decode PLY header: not enough lines")
	}
	if lines[len(lines)-1] != "" || lines[len(lines)-2] != "end_header" {
		return nil, errors.New("decode PLY header: incorrect end lines")
	}
	formatLine := strings.Fields(lines[1])
	if len(formatLine) != 3 || formatLine[0] != "format" || formatLine[2] != "1.0" {
		return nil, errors.New("decode PLY header: unrecognized format line")
	}
	var format PLYFormat
	switch formatLine[1] {
	case "ascii":
		format = PLYFormatASCII
	case "binary_little_endian":
		format = PLYFormatBinaryLittle
	case "binary_big_endian":
		format = PLYFormatBinaryBig
	default:
		return nil, errors.New("decode PLY header: unrecognized format in file: '" +
			formatLine[1] + "'")
	}

	result := &PLYHeader{Format: format}

	var curElement *PLYElement
	for i, line := range lines[2 : len(lines)-2] {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "comment":
		case "element":
			if len(parts) != 3 {
				return nil, fmt.Errorf("decode PLY header: line %d: invalid element command '%s'",
					i+1, line)
			}
			name := parts[1]
			count, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("decode PLY header: line %d: invalid element count: %s",
					i+1, parts[2])
			}
			if curElement != nil {
				result.Elements = append(result.Elements, curElement)
			}
			curElement = &PLYElement{
				Name:  name,
				Count: count,
			}
		case "property":
			if curElement == nil {
				return nil, fmt.Errorf("decode PLY header: line %d: property declared too early",
					i+1)
			}
			prop, err := NewPLYPropertyString(line)
			if err != nil {
				return nil, fmt.Errorf("decode PLY header: line %d: %s", i+1, err.Error())
			}
			curElement.Properties = append(curElement.Properties, prop)
		default:
			return nil, fmt.Errorf("decode PLY header: line %d: unknown command '%s'",
				i+1, parts[0])
		}
	}
	if curElement != nil {
		result.Elements = append(result.Elements, curElement)
	}
	return result, nil
}

func (p *PLYHeader) Encode() string {
	var header strings.Builder
	header.WriteString("ply\n")
	switch p.Format {
	case PLYFormatASCII:
		header.WriteString("format ascii 1.0\n")
	case PLYFormatBinaryLittle:
		header.WriteString("format binary_little_endian 1.0\n")
	case PLYFormatBinaryBig:
		header.WriteString("format binary_big_endian 1.0\n")
	default:
		panic("unknown PLY format: " + strconv.Itoa(int(p.Format)))
	}
	for _, el := range p.Elements {
		header.WriteString(el.Encode())
	}
	header.WriteString("end_header\n")
	return header.String()
}

// A PLYWriter encodes an arbitrary PLY file.
//
// This may use buffering as it writes the file, but the
// full file will always be flushed by the time the last
// element is written.
type PLYWriter struct {
	w *bufio.Writer

	header            PLYHeader
	curElement        int
	curElementWritten int64
}

// NewPLYWriter creates a new PLYWriter and writes the
// file header.
func NewPLYWriter(w io.Writer, h *PLYHeader) (*PLYWriter, error) {
	res := &PLYWriter{
		w:      bufio.NewWriter(w),
		header: *h,
	}
	if _, err := res.w.WriteString(h.Encode()); err != nil {
		return nil, errors.Wrap(err, "write PLY header")
	}
	if err := res.w.Flush(); err != nil {
		return nil, errors.Wrap(err, "write PLY header")
	}
	return res, nil
}

// Write writes a single element row.
//
// It is assumed that the fields are in the correct order
// and in the correct type for the current element.
func (p *PLYWriter) Write(fields []PLYValue) error {
	el, err := p.nextElement()
	if err != nil {
		return errors.Wrap(err, "write PLY element")
	}
	if len(fields) != len(el.Properties) {
		return fmt.Errorf("write PLY element: declared %d properties but writing %d fields",
			len(el.Properties), len(fields))
	}
	switch p.header.Format {
	case PLYFormatASCII:
		parts := make([]string, len(fields))
		for i, value := range fields {
			parts[i] = value.EncodeString()
		}
		_, err = p.w.WriteString(strings.Join(parts, " ") + "\n")
	case PLYFormatBinaryLittle, PLYFormatBinaryBig:
		var encoding binary.ByteOrder
		if p.header.Format == PLYFormatBinaryLittle {
			encoding = binary.LittleEndian
		} else {
			encoding = binary.BigEndian
		}
		for _, value := range fields {
			_, err = p.w.Write(value.EncodeBinary(encoding))
			if err != nil {
				break
			}
		}
	default:
		panic("unknown file format in header")
	}
	if err == nil && p.isDone() {
		err = p.w.Flush()
	}
	if err != nil {
		return errors.Wrap(err, "write PLY element")
	}
	return nil
}

func (p *PLYWriter) nextElement() (*PLYElement, error) {
	if p.curElement >= len(p.header.Elements) {
		return nil, errors.New("wrote too many PLY elements")
	}
	el := p.header.Elements[p.curElement]
	if p.curElementWritten >= el.Count {
		p.curElementWritten = 0
		p.curElement++
		return p.nextElement()
	}
	p.curElementWritten++
	return el, nil
}

func (p *PLYWriter) isDone() bool {
	return p.curElement >= len(p.header.Elements) ||
		(p.curElement == len(p.header.Elements)-1 &&
			p.curElementWritten >= p.header.Elements[p.curElement].Count)
}

// A PLYMeshWriter encodes a colored mesh PLY file.
//
// This may use buffering as it writes the file, but the
// full file will always be flushed by the time the last
// triangle is written.
type PLYMeshWriter struct {
	w *PLYWriter
}

// NewPLYMeshWriter creates a new PLYMeshWriter and writes the
// file header.
func NewPLYMeshWriter(w io.Writer, numCoords, numTris int) (*PLYMeshWriter, error) {
	header := &PLYHeader{
		Format: PLYFormatASCII,
		Elements: []*PLYElement{
			NewPLYElementColoredVertex(int64(numCoords)),
			NewPLYElementFace(int64(numTris)),
		},
	}
	pw, err := NewPLYWriter(w, header)
	if err != nil {
		return nil, err
	}
	return &PLYMeshWriter{w: pw}, nil
}

// WriteCoord writes the next coordinate to the file.
//
// This should be called exactly numCoords times.
func (p *PLYMeshWriter) WriteCoord(c [3]float64, color [3]uint8) error {
	return p.w.Write([]PLYValue{
		PLYValueFloat32{float32(c[0])},
		PLYValueFloat32{float32(c[1])},
		PLYValueFloat32{float32(c[2])},
		PLYValueUint8{color[0]},
		PLYValueUint8{color[1]},
		PLYValueUint8{color[2]},
	})
}

// WriteTriangle writes the next triangle to the file.
//
// This should be called exactly numTris times.
func (p *PLYMeshWriter) WriteTriangle(coords [3]int) (err error) {
	return p.w.Write([]PLYValue{
		PLYValueList{
			Length: PLYValueUint8{uint8(3)},
			Values: []PLYValue{
				PLYValueInt32{int32(coords[0])},
				PLYValueInt32{int32(coords[1])},
				PLYValueInt32{int32(coords[2])},
			},
		},
	})
}
