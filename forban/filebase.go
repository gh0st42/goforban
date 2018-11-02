package forban

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileEntry describes a file in shared web root
type FileEntry struct {
	Name     string
	FullPath string
	Size     int64
}

// FileList List of files currently in store
var FileList = []FileEntry{}

// ScanDirectory Scan a given path for files and report them all including their sizes
func ScanDirectory(scanPath string) []FileEntry {
	fileList := []FileEntry{}
	err := filepath.Walk(scanPath, func(path string, f os.FileInfo, err error) error {
		//fmt.Println(f)
		if f.IsDir() != true {
			newEntry := FileEntry{f.Name(), path, f.Size()}
			fileList = append(fileList, newEntry)
		}
		return nil
	})
	CheckError(err)
	return fileList
}

// UpdateFileIndex Updates the file index, calculates HMAC and writes both files to BasePath
func UpdateFileIndex() {
	FileList := ScanDirectory(FileBasePath)

	var buffer bytes.Buffer
	MyFiles = nil
	for _, file := range FileList {
		//fmt.Println(file.FullPath, file.Size)
		buffer.WriteString(strings.TrimPrefix(file.FullPath, FileBasePath+"/"))
		MyFiles = append(MyFiles, strings.TrimPrefix(file.FullPath, FileBasePath+"/"))
		buffer.WriteString(",")
		buffer.WriteString(strconv.FormatInt(file.Size, 10))
		buffer.WriteString("\n")
	}
	//fmt.Println(buffer.String())
	err := ioutil.WriteFile(FileBasePath+"/forban/index", buffer.Bytes(), 0644)
	CheckError(err)

	indexHmac := GetIndexHmac()
	err = ioutil.WriteFile(FileBasePath+"/forban/index.hmac", []byte(indexHmac), 0644)
	CheckError(err)

}
