package dependencies

import (
	"encoding/json"
	"errors"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io"
	"io/ioutil"
	"strings"
)

func init() {
	var err error
	pipDependencyMapScriptPath, err = pip.GetPipDepTreeScriptPath()
	if err != nil {
		panic("Failed initializing dependency-map script.")
	}
}

var pipDependencyMapScriptPath string

// The extractor responsible to calculate the project dependencies.
type Extractor interface {
	// Get all the dependencies for the project.
	AllDependencies() map[string]*buildinfo.Dependency
	// Get all the root dependencies of the project.
	DirectDependencies() []string
	// Dependencies relations map.
	ChildrenMap() map[string][]string
	// Decide package name.
	PackageName() (string, error)

	Extract() error
}

// Execute pip-dependency-map script, return dependency map of all installed pip packages in current environment.
// pythonExecPath - Execution path python.
// envvars - Environment variables for script execution, used to pass the PATH variable when running inside a virtual environment.
func BuildPipDependencyMap(pythonExecPath string, envvars map[string]string) (map[string]pipDependencyPackage, error) {
	// Get python execution path.
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	// Execute the python pip-dependency-map script.
	pipDependencyMapCmd := &pip.PipCmd{
		Executable:  pythonExecPath,
		Command:     pipDependencyMapScriptPath,
		CommandArgs: []string{"--json"},
		EnvVars:     envvars,
		StrWriter:   pipeWriter,
		ErrWriter:   nil,
	}
	var pythonErr error
	go func() {
		pythonErr = gofrogcmd.RunCmd(pipDependencyMapCmd)
	}()
	data, err := ioutil.ReadAll(pipeReader)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	if pythonErr != nil {
		return nil, errorutils.CheckError(pythonErr)
	}

	// Parse the result.
	return parsePipDepTreeOutput(data)
}

// Parse pip-dependency-map raw output to dependencies map.
func parsePipDepTreeOutput(data []byte) (map[string]pipDependencyPackage, error) {
	// Parse into array.
	packages := make([]pipDependencyPackage, 0)
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Create packages map.
	packagesMap := make(map[string]pipDependencyPackage)
	for _, pkg := range packages {
		packagesMap[pkg.Package.Key] = pkg
	}

	return packagesMap, nil
}

// Extract all dependencies, based on 'dependencies'.
// Resolve allDependencies and childrenMap.
func extractDependencies(dependencies []string, environmentPackages map[string]pipDependencyPackage) (allDependencies map[string]*buildinfo.Dependency, childrenMap map[string][]string, err error) {
	allDependencies = make(map[string]*buildinfo.Dependency)
	childrenMap = make(map[string][]string)
	// Iterate over dependencies, resolve and discover more dependencies.
	index := -1
	var currentDep string
	for {
		index++

		// Check if should stop.
		if len(dependencies) < index+1 {
			break
		}

		currentDep = dependencies[index]
		// Check if current dependency already resolved.
		if _, ok := allDependencies[currentDep]; ok {
			// Already resolved.
			continue
		}

		// Resolve dependency.
		depTreePkg, ok := environmentPackages[currentDep]
		if !ok {
			err = errorutils.CheckError(errors.New(fmt.Sprintf("Failed getting information for dependency: %s", currentDep)))
			return
		}

		// Extract pip-dependency from pip-package.
		var pipDep *pipDependency
		pipDep, err = depTreePkg.extractPipDependency()
		if err != nil {
			return
		}

		// Update extractor's map.
		if pipDep != nil {
			allDependencies[pipDep.id] = pipDep.dependency
			childrenMap[pipDep.id] = pipDep.dependencies
		}

		// Add pipDep dependency for resolution.
		dependencies = append(dependencies, pipDep.dependencies...)
	}
	return
}

type pipDependency struct {
	id           string
	version      string
	dependency   *buildinfo.Dependency
	dependencies []string
}

func (pipDepTreePkg *pipDependencyPackage) extractPipDependency() (*pipDependency, error) {
	// Create pip-dependency.
	pipDependency := &pipDependency{id: pipDepTreePkg.Package.Key, version: pipDepTreePkg.Package.InstalledVersion, dependencies: pipDepTreePkg.getDependencies()}

	// Build build-info dependency.
	pipDependency.dependency = &buildinfo.Dependency{Id: pipDepTreePkg.Package.PackageName + ":" + pipDepTreePkg.Package.InstalledVersion}

	return pipDependency, nil
}

func (pipDepTreePkg *pipDependencyPackage) getDependencies() []string {
	var dependencies []string
	for _, dep := range pipDepTreePkg.Dependencies {
		dependencies = append(dependencies, strings.ToLower(dep.Key))
	}
	return dependencies
}

// Structs for parsing the pip-dependency-map result.

type pipDependencyPackage struct {
	Package      packageType  `json:"package,omitempty"`
	Dependencies []dependency `json:"dependencies,omitempty"`
}

type packageType struct {
	Key              string `json:"key,omitempty"`
	PackageName      string `json:"package_name,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
}

type dependency struct {
	Key              string `json:"key,omitempty"`
	PackageName      string `json:"package_name,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
}
