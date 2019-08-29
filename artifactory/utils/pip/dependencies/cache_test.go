package dependencies

import (
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"reflect"
	"testing"
)

func TestDependenciesCache(t *testing.T) {
	cacheMap := make(map[string]*buildinfo.Dependency)
	csA := buildinfo.Checksum{Sha1: "sha1A", Md5: "md5A"}
	depenA := buildinfo.Dependency{
		Id:       "depenA-1.0-A.zip",
		Checksum: &csA,
	}
	cacheMap["A"] = &depenA
	csC := buildinfo.Checksum{Sha1: "sha1C", Md5: "md5C"}
	depenC := buildinfo.Dependency{
		Id:       "depenC-3.4-C.gzip",
		Checksum: &csC,
	}
	cacheMap["C"] = &depenC
	err := UpdateDependenciesCache(cacheMap)
	if err != nil {
		t.Error("Failed creating dependencies cache: " + err.Error())
	}
	cache, err := GetProjectDependenciesCache()
	if cache == nil {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = "Cache file does not exist."
		}
		t.Error("Failed reading dependencies cache: " + errMsg)
	}

	if !reflect.DeepEqual(*cache.GetDependency("A"), depenA) {
		t.Error("Failed retrieving dependency A!!!")
	}
	if cache.GetDependency("B") != nil {
		t.Error("Retrieving non-existing dependency B should return nil!!!")
	}
	if !reflect.DeepEqual(*cache.GetDependency("C"), depenC) {
		t.Error("Failed retrieving dependency C!!!")
	}
	if cache.GetDependency("T") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum!!!")
	}

	delete(cacheMap, "A")
	csT := buildinfo.Checksum{Sha1: "sha1T", Md5: "md5T"}
	depenT := buildinfo.Dependency{
		Id:       "depenT-6.0.68-T.zip",
		Checksum: &csT,
	}
	cacheMap["T"] = &depenT
	err = UpdateDependenciesCache(cacheMap)
	if err != nil {
		t.Error("Failed creating dependencies cache: " + err.Error())
	}

	cache, err = GetProjectDependenciesCache()
	if cache == nil {
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = "Cache file does not exist."
		}
		t.Error("Failed reading dependencies cache: " + errMsg)
	}
	if cache.GetDependency("A") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum!!!")
	}
	if !reflect.DeepEqual(*cache.GetDependency("T"), depenT) {
		t.Error("Failed retrieving dependency T!!!")
	}
	if !reflect.DeepEqual(*cache.GetDependency("C"), depenC) {
		t.Error("Failed retrieving dependency C!!!")
	}
}
