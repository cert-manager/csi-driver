package v1alpha1

type Volume struct {
	Name string
	ID   string
	Size int64
	Path string

	PodName      string
	PodNamespace string
}
