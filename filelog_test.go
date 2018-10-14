package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var filelogFiles = []string{"./examples/hello.c", "./examples/helloprint.c", "./examples/helloprint.h"}
var filelogExtraCopyFiles = []string{"./examples/build.sf"}
var filelogExpectedSize = []int64{112, 77, 96}
var filelogExpectedSha = []string{"9ac20f410e45c1f6f14fd26a7a324e268d89b2139b0aad3d776cc4f709764d4e", "b4dfc8a845621e80c6d7d680a7d4d207b6e854d4596f05885ebad97434128d19", "6b05c842a049f5dc6413d97933e75a59cde4bd3a050ceeec55db75c700af634b"}
var filelogExpectedContent = []FileData{[]byte{}, []byte("#include <stdio.h>\n\nvoid printHello(void) {\n    printf(\"Hello, world!\\n\");\n}\n"), []byte("#ifndef HELLOPRINT_H_\n#define HELLOPRINT_H_\n\nvoid printHello(void);\n\n#endif /* HELLOPRINT_H_ */\n")}

var filelogExpectedBase64 = []string{"H4sIAAAAAAAA/wAAAP//AQAA//8AAAAAAAAAAA==", "H4sICKpGHVsAA2hlbGxvcHJpbnQuYwBTzsxLzilNSVWwKS5JyczXy7Dj4irLz0xRKCjKzCvxSM3JydcA8TUVqrkUgAAsnKahBJbRUSjPL8pJUYzJU9K05qrlAgAU26rXTQAAAA==", "H4sICKpGHVsAA2hlbGxvcHJpbnQuaABTzkzLS0lNU/Bw9fHxDwjy9AuJ94jnUgYKZealoolyleVnpigUFGXmlXik5uTka4D4mtZcXMqpeSmZaQr6WqgaFLT0uQAfsGHpYAAAAA=="}

type TempDir struct {
	tempDir     string
	originalCwd string
}

func NewTempDir(prefix string) (*TempDir, error) {
	originalCwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tmpDir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return nil, err
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s tmpdir: %s\n", prefix, tmpDir)
	allCopyFiles := append(filelogFiles, filelogExtraCopyFiles...)
	for _, v := range allCopyFiles {
		baseDir := path.Dir(v)
		err = os.MkdirAll(baseDir, 0700)
		if err != nil {
			return nil, err
		}

		destFile, err := os.OpenFile(v, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
		defer destFile.Close()
		srcFile, err := os.Open(path.Join(originalCwd, v))
		if err != nil {
			return nil, err
		}
		defer srcFile.Close()
		io.Copy(destFile, srcFile)
	}
	return &TempDir{tempDir: tmpDir, originalCwd: originalCwd}, nil
}

func (v *TempDir) Close() error {
	err := os.RemoveAll(v.tempDir)
	if err != nil {
		return err
	}
	return os.Chdir(v.originalCwd)
}

func checkFileLog(t *testing.T, logs []FileLog, hint string) {
	for i, v := range filelogFiles {
		if logs[i].Relpath != v {
			t.Fatalf("bad relative path: %s", logs[0].Relpath)
		}
		if logs[i].AbsPath[0] != '/' {
			t.Fatalf("bad absolute path: %s", logs[0].AbsPath)
		}
		if fmt.Sprintf("%x", logs[i].Sha256Sum) != filelogExpectedSha[i] {
			t.Fatalf("bad sha256hash for %s : actual: %x / expected: %s", v, logs[i].Sha256Sum, filelogExpectedSha[i])
		}
		if logs[i].Size != filelogExpectedSize[i] {
			t.Fatalf("bad file size for %s : %d", v, filelogExpectedSize[i])
		}
		//if !reflect.DeepEqual(logs[i].Content, filelogExpectedContent[i]) {
		//	t.Fatalf("%s bad content for %s\n%s\n%s\n%d\n%d\n", hint, v, logs[i].Content, filelogExpectedContent[i], len(logs[i].Content), len(filelogExpectedContent[i]))
		//}
		if logs[i].IsDir != false {
			t.Fatalf("bad is dir result for %s", v)
		}
	}
}

func TestFileLog(t *testing.T) {
	ClearCache()
	// create temporary directory, copy files and change current directory
	tmp, err := NewTempDir("filelog")
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	// validation start
	logs, err := CreateFileLog(filelogFiles, false, 100)
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}
	checkFileLog(t, logs, "first check")

	byteData, err := json.Marshal(logs)
	if err != nil {
		t.Fatalf("error: %s", err.Error())
	}

	//jsonResult := make([]fileLogForJSON, 0)
	//json.Unmarshal(byteData, &jsonResult)

	/*
		for i, v := range filelogFiles {
			if jsonResult[i].Relpath != v {
				t.Fatalf("bad relative path: %s", logs[0].relpath)
			}
			if jsonResult[i].Abspath[0] != '/' {
				t.Fatalf("bad absolute path: %s", jsonResult[0].Abspath)
			}
			if fmt.Sprintf("%s", jsonResult[i].Sha256sum) != filelogExpectedSha[i] {
				t.Fatalf("bad sha256hash for %s / actual:%s / expected: %s", v, jsonResult[i].Sha256sum, filelogExpectedSha[i])
			}
			if jsonResult[i].Size != filelogExpectedSize[i] {
				t.Fatalf("bad file size for %s : %d", v, filelogExpectedSize[i])
			}
			//if jsonResult[i].Content != filelogExpectedBase64[i] {
			//	t.Fatalf("bad content for %s\n%s\n%s", v, jsonResult[i].Content, filelogExpectedContent[i])
			//}
			if jsonResult[i].IsDir != false {
				t.Fatalf("bad is dir result for %s", v)
			}
		}
	*/

	filelogRead := make([]FileLog, 0)
	err = json.Unmarshal(byteData, &filelogRead)
	if err != nil {
		t.Fatalf("failed to unmarshal: %s", err.Error())
	}

	for i, v := range logs {
		if v.Modified.Unix() != filelogRead[i].Modified.Unix() {
			t.Fatalf("bad modified data / actual: %s / expected: %s", filelogRead[i].Modified, v.Modified)
		}

		if v.Modified.UnixNano() != filelogRead[i].Modified.UnixNano() {
			t.Fatalf("bad modified data / actual: %s / expected: %s", filelogRead[i].Modified, v.Modified)
		}
	}

	//fmt.Printf("\"%s\"\n\"%s\"\n\"%s\"\n", filelogRead[0].content, filelogRead[1].content, filelogRead[2].content)

	checkFileLog(t, filelogRead, "json decoded")

	//time.Sleep(1100 * time.Millisecond)
	//time.Sleep(100 * time.Millisecond)

	for _, v := range filelogRead {
		if result, err := v.IsChanged(); err != nil || result != false {
			t.Fatalf("bad is changed result: %s", err)
		}

		file, err := os.OpenFile(v.Relpath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}
		file.WriteString("hoge")
		file.Close()

		ClearCache()

		if result, err := v.IsChanged(); err != nil || result != true {
			t.Fatalf("bad is changed result: %s / %s", err, v.Relpath)
		}
	}
	defer tmp.Close()
}
