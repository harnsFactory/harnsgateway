package storage

const (
	storePath = "/var/lib/harnsgateway"
)

func isEphemeralError(err error) bool {
	return false
}
