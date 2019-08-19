package utils

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/spf13/viper"
	"path/filepath"
	"reflect"
)

const (
	PROJECT_CONFIG_RESOLVER_PREFIX = "resolver"
	PROJECT_CONFIG_DEPLOYER_PREFIX = "deployer"
	PROJECT_CONFIG_REPO            = "repo"
	PROJECT_CONFIG_SERVER_ID       = "serverId"
)

type ProjectType int

const (
	GO ProjectType = iota
	PIP
)

var ProjectTypes = []string{
	"go",
	"pip",
}

func (projectType ProjectType) String() string {
	return ProjectTypes[projectType]
}

type Repository struct {
	Repo     string `yaml:"repo,omitempty"`
	ServerId string `yaml:"serverId,omitempty"`
}

type RepositoryConfig struct {
	targetRepo string
	rtDetails  *config.ArtifactoryDetails
}

// If configuration file exists in the working dir return its path, otherwise return the global configuration file path
func GetProjectConfFilePath(projectType ProjectType) (confFilePath string, exists bool, err error) {
	confFileName := filepath.Join("projects", projectType.String()+".yaml")
	projectDir, exists, err := fileutils.FindUpstream(".jfrog", fileutils.Dir)
	if err != nil {
		return "", false, err
	}
	if exists {
		confFilePath = filepath.Join(projectDir, ".jfrog", confFileName)
		exists, err = fileutils.IsFileExists(confFilePath, false)
		if err != nil {
			return "", false, err
		}

		if exists {
			return
		}
	}
	// If missing in the root project, check in the home dir
	jfrogHomeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		return "", exists, err
	}
	confFilePath = filepath.Join(jfrogHomeDir, confFileName)
	exists, err = fileutils.IsFileExists(confFilePath, false)
	return
}

func GetRepoConfigByPrefix(configFilePath, prefix string, vConfig *viper.Viper) (*RepositoryConfig, error) {
	if !vConfig.IsSet(prefix) {
		return nil, errorutils.CheckError(fmt.Errorf("%s information is missing within %s", prefix, configFilePath))
	}
	log.Debug(fmt.Sprintf("Found %s in the config file %s", prefix, configFilePath))
	repo := vConfig.GetString(prefix + "." + PROJECT_CONFIG_REPO)
	if repo == "" {
		return nil, fmt.Errorf("Missing repository for %s within %s", prefix, configFilePath)
	}
	serverId := vConfig.GetString(prefix + "." + PROJECT_CONFIG_SERVER_ID)
	if serverId == "" {
		return nil, fmt.Errorf("Missing server ID for %s within %s", prefix, configFilePath)
	}
	rtDetails, err := config.GetArtifactoryConf(serverId)
	if err != nil {
		return nil, err
	}
	return &RepositoryConfig{targetRepo: repo, rtDetails: rtDetails}, nil
}

func (repo *RepositoryConfig) IsRtDetailsEmpty() bool {
	if repo.rtDetails != nil && reflect.DeepEqual(config.ArtifactoryDetails{}, repo.rtDetails) {
		return false
	}
	return true
}

func (repo *RepositoryConfig) SetTargetRepo(targetRepo string) *RepositoryConfig {
	repo.targetRepo = targetRepo
	return repo
}

func (repo *RepositoryConfig) TargetRepo() string {
	return repo.targetRepo
}

func (repo *RepositoryConfig) SetRtDetails(rtDetails *config.ArtifactoryDetails) *RepositoryConfig {
	repo.rtDetails = rtDetails
	return repo
}

func (repo *RepositoryConfig) RtDetails() (*config.ArtifactoryDetails, error) {
	return repo.rtDetails, nil
}
