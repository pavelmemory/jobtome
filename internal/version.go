package internal

var (
	CommitSHA      = "N/A"
	BuildTimestamp = "N/A"
	Version        = "N/A"
)

var versionInfo = struct {
	Version        string `json:"version"`
	CommitSHA      string `json:"commit_sha"`
	BuildTimestamp string `json:"build_timestamp"`
}{
	Version: Version, CommitSHA: CommitSHA, BuildTimestamp: BuildTimestamp,
}

// WriteVersion writes build information about binary into provided encoder.
func WriteVersion(encoder interface{ Encode(interface{}) error }) error {
	// we are pretty sure it would be convertible into JSON
	return encoder.Encode(versionInfo)
}
