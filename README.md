# model3d

This is a collection of tools for programmatically creating and manipulating 3D models, with some focus on 3D printing.

# Examples

The [examples](examples) directory has a lot of examples for how to use this library. Most of these examples produce models which can be 3D printed. Lots of these examples include renderings of the models, produced by my ray casting API.

# APIs

The main API in this project is the Solid API. A Solid is a simple interface:

```go
type Solid interface {
	Min() Coord3D
	Max() Coord3D
	Contains(p Coord3D) bool
}
```

In essence, a Solid represents an object as a boolean function. Therefore, it is very easy to compose solids, construct solids with code, construct solids out of images/other models, etc.

Once you have a Solid that you want to print/render, you can convert it into a Mesh using the `SolidToMesh()` API found in [solid_to_mesh.go](solid_to_mesh.go).
