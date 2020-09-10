# plot-2d-sdf

This example plots a two-dimensional signed distance field for any simple 2D shape (encoded as an image). The signed distance field is plotted using a color code, so that blue is positive and red is negative; intensity is used to indicate magnitude.

This is intended to help visualize how the signed distance field behaves inside of a shape: where it changes derivatives, sign, magnitude, etc. It also helps visualize local minima and saddle points in the SDF landscape.

# Output

Here is an example output, which takes in the heart image on the left, and produces the distance field on the right:

![Example rendering](example.png)
