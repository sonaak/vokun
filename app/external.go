package app

import "github.com/go-zoo/bone"

type external struct {
	newMux func() Mux
}

var __external__ *external = nil

func External() *external {
	if __external__ == nil {
		// create _external_
		__external__ = &external {
			newMux: func() Mux {
				return bone.New()
			},
		}
	}

	return __external__
}
