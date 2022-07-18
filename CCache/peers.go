package ccache

import (
	"ccache/ccachepb"
)

// PeerGetter ...
type PeerGetter interface {
	Get(*ccachepb.Request) (*ccachepb.Response, error)
}

// PeerPicker ...
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}
