package k8s

import "yunion.io/x/onecloud/pkg/mcclient/modules"

var ImageRepos *ImageRepoManager

type ImageRepoManager struct {
	*NamespaceResourceManager
}

func init() {
	ImageRepos = &ImageRepoManager{
		NamespaceResourceManager: NewNamespaceResourceManager(
			"image_repo",
			"image_repos",
			NewColumns(),
			NewColumns())}
	modules.Register(ImageRepos)
}
