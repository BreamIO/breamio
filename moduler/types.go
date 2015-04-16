package moduler

type Runner interface {
	Run(Logic)
}

// A specification for creation of new objects.
// Type should be a type available for creation by the logic implementation.
// Data is a context sensitive string, which syntax depends on the type.
// Emitter is a integer, identifying the emitter number to link the new object to.
type Spec struct {
	Emitter int
	Data    map[string]interface{}
}
