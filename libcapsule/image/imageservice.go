package image

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/facade"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type ImageService interface {
	Create(id string, tarPath string) error
	Delete(id string) error
	List() ([]Image, error)
	Get(id string) (Image, error)
	Run(imageRunArgs *ImageRunArgs) error
	Destroy(containerId string) error
}

func NewImageService(runtimeRoot string) (ImageService, error) {
	factory, err := libcapsule.NewFactory(runtimeRoot, true)
	if err != nil {
		return nil, err
	}
	// 这里是将一个存在的tar文件作为新的rootfs
	imageRoot := filepath.Join(runtimeRoot, constant.ImageDir)
	if _, err := os.Stat(imageRoot); err != nil {
		if os.IsNotExist(err) {
			//logrus.Infof("mkdir generateImageDir if not exists: %s", imageRoot)
			if err := os.MkdirAll(imageRoot, 0700); err != nil {
				return nil, exception.NewGenericError(err, exception.ImageServiceError)
			}
		} else {
			return nil, exception.NewGenericError(err, exception.ImageServiceError)
		}
	}
	repositoriesPath := filepath.Join(imageRoot, constant.ImageRepositoriesFilename)
	repositories := make(map[string]string)
	if _, err := os.Stat(repositoriesPath); err != nil {
		if !os.IsNotExist(err) {
			// 如果文件存在,但stat返回错误,则退出
			return nil, exception.NewGenericError(err, exception.ImageServiceError)
		}
		// 文件不存在,则不动
	} else {
		// 文件存在,则读取
		bytes, err := ioutil.ReadFile(repositoriesPath)
		if err != nil {
			return nil, exception.NewGenericError(err, exception.ImageServiceError)
		}
		if err := json.Unmarshal(bytes, &repositories); err != nil {
			return nil, exception.NewGenericError(err, exception.ImageServiceError)
		}
	}
	//logrus.Infof("loaded repositories.json: %#v", repositories)
	return &imageService{
		factory:      factory,
		imageRoot:    imageRoot,
		repositories: repositories,
	}, nil
}

type imageService struct {
	mutex     sync.Mutex
	factory   libcapsule.Factory
	imageRoot string
	// key -> image id
	// value -> layer id
	repositories map[string]string
}

func (service *imageService) Destroy(containerId string) error {
	// 删除layer
	panic("implement me")
}

func (service *imageService) Run(imageRunArgs *ImageRunArgs) (err error) {
	// 首先要准备spec
	// 1. 后面会添加一个/etc/hosts, /etc/resolv.conf
	// 2. mount
	// 3. 创建spec
	if exists := service.factory.Exists(imageRunArgs.ContainerId); exists {
		return exception.NewGenericError(fmt.Errorf("container already exists: %s", imageRunArgs.ContainerId), exception.ContainerIdExistsError)
	}
	bundle := filepath.Join(service.factory.GetRuntimeRoot(), constant.ContainerDir, imageRunArgs.ContainerId)
	// 创建一个write layer
	var rootfsPath string
	var spec *specs.Spec
	rootfsPath, err = service.prepareUnionFs(imageRunArgs.ImageId)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			logrus.Warnf("run container in image way failed, clean union fs")

		}
	}()
	spec, err = service.prepareBundle(rootfsPath, bundle, imageRunArgs)
	if err != nil {
		return err
	}
	if err = facade.CreateOrRunContainer(service.factory.GetRuntimeRoot(), imageRunArgs.ContainerId, bundle, spec, facade.ContainerActRun, imageRunArgs.Detach, imageRunArgs.Network, imageRunArgs.PortMappings); err != nil {
		return err
	}
	return nil
}

func (service *imageService) prepareUnionFs(image string) (string, error) {
	return "", nil
}

