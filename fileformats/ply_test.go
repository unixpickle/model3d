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
