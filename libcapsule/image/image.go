package image

import "time"

type Image struct {
	Id         string
	CreateTime time.Time
	Size       int64
}

func (image *Image) Run() {

}
