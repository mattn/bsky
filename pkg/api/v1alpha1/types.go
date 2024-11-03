package v1alpha1

const (
	// Group for MLP tasks.
	Group = "bsctl.dev"
	// Version for tasks.
	Version = "v1alpha1"
)

// Metadata holds an optional name of the project.
type Metadata struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	// ResourceVersion is used for optimistic concurrency.
	// Ref: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
	// This should be treated as an opaque value by clients.
	ResourceVersion string `yaml:"resourceVersion,omitempty"`
}
