package assistant

// DeprecationInfo holds deprecation metadata for a model.
// Zero value represents a non-deprecated model.
type DeprecationInfo struct {
	IsDeprecated bool
	RemovedIn    string // semver version string (e.g. "v2.0.0"), empty if not set
}

func (d DeprecationInfo) Deprecated() bool            { return d.IsDeprecated }
func (d DeprecationInfo) DeprecatedRemovedIn() string { return d.RemovedIn }
