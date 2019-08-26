package dependencies

import (
	"bytes"
	"encoding/json"
	"fmt"
	logUtils "github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePipDepTree(t *testing.T) {
	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// Create file path.
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	pipdeptreeTestFilePath := filepath.Join(pwd, "testsdata/pipdeptree_output")

	// Read file.
	content, err := fileutils.ReadFile(pipdeptreeTestFilePath)
	if err != nil {
		t.Error("Failed reading file!!!")
	}

	// Parse content.
	depTree, err := parsePipDepTreeOutput(content)
	if err != nil {
		t.Error("Failed parsing dep tree!!!")
	}

	// Print results.
	//log.Info(fmt.Sprintf("Result:\n%+v\n", depTree))
	s, _ := json.MarshalIndent(depTree, "", "\t")
	log.Info(fmt.Sprintf("Result:\n%s", s))
}

func TestRunPipDepTreeAndParse(t *testing.T) {
	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	pythonPath := "/Users/barb/trash/venv-test2/bin/python"
	pathVar := os.Getenv("PATH")
	os.Setenv("PATH", "/Users/barb/trash/venv-test2/bin")
	defer os.Setenv("PATH", pathVar)

	// Run.
	depTree, err := BuildPipDependencyMap(pythonPath, nil)
	if err != nil {
		t.Error("FAILED!!!!")
	}

	// Print results.
	//log.Info(fmt.Sprintf("Result:\n%+v\n", depTree))
	s, _ := json.MarshalIndent(depTree, "", "\t")
	log.Info(fmt.Sprintf("Result:\n%s", s))
}

func TestExtractDependencies(t *testing.T) {
	// Create log.
	newLog := log.NewLogger(logUtils.GetCliLogLevel(), nil)
	buffer := &bytes.Buffer{}
	newLog.SetOutputWriter(buffer)
	log.SetLogger(newLog)

	// GET PIPDEPTREE OUTPUT

	// Create file path.
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	pipdeptreeTestFilePath := filepath.Join(pwd, "testsdata/pipdeptree_output")

	// Read file.
	content, err := fileutils.ReadFile(pipdeptreeTestFilePath)
	if err != nil {
		t.Error("Failed reading file!!!")
	}

	// Parse content.
	depTree, err := parsePipDepTreeOutput(content)
	if err != nil {
		t.Error("Failed parsing dep tree!!!")
	}

	// GET ROOT DEPS

	rootDeps := []string{"pyinstaller", "pipdeptree", "macholib"}

	// RUN

	allDeps, childMap, err := extractDependencies(rootDeps, depTree)

	// Print results.
	log.Info(fmt.Sprintf("ALL DEPS:\n%v", allDeps))
	log.Info(fmt.Sprintf("CHILDREN MAP:\n%v", childMap))
}

func TestDependenciesCache(t *testing.T) {
	cache := make(DependenciesCache)
	csA := buildinfo.Checksum{Sha1: "sha1A", Md5: "md5A"}
	depenA := buildinfo.Dependency{
		Id:       "depenA-1.0-A.zip",
		Checksum: &csA,
	}
	cache["A"] = &depenA
	depenB := buildinfo.Dependency{
		Id: "depenB-1.17-B.exe",
	}
	cache["B"] = &depenB
	csC := buildinfo.Checksum{Sha1: "sha1C", Md5: "md5C"}
	depenC := buildinfo.Dependency{
		Id:       "depenC-3.4-C.gzip",
		Checksum: &csC,
	}
	cache["C"] = &depenC
	csD := buildinfo.Checksum{Md5: "md5D"}
	depenD := buildinfo.Dependency{
		Id:       "depenD-2.0-D.mod",
		Checksum: &csD,
	}
	cache["D"] = &depenD
	csE := buildinfo.Checksum{Md5: "sha1E"}
	depenE := buildinfo.Dependency{
		Id:       "depenE-2.7.61-D.tar.gz",
		Checksum: &csE,
	}
	cache["E"] = &depenE
	err := UpdateDependenciesCache(cache)
	if err != nil {
		t.Error("Failed creating dependencies cache!!!")
	}
	newCache, err := GetProjectDependenciesCache()
	if newCache == nil || err != nil {
		t.Error("Failed reading dependencies cache!!!")
	}

	if *newCache.GetDependencyChecksum("A") != csA {
		t.Error("Failed retrieving checksum of dependency A!!!")
	}
	if newCache.GetDependencyChecksum("B") != nil {
		t.Error("Retrieving missing checksum of dependency B should return nil!!!")
	}
	if *newCache.GetDependencyChecksum("C") != csC {
		t.Error("Failed retrieving checksum of dependency C!!!")
	}
	if newCache.GetDependencyChecksum("D") != nil {
		t.Error("Retrieving a non-vaild checksum of dependency D should return nil!!!")
	}
	if newCache.GetDependencyChecksum("E") != nil {
		t.Error("Retrieving non-vaild checksum of dependency E should return nil!!!")
	}
	if newCache.GetDependencyChecksum("T") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum!!!")
	}

	delete(*newCache, "A")
	csT := buildinfo.Checksum{Sha1: "sha1T", Md5: "md5T"}
	depenT := buildinfo.Dependency{
		Id:       "depenT-6.0.68-T.zip",
		Checksum: &csT,
	}
	(*newCache)["T"] = &depenT
	err = UpdateDependenciesCache(cache)
	if err != nil {
		t.Error("Failed creating dependencies cache!!!")
	}

	lastCache, err := GetProjectDependenciesCache()
	if lastCache == nil || err != nil {
		t.Error("Failed reading dependencies cache!!!")
	}
	if newCache.GetDependencyChecksum("A") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum!!!")
	}
	if *newCache.GetDependencyChecksum("T") != csT {
		t.Error("Failed retrieving checksum of dependency T!!!")
	}

}
