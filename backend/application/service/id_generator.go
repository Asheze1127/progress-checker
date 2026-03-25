package service

// IDGenerator defines the interface for generating unique IDs.
type IDGenerator interface {
	Generate() string
}
