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
	"strings"
	"sync"
)

type layerType string

const (
	readOnlyLayer  layerType = "read_only"
	readWriteLayer           = "read_write"
	initLayer                = "init"
)

type ImageService interface {
	Create(id string, tarPath string) error
	Delete(id string) error
	List() ([]Image, error)
	Get(id string) (Image, error)
	Run(imageRunArgs *ImageRunArgs) error
	Destroy(container libcapsule.Container) error
}

func NewImageService(runtimeRoot string) (ImageService, error) {
	factory, err := libcapsule.NewFactory(runtimeRoot, true)
	if err != nil {
		return nil, err
	}
	imageRoot := filepath.Join(runtimeRoot, constant.ImageDir)
	if _, err := os.Stat(imageRoot); err != nil {
		if os.IsNotExist(err) {
			logrus.Infof("mkdir generateLayerPath if not exists: %s", imageRoot)
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

func (service *imageService) Destroy(container libcapsule.Container) (err error) {
	// 删除layer
	if err = container.Destroy(); err != nil {
		logrus.Warnf(err.Error())
	}
	if err = service.cleanContainer(container.ID()); err != nil {
		logrus.Warnf(err.Error())
	}
	return err
}

func (service *imageService) cleanContainer(containerId string) (err error) {
	// umount init layer
	// 删除 /var/run/capsule/images/layers/$init_layer
	// 删除 /var/run/capsule/images/layers/$read_write_layer
	// 删除 /var/run/capsule/images/mounts/$container_id
	// 删除 /var/run/capsule/images/containers/$container_id
	// 1. umount
	initLayer := filepath.Join(service.imageRoot, constant.ImageMountsDir, containerId, initLayer)
	var initLayerIdBytes []byte
	if initLayerIdBytes, err = ioutil.ReadFile(initLayer); err != nil || len(initLayerIdBytes) == 0 {
		logrus.Warnf("read init layer id failed, skip clean init layer")
		logrus.Warnf(err.Error())
	} else {
		initLayerPath := filepath.Join(service.imageRoot, constant.ImageLayersDir, string(initLayerIdBytes))
		logrus.Infof("umount %s", initLayerPath)
		cmd := exec.Command("umount", initLayerPath)
		if err := cmd.Run(); err != nil {
			logrus.Warnf("unmount failed, cause: %s", err.Error())
		}
		// 2. 删除init layer
		logrus.Infof("removing container init layer data, layer id is %s", string(initLayerIdBytes))
		if err := os.RemoveAll(initLayerPath); err != nil {
			logrus.Warnf(err.Error())
		}
	}

	// 3. 删除read write layer
	readWriteLayer := filepath.Join(service.imageRoot, constant.ImageMountsDir, containerId, readWriteLayer)
	var rwLayerIdBytes []byte
	if rwLayerIdBytes, err = ioutil.ReadFile(readWriteLayer); err != nil || len(rwLayerIdBytes) == 0 {
		logrus.Warnf("read read write layer id failed, skip clean read write layer")
		logrus.Warnf(err.Error())
	} else {
		logrus.Infof("removing container read write layer data, layer id is %s", string(rwLayerIdBytes))
		if err := os.RemoveAll(filepath.Join(service.imageRoot, constant.ImageLayersDir, string(rwLayerIdBytes))); err != nil {
			logrus.Warnf(err.Error())
		}
	}

	// 4. 删除container mount数据
	containerMountPath := filepath.Join(service.imageRoot, constant.ImageContainersDir, containerId)
	logrus.Infof("removing container mount path: %s", containerMountPath)
	if err := os.RemoveAll(containerMountPath); err != nil {
		logrus.Warnf(err.Error())
	}

	// 5. 删除container config数据
	containerConfigPath := filepath.Join(service.imageRoot, constant.ImageMountsDir, containerId)
	logrus.Infof("removing container config path: %s", containerConfigPath)
	if err := os.RemoveAll(containerConfigPath); err != nil {
		logrus.Warnf(err.Error())
	}
	return err
}

func (service *imageService) Run(imageRunArgs *ImageRunArgs) (err error) {
	// 1. 检查是否已经存在该容器
	if exists := service.factory.Exists(imageRunArgs.ContainerId); exists {
		return exception.NewGenericError(fmt.Errorf("container already exists: %s", imageRunArgs.ContainerId), exception.ContainerIdExistsError)
	}
	// 2. 创建bundle目录
	// /var/run/capsule/images/containers/$container_id
	bundle := filepath.Join(service.imageRoot, constant.ImageContainersDir, imageRunArgs.ContainerId)
	if _, err := os.Stat(bundle); err != nil && !os.IsNotExist(err) {
		return exception.NewGenericError(err, exception.ContainerIdExistsError)
	}
	if err := os.MkdirAll(bundle, 0644); err != nil {
		return exception.NewGenericError(err, exception.BundleCreateError)
	}
	defer func() {
		if err != nil {
			logrus.Warnf("imageService#Run failed, clean data")
			if cleanErr := service.cleanContainer(imageRunArgs.ContainerId); cleanErr != nil {
				logrus.Warnf(cleanErr.Error())
			}
		}
	}()
	var rootfsPath string
	var spec *specs.Spec
	// 3. 准备/etc/hosts,会在/var/run/capsule/images/containers/$container_id下创建一个hosts
	hostsMount, err := service.prepareHosts(imageRunArgs.ContainerId, imageRunArgs.Links)
	if err != nil {
		return err
	}

	// 4. 准备/etc/resolv.conf,会在/var/run/capsule/images/containers/$container_id下创建一个resolv.conf
	dnsMount, err := service.prepareDns(imageRunArgs.ContainerId)
	if err != nil {
		return err
	}

	// 5. 准备volume
	volumeMounts, err := service.prepareVolumes(imageRunArgs.Volumes)
	if err != nil {
		return err
	}

	// 6. 准备rootfs
	if rootfsPath, err = service.prepareUnionFs(imageRunArgs.ContainerId, imageRunArgs.ImageId); err != nil {
		return err
	}

	// 7. 准备spec
	specMounts := []specs.Mount{hostsMount, dnsMount}
	specMounts = append(specMounts, volumeMounts...)
	if spec, err = service.prepareSpec(rootfsPath, bundle, imageRunArgs, specMounts); err != nil {
		return err
	}

	// 8. 运行容器,如果运行出错,或者前台运行正常退出,则清理
	if err = facade.CreateOrRunContainer(service.factory.GetRuntimeRoot(), imageRunArgs.ContainerId, bundle, spec, facade.ContainerActRun, imageRunArgs.Detach, imageRunArgs.Network, imageRunArgs.PortMappings); err != nil {
		if cleanErr := service.cleanContainer(imageRunArgs.ContainerId); cleanErr != nil {
			logrus.Warnf(cleanErr.Error())
		}
		return err
	}
	if !imageRunArgs.Detach {
		if cleanErr := service.cleanContainer(imageRunArgs.ContainerId); cleanErr != nil {
			logrus.Warnf(cleanErr.Error())
		}
	}
	return nil
}

func (service *imageService) prepareUnionFs(containerId string, imageId string) (string, error) {
	// 1. 拿到read only layer path, 并将其作为容器的read only layer
	if _, exists := service.repositories[imageId]; !exists {
		return "", exception.NewGenericError(fmt.Errorf("image %s not exists", imageId), exception.ImageIdNotExistsError)
	}
	roLayerPath := service.generateLayerPath(imageId)
	if _, err := os.Stat(roLayerPath); err != nil {
		return "", exception.NewGenericError(err, exception.ImageIdNotExistsError)
	}
	if _, err := service.prepareMountPath(containerId, service.repositories[imageId], readOnlyLayer); err != nil {
		return "", exception.NewGenericError(err, exception.UnionFsError)
	}
	// 2. 创建read write layer
	rwUuids, err := uuid.NewV4()
	if err != nil {
		return "", exception.NewGenericError(err, exception.UnionFsError)
	}
	rwLayerId := rwUuids.String()
	rwLayerPath, err := service.prepareMountPath(containerId, rwLayerId, readWriteLayer)
	if err != nil {
		return "", exception.NewGenericError(err, exception.UnionFsError)
	}

	// 3. 创建init layer
	initUuids, err := uuid.NewV4()
	if err != nil {
		return "", exception.NewGenericError(err, exception.UnionFsError)
	}
	initLayerId := initUuids.String()
	initLayerPath, err := service.prepareMountPath(containerId, initLayerId, initLayer)
	if err != nil {
		return "", exception.NewGenericError(err, exception.UnionFsError)
	}

	// 4. 将ro,rw 一起mount到init layer中
	// 这里dirs是第一个为rw,后面的均为ro
	dirs := fmt.Sprintf("dirs=%s:%s", rwLayerPath, roLayerPath)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", initLayerPath)
	logrus.Infof("executing %v", cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", exception.NewGenericError(err, exception.UnionFsMountError)
	}
	return initLayerPath, nil
}

func (service *imageService) prepareSpec(rootfsPath string, bundle string, imageRunArgs *ImageRunArgs, mounts []specs.Mount) (*specs.Spec, error) {
	spec := buildSpec(rootfsPath, imageRunArgs.Args, imageRunArgs.Env, imageRunArgs.Cwd, imageRunArgs.Hostname, imageRunArgs.Cpushare, imageRunArgs.Memory, imageRunArgs.Annotations, mounts)
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

func (service *imageService) generateLayerPath(imageId string) string {
	return filepath.Join(service.imageRoot, constant.ImageLayersDir, service.repositories[imageId])
}

/*
1. 在/var/run/capsule/images/mounts/$container_id/ 下创建对应layerType的文件,文件名为layerType的值,文件内容为layer_id
2. 在/var/run/capsule/images/layers/$layer_id 下创建对应的目录(对于read only不需要)
*/
func (service *imageService) prepareMountPath(containerId string, layerId string, t layerType) (string, error) {
	// 1.
	containerMountPath := filepath.Join(service.imageRoot, constant.ImageMountsDir, containerId)
	logrus.Infof("preparing container[%s] %s layer, layerId is %s", containerId, t, layerId)
	if err := os.MkdirAll(containerMountPath, 0644); err != nil {
		return "", err
	}
	file, err := os.Create(filepath.Join(containerMountPath, string(t)))
	if err != nil {
		return "", err
	}
	if _, err := file.WriteString(layerId); err != nil {
		return "", err
	}
	defer file.Close()

	// 2.
	layerPath := filepath.Join(service.imageRoot, constant.ImageLayersDir, layerId)
	if t != readOnlyLayer {
		// 对于读写和init layer,都需要创建
		if err := os.MkdirAll(layerPath, 0644); err != nil {
			return "", err
		}
	}
	return layerPath, nil
}

func (service *imageService) prepareHosts(containerId string, links []string) (specs.Mount, error) {
	hostsPath := filepath.Join(service.imageRoot, constant.ImageContainersDir, containerId, "hosts")
	file, err := os.OpenFile(hostsPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return specs.Mount{}, exception.NewGenericError(err, exception.HostsError)
	}
	defer file.Close()
	if _, err := file.WriteString("127.0.0.1 localhost\n"); err != nil {
		return specs.Mount{}, exception.NewGenericError(err, exception.HostsError)
	}
	for _, link := range links {
		splits := strings.SplitN(link, ":", 2)
		linkedContainerId := splits[0]
		alias := splits[1]
		container, err := service.factory.Load(linkedContainerId)
		if err != nil {
			return specs.Mount{}, exception.NewGenericError(err, exception.HostsError)
		}
		stateStorage, err := container.State()
		if err != nil {
			return specs.Mount{}, exception.NewGenericError(err, exception.HostsError)
		}
		ip := stateStorage.Endpoint.IpAddress.String()
		if _, err := file.WriteString(fmt.Sprintf("%s %s\n", ip, alias)); err != nil {
			return specs.Mount{}, exception.NewGenericError(err, exception.HostsError)
		}
	}
	return specs.Mount{
		Destination: "/etc/hosts",
		Type:        "bind",
		Source:      hostsPath,
		Options: []string{
			"rbind",
			"rprivate",
		},
	}, nil
}

func (service *imageService) prepareDns(containerId string) (specs.Mount, error) {
	resolvPath := filepath.Join(service.imageRoot, constant.ImageContainersDir, containerId, "resolv.conf")
	file, err := os.OpenFile(resolvPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return specs.Mount{}, exception.NewGenericError(err, exception.DnsError)
	}
	defer file.Close()
	bytes, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return specs.Mount{}, exception.NewGenericError(err, exception.DnsError)
	}
	if _, err := file.Write(bytes); err != nil {
		return specs.Mount{}, exception.NewGenericError(err, exception.DnsError)
	}
	return specs.Mount{
		Destination: "/etc/resolv.conf",
		Type:        "bind",
		Source:      resolvPath,
		Options: []string{
			"rbind",
			"rprivate",
		},
	}, nil
}

func (service *imageService) prepareVolumes(volumes []string) ([]specs.Mount, error) {
	var mounts []specs.Mount
	for _, volume := range volumes {
		var hostDir string
		var containerDir string
		if strings.Contains(volume, ":") {
			// host_dir:container_dir
			splits := strings.SplitN(volume, ":", 2)
			hostDir = splits[0]
			containerDir = splits[1]
			if _, err := os.Stat(hostDir); err != nil {
				return nil, exception.NewGenericError(err, exception.VolumeError)
			}
		} else {
			// random_dir:container_dir
			uuids, err := uuid.NewV4()
			if err != nil {
				return nil, exception.NewGenericError(err, exception.VolumeError)
			}
			volumeId := uuids.String()
			hostDir = filepath.Join(service.imageRoot, constant.ImageVolumesDir, volumeId)
			if err := os.MkdirAll(hostDir, 0644); err != nil {
				return nil, err
			}
			containerDir = volume
		}
		mounts = append(mounts, specs.Mount{
			Destination: containerDir,
			Type:        "bind",
			Source:      hostDir,
			Options: []string{
				"rbind",
				"rprivate",
			},
		})
	}
	return mounts, nil
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
	imageDir := service.generateLayerPath(id)
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
		return exception.NewGenericError(fmt.Errorf("image %s not exists", id), exception.ImageIdNotExistsError)
	}

	imageDir := service.generateLayerPath(id)
	if _, err := os.Stat(imageDir); err != nil {
		return exception.NewGenericError(err, exception.ImageIdNotExistsError)
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
		fileInfo, err := os.Stat(service.generateLayerPath(id))
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
	imageDir := service.generateLayerPath(id)
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
