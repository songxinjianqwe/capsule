package image

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type ImageService interface {
	Create(id string, tarPath string) error
	Delete(id string) error
	List() ([]Image, error)
	Get(id string) (Image, error)
}

func NewImageService(runtimeRoot string) (ImageService, error) {
	// 这里是将一个存在的tar文件作为新的rootfs
	imageDir := filepath.Join(runtimeRoot, constant.ImageDir)
	if _, err := os.Stat(imageDir); err != nil {
		if os.IsNotExist(err) {
			logrus.Infof("mkdir imageDir if not exists: %s", imageDir)
			if err := os.MkdirAll(imageDir, 0700); err != nil {
				return nil, exception.NewGenericError(err, exception.ImageServiceError)
			}
		} else {
			return nil, exception.NewGenericError(err, exception.ImageServiceError)
		}
	}
	return &imageService{
		root: imageDir,
	}, nil
}

type imageService struct {
	root string
}

func (service *imageService) Create(id string, tarPath string) (err error) {
	imageDir := filepath.Join(service.root, id)
	if _, err := os.Stat(imageDir); err == nil {
		return exception.NewGenericError(fmt.Errorf("image with id exists: %v", id), exception.ImageIdExistsError)
	} else if !os.IsNotExist(err) {
		return exception.NewGenericError(err, exception.ImageLoadError)
	}
	if err := os.MkdirAll(imageDir, 0700); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			logrus.Warnf("create image error, clean %s dir", imageDir)
			os.RemoveAll(imageDir)
		}
	}()
	file, err := os.Open(tarPath)
	if err != nil {
		return exception.NewGenericError(err, exception.ImageCreateError)
	}
	defer file.Close()
	logrus.Infof("starting to read tar file...")
	command := exec.Command("tar", "-xvf", tarPath, "-C", imageDir)
	if err := command.Run(); err != nil {
		return exception.NewGenericError(err, exception.ImageCreateError)
	}
	logrus.Infof("create image %s succeeded", id)
	return nil
}

func (service *imageService) Delete(id string) error {
	imageDir := filepath.Join(service.root, id)
	if _, err := os.Stat(imageDir); err != nil && os.IsNotExist(err) {
		return exception.NewGenericError(fmt.Errorf("image %s not exists", id), exception.ImageLoadError)
	}
	if err := os.RemoveAll(imageDir); err != nil {
		return err
	}
	return nil
}

func (service *imageService) List() ([]Image, error) {
	var images []Image
	if _, err := os.Stat(service.root); err != nil {
		if os.IsNotExist(err) {
			return images, nil
		}
	}
	list, err := ioutil.ReadDir(service.root)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range list {
		images = append(images, Image{
			Id:         fileInfo.Name(),
			CreateTime: fileInfo.ModTime(),
			Size:       fileInfo.Size(),
		})
	}
	return images, nil
}

func (service *imageService) Get(id string) (Image, error) {
	imageDir := filepath.Join(service.root, id)
	fileInfo, err := os.Stat(imageDir)
	if err != nil && os.IsNotExist(err) {
		return Image{}, exception.NewGenericError(fmt.Errorf("image %s not exists", id), exception.ImageLoadError)
	}
	return Image{
		Id:         fileInfo.Name(),
		CreateTime: fileInfo.ModTime(),
		Size:       fileInfo.Size(),
	}, nil
}
