package utils

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
)

const (
	// This env var is mandatory for the 'publish' command.
	// The env var is optional for the install command - if provided, the plugin will be downloaded from a custom
	// plugins server, instead of the official registry.
	// The env var should store a server ID configured by JFrog CLI.
	PluginsServerEnv = "JFROG_CLI_PLUGINS_SERVER"
	// Used to set a custom plugins repo for the 'publish' & 'install' commands.
	PluginsRepoEnv     = "JFROG_CLI_PLUGINS_REPO"
	DefaultPluginsRepo = "jfrog-cli-plugins"

	PluginsOfficialRegistryUrl = "https://releases.jfrog.io/artifactory/"

	LatestVersionName = "latest"
)

var ArchitecturesMap = map[string]Architecture{
	"linux-386":     {"linux", "386", ""},
	"linux-amd64":   {"linux", "amd64", ""},
	"linux-s390x":   {"linux", "s390x", ""},
	"linux-arm64":   {"linux", "arm64", ""},
	"linux-arm":     {"linux", "arm", ""},
	"mac-386":       {"darwin", "amd64", ""},
	"windows-amd64": {"windows", "amd64", ".exe"},
}

func GetLocalPluginExecutableName(pluginName string) string {
	if coreutils.IsWindows() {
		return pluginName + ".exe"
	}
	return pluginName
}

// Returns the full path of a plugin in Artifactory.
// Example path: "repo-name/plugin-name/version/architecture-name/executable-name"
func GetPluginPathInArtifactory(pluginName, pluginVersion, architecture string) string {
	return path.Join(GetPluginVersionDirInArtifactory(pluginName, pluginVersion), architecture, pluginName+ArchitecturesMap[architecture].FileExtension)
}

// Example path: "repo-name/plugin-name/v1.0.0/"
func GetPluginVersionDirInArtifactory(pluginName, pluginVersion string) string {
	return path.Join(GetPluginsRepo(), pluginName, pluginVersion)
}

// Returns a custom plugins repo if provided, default otherwise.
func GetPluginsRepo() string {
	repo := os.Getenv(PluginsRepoEnv)
	if repo != "" {
		return repo
	}
	return DefaultPluginsRepo
}

type Architecture struct {
	Goos          string
	Goarch        string
	FileExtension string
}

// Get the local architecture name corresponding to the architectures that exist in registry.
func GetLocalArchitecture() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows-amd64", nil
	case "darwin":
		return "mac-386", nil
	}
	// Assuming linux.
	switch runtime.GOARCH {
	case "amd64":
		return "linux-amd64", nil
	case "arm64":
		return "linux-arm64", nil
	case "arm":
		return "linux-arm", nil
	case "386":
		return "linux-386", nil
	case "s390x":
		return "linux-s390x", nil
	}
	return "", errorutils.CheckError(errors.New("no compatible plugin architecture was found for the architecture of this machine"))
}

func CreatePluginsHttpDetails(rtDetails *config.ServerDetails) httputils.HttpClientDetails {
	if rtDetails.AccessToken != "" && rtDetails.RefreshToken == "" {
		return httputils.HttpClientDetails{AccessToken: rtDetails.AccessToken}
	}
	return httputils.HttpClientDetails{
		User:     rtDetails.User,
		Password: rtDetails.Password,
		ApiKey:   rtDetails.ApiKey}
}

// Command used to build plugins.
type PluginBuildCmd struct {
	OutputFullPath string
	Env            map[string]string
}

func (buildCmd *PluginBuildCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, []string{"go", "build", "-o"}...)
	cmd = append(cmd, buildCmd.OutputFullPath)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (buildCmd *PluginBuildCmd) GetEnv() map[string]string {
	buildCmd.Env["CGO_ENABLED"] = "0"
	return buildCmd.Env
}

func (buildCmd *PluginBuildCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (buildCmd *PluginBuildCmd) GetErrWriter() io.WriteCloser {
	return nil
}