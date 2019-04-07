package constant

const (
	ImageDir                  = "images"
	ImageContainersDir        = "containers"
	ImageLayersDir            = "layers"
	ImageMountsDir            = "mounts"
	ImageVolumesDir           = "volumes"
	ImageRepositoriesFilename = "repositories.json"
)

// 三种layer:原生镜像(read-only layer);容器对应的write layer;容器对应的mnt layer
// runtimeRoot/images
// - /layers/$layer_id/解压后的rootfs

// - /mounts/$container_id
//	- read_only: 存放镜像的layer_id
//  - read_write: 存放读写层的layer_id
//  - init: 存放挂载后的layer_id

// runtimeRoot/images/repositories.json image_name->layer_id
