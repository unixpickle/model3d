// Package model3d provides a set of APIs for creating,
// manipulating, and storing 3D models.
//
// This includes a render3d sub-package for rendering,
// a model2d package for creating 2-dimensional models,
// and a toolbox3d package for creating usable parts for
// 3D printing.
// In additional, model3d comes with a large collection of
// practical examples.
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
// The Mesh type provides various methods to check for
// singularities, fix small holes, eliminate redundant
// triangles, etc.
// There are also APIs that operate on a Mesh in more
// complex ways, making it easier to generate meshes
// programmatically:
//
//     * Decimator - polygon reduction.
//     * Smoother - smoothing for reducing sharp edges or
//                  corners.
//     * Subdivider - edge-based sub-division to add
//                    resolution where it is needed.
//
package model3d
