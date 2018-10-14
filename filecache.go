package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

var fileStatCache = make(map[string]os.FileInfo)

func ClearCache() {
	if Sha256CacheConnection != nil {
		Sha256CacheConnection.Connection.Exec("DROP TABLE IF EXISTS Sha256cache")
		Sha256CacheConnection.Connection.Close()
		Sha256CacheConnection = nil
	}
	fileStatCache = make(map[string]os.FileInfo)
}

func Stat(name string) (os.FileInfo, error) {
	if stat, ok := fileStatCache[name]; ok {
		return stat, nil
	}
	stat, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	fileStatCache[name] = stat
	return stat, nil
}

type Sha256Cache struct {
	CacheFilePath string
	Connection    *sql.DB
}

func NewSha256Cache() *Sha256Cache {
	err := os.MkdirAll(WorkflowLogDir, 0755)
	if err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "(Ignored) Cannot create workflow log directory: %s\n", err)
		return nil
		// panic(err.Error())
	}

	filename := path.Join(WorkflowLogDir, "files.sqlite3")
	conn, err := sql.Open("sqlite3", filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "(Ignored) Cannot open sqlite3 file: %s\n", err)
		return nil
		// panic(err.Error())
	}

	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS Sha256Cache(path TEXT, modified TEXT, size INTEGER, sha256 TEXT, PRIMARY KEY(path, modified, size))")
	if err != nil {
		fmt.Fprintf(os.Stderr, "(Ignored) Cannot create cache table: %s\n", err)
		return nil
		// panic(err.Error())
	}

	return &Sha256Cache{
		CacheFilePath: filename,
		Connection:    conn,
	}
}

var Sha256CacheConnection *Sha256Cache

func CalcSha256ForFile(filepath string, maximumContentLogSize int64) (HashSum, error) {
	if Sha256CacheConnection == nil {
		Sha256CacheConnection = NewSha256Cache()
	}

	stat, err := Stat(filepath)
	if err != nil {
		return nil, err
	}

	if Sha256CacheConnection != nil {
		result := Sha256CacheConnection.Connection.QueryRow("SELECT sha256 FROM Sha256Cache WHERE path = ? AND modified = ? AND size = ?", filepath, stat.ModTime().String(), stat.Size())
		var hashString string
		err = result.Scan(&hashString)
		if err == nil {
			var hashSum HashSum
			_, err = fmt.Sscanf(hashString, "%x", &hashSum)
			return hashSum, nil
		}
		if err != sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "(Ignored) Cannot scan sqlite database: %s\n", err)
			//return nil, err
		}
	}

	backupContent := stat.Size() <= maximumContentLogSize
	var content = bytes.NewBuffer(nil)
	var reader io.Reader

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fmt.Printf("calculating SHA256 for %s\n", filepath)

	if backupContent {
		reader = io.TeeReader(file, content)
	} else {
		reader = file
	}

	hash := sha256.New()
	_, err = io.Copy(hash, reader)
	if err != nil {
		return nil, err
	}

	hashResult := hash.Sum(nil)
	hashString := fmt.Sprintf("%x", hashResult)

	// backup copy
	if backupContent {
		backupPath := path.Join(WorkflowLogDir, "__backup", hashString[:1], hashString[:2], hashString+".gz")
		backupDir := path.Dir(backupPath)
		err = os.MkdirAll(backupDir, 0755)
		if err != nil && !os.IsExist(err) {
			return nil, err
		}
		backupFile, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer backupFile.Close()
		gzipEncoder := gzip.NewWriter(backupFile)
		defer gzipEncoder.Close()
		gzipEncoder.Write(content.Bytes())
	}

	// register to file log
	if Sha256CacheConnection != nil {
		_, err = Sha256CacheConnection.Connection.Exec("INSERT INTO Sha256Cache(path, modified, size, sha256) VALUES(?, ?, ?, ?)", filepath, stat.ModTime().String(), stat.Size(), hashString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "(Ignored) Cannot log SHA256 into database: %s\n", err.Error())
			return hashResult, nil
		}
	}

	return hashResult, nil
}
