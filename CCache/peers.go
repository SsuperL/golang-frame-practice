package ccache

// PeerGetter ...
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// PeerPicker ...
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}