func (service *imageService) prepareBundle(rootfsPath string, bundle string, imageRunArgs *ImageRunArgs) (*specs.Spec, error) {
	spec := buildSpec(rootfsPath, imageRunArgs.Args, imageRunArgs.Env, imageRunArgs.Cwd, imageRunArgs.Hostname, imageRunArgs.Cpushare, imageRunArgs.Memory, imageRunArgs.Annotations)
	if err := os.MkdirAll(bundle, 0644); err != nil {
		return nil, err
	}
	specFile, err := os.OpenFile(filepath.Join(bundle, constant.ContainerConfigFilename), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, exception.NewGenericError(err, exception.SpecSaveError)
	}
	defer specFile.Close()
	bytes, err := json.Marshal(spec)
	if err != nil {
		return nil, exception.NewGenericError(err, exception.SpecSaveError)
	}
	if _, err := specFile.Write(bytes); err != nil {
		return nil, exception.NewGenericError(err, exception.SpecSaveError)
	}
	return spec, nil
}

func (service *imageService) generateImageDir(id string) string {
	return filepath.Join(service.imageRoot, constant.ImageLayersDir, service.repositories[id])
}

func (service *imageService) flushRepositories() error {
	file, err := os.OpenFile(filepath.Join(service.imageRoot, constant.ImageRepositoriesFilename), os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return exception.NewGenericError(err, exception.ImageRepositoriesDumpError)
	}
	defer file.Close()
	bytes, err := json.Marshal(service.repositories)
	if err != nil {
		return exception.NewGenericError(err, exception.ImageRepositoriesDumpError)
	}
	if _, err := file.Write(bytes); err != nil {
		return exception.NewGenericError(err, exception.ImageRepositoriesDumpError)
	}
	return nil
}

func (service *imageService) Create(id string, tarPath string) (err error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if _, exist := service.repositories[id]; exist {
		return exception.NewGenericError(fmt.Errorf("image with id exists: %v", id), exception.ImageIdExistsError)
	}
	uuids, err := uuid.NewV4()
	if err != nil {
		return exception.NewGenericError(err, exception.ImageCreateError)
	}
	layerId := uuids.String()
	service.repositories[id] = layerId
	// /var/run/capsule/images/layers/$layerId
	imageDir := service.generateImageDir(id)
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
	if err := service.flushRepositories(); err != nil {
		return err
	}
	return nil
}

func (service *imageService) Delete(id string) error {
	service.mutex.Lock()
	service.mutex.Unlock()
	if _, exist := service.repositories[id]; !exist {
		return exception.NewGenericError(fmt.Errorf("image %s not exists: %v", id), exception.ImageLoadError)
	}

	imageDir := service.generateImageDir(id)
	if _, err := os.Stat(imageDir); err != nil && os.IsNotExist(err) {
		return exception.NewGenericError(fmt.Errorf("image %s not exists", id), exception.ImageLoadError)
	}
	if err := os.RemoveAll(imageDir); err != nil {
		return err
	}
	delete(service.repositories, id)
	if err := service.flushRepositories(); err != nil {
		return err
	}
	return nil
}

func (service *imageService) List() ([]Image, error) {
	service.mutex.Lock()
	service.mutex.Unlock()
	var images []Image
	for id := range service.repositories {
		fileInfo, err := os.Stat(service.generateImageDir(id))
		if err != nil {
			return nil, exception.NewGenericError(err, exception.ImageLoadError)
		}
		images = append(images, Image{
			Id:         id,
			LayerId:    fileInfo.Name(),
			CreateTime: fileInfo.ModTime(),
			Size:       fileInfo.Size(),
		})
	}
	return images, nil
}

func (service *imageService) Get(id string) (Image, error) {
	service.mutex.Lock()
	service.mutex.Unlock()
	if _, exist := service.repositories[id]; !exist {
		return Image{}, exception.NewGenericError(fmt.Errorf("image %s not exists", id), exception.ImageLoadError)
	}
	imageDir := service.generateImageDir(id)
	fileInfo, err := os.Stat(imageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return Image{}, exception.NewGenericError(fmt.Errorf("image %s not exists", id), exception.ImageLoadError)
		} else {
			return Image{}, exception.NewGenericError(err, exception.ImageLoadError)
		}
	}
	return Image{
		Id:         id,
		LayerId:    fileInfo.Name(),
		CreateTime: fileInfo.ModTime(),
		Size:       fileInfo.Size(),
	}, nil
}
