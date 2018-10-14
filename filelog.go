package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

type FileData []byte
type HashSum []byte

type FileLog struct {
	Relpath   string
	AbsPath   string
	Size      int64
	Modified  time.Time
	IsDir     bool
	Sha256Sum HashSum
}

func (v FileLog) String() string {
	return fmt.Sprintf("FileLog{%s, %d}", v.Relpath, v.Size)
}

// MaximumContentLogSize is maximum size of file that will be saved in log file
const MaximumContentLogSize = 20 * 1024 * 1024 // 20MB

//var FileLogCache = make(map[string]FileLog)

func CreateFileLog(files []string, skipSha bool, maximumContentLogSize int64) ([]FileLog, error) {
	fileLogs := make([]FileLog, len(files))

	for i, v := range files {
		//if cache, ok := FileLogCache[v]; ok {
		//	fileLogs[i] = cache
		//	continue
		//}

		absPath := Abs(v)
		fileInfo, err := Stat(v)
		if err != nil {
			return nil, err
		}

		hashResult, err := CalcSha256ForFile(v, maximumContentLogSize)
		if err != nil {
			return nil, err
		}

		//backupContent := fileInfo.Size() <= maximumContentLogSize
		//var content *bytes.Buffer
		//if backupContent {
		//	content = bytes.NewBuffer(nil)
		//}
		//hash := sha256.New()
		//file, err := os.Open(v)
		//if err != nil {
		//	return nil, err
		//}
		//var reader io.Reader = file
		//if backupContent {
		//	reader = io.TeeReader(file, content)
		//}
		//
		//defer file.Close()
		//io.Copy(hash, reader)
		//
		//contentBytes := []byte{}
		//if backupContent {
		//	contentBytes = content.Bytes()
		//}

		fileLogs[i] = FileLog{
			Relpath:   v,
			AbsPath:   absPath,
			Size:      fileInfo.Size(),
			Modified:  fileInfo.ModTime(),
			Sha256Sum: hashResult,
		}

		//FileLogCache[v] = fileLogs[i]
	}

	return fileLogs, nil
}

var isChangedCache = make(map[string]bool)

// IsChanged function check whether the file is changed or not.
// Return true if the file is changed.
// This function does not check SHA256, but only existance and modification date
func (v *FileLog) IsChanged() (bool, error) {
	//if changed, ok := isChangedCache[v.Relpath]; ok {
	//		return changed, nil
	//	}

	stat, err := Stat(v.Relpath)
	changed := false
	if err != nil && os.IsNotExist(err) {
		changed = true
		return true, nil
	} else if err != nil {
		return true, err
	}
	if stat.ModTime().Unix() != v.Modified.Unix() || stat.ModTime().UnixNano() != v.Modified.UnixNano() || stat.Size() != v.Size {
		changed = true
	}

	isChangedCache[v.Relpath] = changed
	return changed, nil
}

func (v HashSum) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", v))
}

func (v FileData) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	buf := bytes.NewBuffer(nil)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	gzip := zlib.NewWriter(encoder)
	gzip.Write(v)
	gzip.Flush()
	gzip.Close()
	encoder.Close()

	return json.Marshal(string(buf.Bytes()))
}

func (v *HashSum) UnmarshalJSON(data []byte) error {
	var str string
	e := json.Unmarshal(data, &str)
	if e != nil {
		return e
	}
	_, e = fmt.Sscanf(str, "%x", v)
	return e
}

func (v *FileData) UnmarshalJSON(data []byte) error {
	if reflect.DeepEqual(data, []byte("null")) {
		*v = nil
		return nil
	}

	var str string
	e := json.Unmarshal(data, &str)
	if e != nil {
		return e
	}
	buffer := bytes.NewBuffer([]byte(str))
	decoder := base64.NewDecoder(base64.StdEncoding, buffer)
	gzipReader, err := zlib.NewReader(decoder)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	*v, err = ioutil.ReadAll(gzipReader)
	return err
}
