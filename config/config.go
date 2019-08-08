package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gobuffalo/packr/v2"

	"github.com/TouchBistro/tb/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var services ServiceMap
var playlists map[string]Playlist
var tbRoot string

const (
	servicesPath             = "services.yml"
	playlistPath             = "playlists.yml"
	dockerComposePath        = "docker-compose.yml"
	localstackEntrypointPath = "localstack-entrypoint.sh"
	ecrURIRoot               = "651264383976.dkr.ecr.us-east-1.amazonaws.com"
)

func setupEnv() error {
	// Set $TB_ROOT so it works in the docker-compose file
	tbRoot = fmt.Sprintf("%s/.tb", os.Getenv("HOME"))
	os.Setenv("TB_ROOT", tbRoot)

	// Create $TB_ROOT directory if it doesn't exist
	if !util.FileOrDirExists(tbRoot) {
		err := os.Mkdir(tbRoot, 0755)
		if err != nil {
			return errors.Wrapf(err, "failed to create $TB_ROOT directory at %s", tbRoot)
		}
	}
	return nil
}

func dumpFile(name string, box *packr.Box) error {
	path := fmt.Sprintf("%s/%s", tbRoot, name)
	buf, err := box.Find(name)
	if err != nil {
		return errors.Wrapf(err, "failed to find packr box %s", name)
	}

	var reason string
	// If file exists compare the checksum to the packr version
	if util.FileOrDirExists(path) {
		log.Debugf("%s exists", path)
		log.Debugf("comparing checksums for %s", name)

		fileBuf, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read contents of %s", path)
		}

		memChecksum, err := util.MD5Checksum(buf)
		if err != nil {
			return errors.Wrapf(err, "failed to get checksum of %s in packr box", name)
		}

		fileChecksum, err := util.MD5Checksum(fileBuf)
		if err != nil {
			return errors.Wrapf(err, "failed to get checksum of %s", path)
		}

		// checksums are the same, leave as is
		if bytes.Equal(memChecksum, fileChecksum) {
			log.Debugf("checksums match, leaving %s as is", name)
			return nil
		}

		reason = "is outdated, recreating file..."
	} else {
		reason = "does not exist, creating file..."
	}

	log.Debugf("%s %s", path, reason)

	err = ioutil.WriteFile(path, buf, 0644)
	return errors.Wrapf(err, "failed to write contents of %s to %s", name, path)
}

func TBRootPath() string {
	return tbRoot
}

func Init() error {
	err := setupEnv()
	if err != nil {
		return errors.Wrap(err, "failed to setup $TB_ROOT env")
	}

	box := packr.New("static", "../static")

	sBuf, err := box.Find(servicesPath)
	if err != nil {
		return errors.Wrapf(err, "failed to find packr box %s", servicesPath)
	}

	err = util.DecodeYaml(bytes.NewReader(sBuf), &services)
	if err != nil {
		return errors.Wrapf(err, "failed decode yaml for %s", servicesPath)
	}

	pBuf, err := box.Find(playlistPath)
	if err != nil {
		return errors.Wrapf(err, "failed to find packr box %s", playlistPath)
	}
	err = util.DecodeYaml(bytes.NewReader(pBuf), &playlists)
	if err != nil {
		return errors.Wrapf(err, "failed decode yaml for %s", playlistPath)
	}

	err = dumpFile(dockerComposePath, box)
	if err != nil {
		return errors.Wrapf(err, "failed to dump file to %s", dockerComposePath)
	}

	err = dumpFile(localstackEntrypointPath, box)
	if err != nil {
		return errors.Wrapf(err, "failed to dump file to %s", localstackEntrypointPath)
	}

	err = applyOverrides(services, tbrc.Overrides)
	if err != nil {
		return errors.Wrap(err, "failed to apply overrides from tbrc")
	}

	// Setup service names image URI env vars for docker-compose
	for name, s := range services {
		serviceName := name
		serviceNameVar := util.StringToUpperAndSnake(name) + "_NAME"
		if s.ECR {
			serviceName += "-ecr"
		}
		os.Setenv(serviceNameVar, serviceName)

		// Set imageURIs for ECR and Dockerhub hosted images.
		// non-ecr images. eg: postgres, redis, localstack
		if !s.ECR && s.DockerhubImage != "" {
			uriVar := util.StringToUpperAndSnake(name) + "_IMAGE_URI"
			os.Setenv(uriVar, s.DockerhubImage)
		}

		// ecr images. eg: 651264383976.dkr.ecr.us-east-1.amazonaws.com/venue-provisioning-service:master-e09270363e044e37c430c7997359d55697e6b165
		if s.ECR && s.ECRTag != "" {
			uri := ResolveEcrURI(name, s.ECRTag)
			uriVar := util.StringToUpperAndSnake(name) + "_IMAGE_URI"
			os.Setenv(uriVar, uri)
		}
	}

	return nil
}

func Services() ServiceMap {
	return services
}

func Playlists() map[string]Playlist {
	return playlists
}

func BaseImages() []string {
	return []string{
		"touchbistro/alpine-node:10-build",
		"touchbistro/alpine-node:10-runtime",
		"touchbistro/alpine-node:12-build",
		"touchbistro/alpine-node:12-runtime",
		"touchbistro/ubuntu16-ruby:2.5.5-build",
	}
}

func GetPlaylist(name string, deps map[string]bool) []string {
	// Initialize no playlists if we couldnt load yaml in Init()
	if playlists == nil {
		playlists = make(map[string]Playlist)
	}
	customList := tbrc.Playlists

	// Check custom playlists first
	if playlist, ok := customList[name]; ok {
		// Resolve parent playlist defined in extends
		if playlist.Extends != "" {
			deps[name] = true
			if deps[playlist.Extends] {
				log.Fatalf("Circular dependency of services, %s and %s", playlist.Extends, name)
			}
			parentPlaylist := GetPlaylist(playlist.Extends, deps)
			return append(parentPlaylist, playlist.Services...)
		}

		return playlist.Services
	} else if playlist, ok := playlists[name]; ok {
		if playlist.Extends != "" {
			deps[name] = true
			if deps[playlist.Extends] {
				log.Fatalf("Circular dependency of services, %s and %s", playlist.Extends, name)
			}
			parentPlaylist := GetPlaylist(playlist.Extends, deps)
			return append(parentPlaylist, playlist.Services...)
		}

		return playlist.Services
	}

	return []string{}
}

func RmFiles() error {
	files := [...]string{dockerComposePath, localstackEntrypointPath}

	for _, file := range files {
		log.Debugf("Removing %s...\n", file)
		path := fmt.Sprintf("%s/%s", tbRoot, file)
		err := os.Remove(path)
		if err != nil {
			return errors.Wrapf(err, "could not remove file at %s", path)
		}
	}

	return nil
}
