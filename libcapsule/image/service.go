package image

import (
	"archive/tar"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ImageService interface {
	Create(id string, tarPath string) error
	Delete(id string) error
	List() ([]Image, error)
	Get(id string) (*Image, error)
}

func NewImageService(runtimeRoot string) (ImageService, error) {
	// 这里是将一个存在的tar文件作为新的rootfs
	imageDir := filepath.Join(runtimeRoot, constant.ImageDir)
	repositories := make(map[string]string)
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
	// 如果文件存在,那么load
	repositoriesPath := filepath.Join(imageDir, constant.ImageRepositoriesFilename)
	if _, err := os.Stat(repositoriesPath); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filepath.Join(runtimeRoot, constant.ImageRepositoriesFilename))
	if err != nil {

	}
	return &imageService{
		root: imageDir,
	}, nil
}

type imageService struct {
	root         string
	repositories map[string]string
}

func (service *imageService) Create(id string, tarPath string) error {
	imageDir := filepath.Join(service.root, id)
	if _, err := os.Stat(imageDir); err == nil {
		return exception.NewGenericError(fmt.Errorf("container with id exists: %v", id), exception.ImageIdExistsError)
	} else if !os.IsNotExist(err) {
		return exception.NewGenericError(err, exception.ImageLoadError)
	}
	if err := os.MkdirAll(imageDir, 0700); err != nil {
		return err
	}
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := tar.NewReader(file)
	for {
		hdr, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		file, err := os.Create(filepath.Join(imageDir, hdr.Name))
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, reader); err != nil {
			return err
		}
	}
	// 此时解压完毕,然后在repositories.json中加入一条

	return nil
}

func (service *imageService) Delete(id string) error {
	panic("implement me")
}

func (service *imageService) List() ([]Image, error) {
	panic("implement me")
}

func (service *imageService) Get(id string) (*Image, error) {
	panic("implement me")
}
