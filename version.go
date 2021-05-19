package sunshine

// version could be overriden by ldflag during build time.
var version = "development"

// Version returns set version information during build.
func Version() string {
	return version
}
