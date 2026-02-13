package memory

// Top-level namespace conventions for the memory key hierarchy.
const (
	NamespaceMemory = "memory"
	NamespaceSkills = "skills"
	NamespaceAgents = "agents"
)

// Entry is a key-value pair in the memory namespace. Keys are /-separated
// hierarchical paths and values are raw bytes.
type Entry struct {
	Key   string
	Value []byte
}
