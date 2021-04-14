package release

// Release represents a single release, with its name, metadata and any categories it is split up by (e.g. 'Added',
// 'Changed')
type Release struct {
	Name       string
	Meta       string
	Categories map[string]string
}
