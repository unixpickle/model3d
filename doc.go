// Package model3d provides a set of APIs for creating,
// manipulating, and storing 3D models.
//
// model3d includes a few sub-packages:
//
//     * render3d - ray tracing, materials, etc.
//     * model2d - 2D graphics, smoothing, bitmaps.
//     * toolbox3d - modular 3D-printable components to
//                   use in larger 3D models.
//
// In addition, model3d comes with a large collection of
// examples for both modeling and rendering.
//
// Representations
//
// Models can be represented in three different ways, and
// model3d can convert between them seamlessly.
//
// In particular, models can be:
//
//     * *Mesh - a triangle mesh, good for exporting to
//               most 3D CAD tools, renderers, etc.
//     * Solid - a 3D boolean function defining which
//               points are contained in the model. Ideal
//               for composition, hand-coding, etc.
//     * Collider - a surface that reports collisions with
//                  rays and other geometric shapes. Ideal
//                  for ray tracing, rendering, and
//                  physics.
//
// Generally, it is easiest to create new models by
// implementing the Solid interface, or by using existing
// solids like *Sphere or *Cylinder and combining them
// with JoinedSolid or SubtractedSolid.
//
// To convert a Solid to a *Mesh, use MarchingCubes() or
// MarchingCubesSearch() for more precision.
// To convert a Solid to a Collider, use SolidCollider or
// simply create a Mesh and convert that to a Collider.
//
// To convert a *Mesh to a Collider, use MeshToCollider().
// To convert a *Mesh to a Solid, first convert it to a
// Collider and then convert that to a Solid.
//
// To convert a Collider to a Solid, use NewColliderSolid()
// or NewColliderSolidHollow().
// To convert a Collider to a *Mesh, the simplest approach
// is to convert it to a Solid and then to convert the
// Solid to a *Mesh.
//
// Creating models
//
// The easiest way to create new models is by defining an
// object that implements the Solid interface.
// Once defined, a Solid can be converted to a Mesh and
// exported to a file (e.g. an STL file for 3D printing).
//
// For example, here's how to implement a sphere as a
// Solid, taken from the actual model3d.Sphere type:
//
//     type Sphere struct {
//         Center Coord3D
//         Radius float64
//     }
//
//     func (s *Sphere) Min() Coord3D {
//         return Coord3D{
//             X: s.Center.X - s.Radius,
//             Y: s.Center.Y - s.Radius,
//             Z: s.Center.Z - s.Radius,
//         }
//     }
//
//     func (s *Sphere) Max() Coord3D {
//         return Coord3D{
//             X: s.Center.X + s.Radius,
//             Y: s.Center.Y + s.Radius,
//             Z: s.Center.Z + s.Radius,
//         }
//     }
//
//     func (s *Sphere) Contains(c Coord3D) bool {
//         return c.Dist(s.Center) <= s.Radius
//     }
//
// Once you have implemented a Solid, you can create a
// mesh and export it to a file like so:
//
//     solid := &Sphere{...}
//     mesh := MarchingCubesSearch(solid, 0.1, 8)
//     mesh.SaveGroupedSTL("output.stl")
//
// In the above example, the mesh is created with an
// epsilon of 0.01 and 8 search steps.
// These parameters control the mesh resolution.
// See MarchingCubesSearch() for more details.
//
// Mesh manipulation
//
// The Mesh type provides various methods to check for
// singularities, fix small holes, eliminate redundant
// triangles, etc.
// There are also APIs that operate on a Mesh in more
// complex ways, making it easier to generate meshes
// programmatically:
//
//     * Decimator - polygon reduction.
//     * MeshSmoother - smoothing for reducing sharp
//                      edges or corners.
//     * Subdivider - edge-based sub-division to add
//                    resolution where it is needed.
//
// Exporting models
//
// Software for 3D printing, rendering, and modeling
// typically expects to import 3D models as triangle
// meshes.
// Thus, model3d provides a number of ways to import and
// export triangle meshes.
// The simplest method is Mesh.SaveGroupedSTL(), which
// exports and STL file to a path.
// For colored models, Mesh.EncodeMaterialOBJ() is the
// method to use.
package model3d
