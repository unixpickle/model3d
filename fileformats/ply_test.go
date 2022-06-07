package fileformats

import (
	"bytes"
	"testing"
)

func TestPLYMeshWriter(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewPLYMeshWriter(&buf, 4, 2)
	if err != nil {
		t.Fatal(err)
	}
	coords := [][3]float64{
		{0, 0, 0},
		{1, 0, 0},
		{1, 1, 0},
		{0, 1, 0},
	}
	indices := [][3]int{{0, 1, 2}, {0, 2, 3}}
	for _, coord := range coords {
		err := w.WriteCoord(coord, [3]uint8{1, 2, 3})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, tri := range indices {
		err := w.WriteTriangle(tri)
		if err != nil {
			t.Fatal(err)
		}
	}

	actual := buf.String()
	expected := "ply\nformat ascii 1.0\nelement vertex 4\nproperty float x\nproperty float y\nproperty float z\nproperty uchar red\nproperty uchar green\nproperty uchar blue\nelement face 2\nproperty list uchar int vertex_index\nend_header\n0 0 0 1 2 3\n1 0 0 1 2 3\n1 1 0 1 2 3\n0 1 0 1 2 3\n3 0 1 2\n3 0 2 3\n"
	if actual != expected {
		t.Errorf("unexpected output: %#v", actual)
	}
}

func TestPLYHeaderDecode(t *testing.T) {
	data := "ply\nformat ascii 1.0\nelement vertex 4\nproperty float x\nproperty float y\nproperty float z\nproperty uchar red\nproperty uchar green\nproperty uchar blue\nelement face 2\nproperty list uchar int vertex_index\nend_header\n0 0 0 1 2 3\n1 0 0 1 2 3\n1 1 0 1 2 3\n0 1 0 1 2 3\n3 0 1 2\n3 0 2 3\n"
	buf := bytes.NewReader([]byte(data))
	dec, err := NewPLYHeaderRead(buf)
	if err != nil {
		t.Fatal(err)
	}
	if dec.Format != PLYFormatASCII {
		t.Fatalf("unexpected format: %v", dec.Format)
	}
	if len(dec.Elements) != 2 {
		t.Fatalf("unexpected elements: %#v", dec.Elements)
	}
	if dec.Elements[1].Name != "face" {
		t.Errorf("unexpected second name: %s", dec.Elements[1].Name)
	}
	if len(dec.Elements[1].Properties) != 1 {
		t.Errorf("unexpected second length: %d", len(dec.Elements[1].Properties))
	}
	actual := dec.Encode()
	expected := "ply\nformat ascii 1.0\nelement vertex 4\nproperty float x\nproperty float y\nproperty float z\nproperty uchar red\nproperty uchar green\nproperty uchar blue\nelement face 2\nproperty list uchar int vertex_index\nend_header\n"
	if actual != expected {
		t.Errorf("unexpected re-encoded header: %#v", actual)
	}
}

func TestPLYReader(t *testing.T) {
	testFormat := func(t *testing.T, f PLYFormat) {
		var buf bytes.Buffer
		header := PLYHeader{
			Format: f,
			Elements: []*PLYElement{
				NewPLYElementColoredVertex(4),
				NewPLYElementFace(2),
			},
		}
		w, err := NewPLYWriter(&buf, &header)
		if err != nil {
			t.Fatal(err)
		}
		c := PLYValueUint8{0x13}
		values := [][]PLYValue{
			{PLYValueFloat32{0}, PLYValueFloat32{0}, PLYValueFloat32{0}, c, c, c},
			{PLYValueFloat32{1}, PLYValueFloat32{0}, PLYValueFloat32{0}, c, c, c},
			{PLYValueFloat32{1}, PLYValueFloat32{1}, PLYValueFloat32{0}, c, c, c},
			{PLYValueFloat32{0}, PLYValueFloat32{1}, PLYValueFloat32{0}, c, c, c},
			{
				PLYValueList{
					Length: PLYValueUint8{uint8(3)},
					Values: []PLYValue{
						PLYValueInt32{int32(0)},
						PLYValueInt32{int32(1)},
						PLYValueInt32{int32(2)},
					},
				},
			},
			{
				PLYValueList{
					Length: PLYValueUint8{uint8(3)},
					Values: []PLYValue{
						PLYValueInt32{int32(0)},
						PLYValueInt32{int32(2)},
						PLYValueInt32{int32(3)},
					},
				},
			},
		}
		for _, row := range values {
			err := w.Write(row)
			if err != nil {
				t.Fatal(err)
			}
		}

		r, err := NewPLYReader(&buf)
		if err != nil {
			t.Fatal(err)
		}
		for i, expectedValue := range values {
			actualValue, _, err := r.Read()
			if err != nil {
				t.Fatal(err)
			}
			if len(expectedValue) != len(actualValue) {
				t.Fatalf("row %d: expected %#v but got %#v", i, expectedValue, actualValue)
			}
			for j, x := range expectedValue {
				a := actualValue[j]
				xString := x.EncodeString()
				aString := a.EncodeString()
				if xString != aString {
					t.Fatalf("row %d index %d: expected %#v but got %#v", i, j, xString, aString)
				}
			}
		}
	}
	t.Run("ASCII", func(t *testing.T) {
		testFormat(t, PLYFormatASCII)
	})
	t.Run("Little", func(t *testing.T) {
		testFormat(t, PLYFormatBinaryLittle)
	})
	t.Run("Big", func(t *testing.T) {
		testFormat(t, PLYFormatBinaryBig)
	})

}
