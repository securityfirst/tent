package transifex

import (
	"time"
)

type Uploader struct {
	trigger chan struct{}
    tick *time.Ticker
}

func (u *Uploader) Update() {
	select {
	case u.trigger <- struct{}{}:
		u.update()
		<-u.trigger
	default:
		return
	}
}
