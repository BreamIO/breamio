package analysis

// The statistics options struct.
type Options struct {
	Screen Dimension
}

// Screen dimensions in pixels
type Dimension struct {
	Width  int `json:width`
	Height int `json:height`
}
