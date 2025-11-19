module github.com/example/calculator

go 1.21

// For this demo, use local library
// In production, use: require github.com/example/mathutils v1.0.0
replace github.com/example/mathutils => ../library-example

require github.com/example/mathutils v0.0.0-00010101000000-000000000000
