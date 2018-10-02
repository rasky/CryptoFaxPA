/* CryptoFaxPA
   License: AGPLv3 */

SIZE = [90, 123, 90];
ANGLE_RADIUS = 15;
BOTTOM_THICKNESS = 2;
LID_THICKNESS = 2;
WALL_THICKNESS = 3;
SHELF_DEPTH = 3;
BUTTONS_DIAMETER = 12;

box();

translate([SIZE[0] + 15, 0, 0])
  lid();

module lid() {
  difference() {
     // main plate
     linear_extrude(height = LID_THICKNESS)
        offset(r = -WALL_THICKNESS)
          rounded_square(SIZE, ANGLE_RADIUS);
      
      // button holes
      for (o = [[33,19,-0.1],[57,19,-0.1]])
        translate(o)
          cylinder(d = BUTTONS_DIAMETER, h = LID_THICKNESS+0.2);
      
      // printer hole
      translate([8.4,36.3,-0.1])
        linear_extrude(height = LID_THICKNESS+0.2)
          rounded_square([74,74], 5);
  };
  
  // top of the rear opening
  translate([ANGLE_RADIUS, SIZE[1]-WALL_THICKNESS-0.1, 0])
    cube([SIZE[0] - ANGLE_RADIUS*2, WALL_THICKNESS+0.1, BOTTOM_THICKNESS]);
}

module box() {
  difference() {
    // external shape
    linear_extrude(height = SIZE[2])
      rounded_square(SIZE, ANGLE_RADIUS);
    
    // internal void
    translate([0,0,BOTTOM_THICKNESS])
      linear_extrude(height = SIZE[2])
        offset(r = -WALL_THICKNESS)
          rounded_square(SIZE, ANGLE_RADIUS);
    
    // rear window
    translate([15,SIZE[1]-WALL_THICKNESS-0.1,BOTTOM_THICKNESS])
      cube([SIZE[0] - ANGLE_RADIUS*2, 10, SIZE[2]]);
    
    // USB port hole
    translate([SIZE[0]-WALL_THICKNESS-0.1, 17, 9.3])
      cube([WALL_THICKNESS+0.2, 16, 9]);
    
    // Raspberry PI mounting holes
    for (o = [[0,0], [49,0], [49,58], [0,58]])
      translate(o)
        translate([19,18,-0.1])
          cylinder(d = 3, h = BOTTOM_THICKNESS+0.2);
  };
  
  // lid shelves
  translate([WALL_THICKNESS, ANGLE_RADIUS, SIZE[2]-LID_THICKNESS])
    lid_mount(SHELF_DEPTH, SIZE[1] - ANGLE_RADIUS*2);
  translate([SIZE[0]-WALL_THICKNESS, SIZE[1] - ANGLE_RADIUS, SIZE[2]-LID_THICKNESS])
    rotate(180)
       lid_mount(SHELF_DEPTH, SIZE[1] - ANGLE_RADIUS*2);
}
module rounded_square(s, r) {
    translate([r,r])
      minkowski() {
          square([s[0] - 2*r, s[1] - 2*r]);
          circle(r);
      }
}

module lid_mount(d, l) {
   translate([0,l,-d])
     rotate([90,0,0])
        linear_extrude(height = l)
           polygon(points = [[0,d], [d,d], [0,0]]);
}
