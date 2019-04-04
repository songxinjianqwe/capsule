package image

import "time"

type Image struct {
	Id         string
	CreateTime time.Time
}

func (image *Image) Run() {

}
