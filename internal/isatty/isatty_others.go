// +build appengine solaris js ppc64 ppc64le

package isatty

// IsTerminal returns true if the file descriptor is terminal which
// is always false on js and appengine classic which is a sandboxed PaaS.
func IsTerminal(fd uintptr) bool {
	return false
}
