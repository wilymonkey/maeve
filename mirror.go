package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type FileInfo struct {
	relPath string
	Size    int64
	ModTime time.Time
}

type FileList struct {
	Files []FileInfo
}

func mirrorDir(localDir, remoteHost, remoteDir string) {
	remoteFiles := getRemoteFileList(remoteHost, remoteDir)
	localFiles := getLocalFileList(localDir)
	changedFiles := compareFiles(localFiles, remoteFiles)
	sendChangedFiles(changedFiles, localDir, remoteHost, remoteDir)
	sendDeletedFiles(localFiles, remoteFiles, remoteHost, remoteDir)
}

// getLocalFileList retrieves files in directory
func getLocalFileList(dir string, pos uint) FileList {
	var files []FileInfo
	filepath.WalkDir(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(dir, path)
			files = append(files, FileInfo{
				relPath: relPath,
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}
		return nil
	})
	return FileList{Files: files}
}

// getRemoteFileList fetches file list from server
func getRemoteFileList(host, dir string) FileList {
	cmd := exec.Command("ssh", host, fmt.Sprintf("%s list %s", os.Args[0], dir))
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	var fileList FileList
	dec := gob.NewDecoder(&stdout)
	if err := dec.Decode(&fileList); err != nil {
		log.Fatal(err)
	}
	return fileList
}

// sendFileList sends directory file list
func sendFileList(dir string) {
	fileList := getLocalFileList(dir)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(fileList); err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(buf.Bytes())
}

// compareFiles identifies changed files
func compareFiles(local, remote FileList) []FileInfo {
	var changed []FileInfo
	localMap := make(map[string]FileInfo)
	for _, f := range local.Files {
		localMap[f.relPath] = f
	}
	for _, rf := range remote.Files {
		lf, exists := localMap[rf.relPath]
		if !exists || lf.Size != rf.Size || !lf.ModTime.Equal(rf.ModTime) {
			changed = append(changed, lf)
		}
	}
	return changed
}

// sendChangedFiles sends modified files to server
func sendChangedFiles(changed []FileInfo, localDir, host, remoteDir string) {
	for _, f := range changed {
		localPath := filepath.Join(localDir, f.relPath)
		cmd := exec.Command("ssh", host, fmt.Sprintf("%s receive %s", os.Args[0], remoteDir))
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			defer stdin.Close()
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			if err := enc.Encode(f); err != nil {
				log.Fatal(err)
			}
			file, err := os.Open(localPath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			io.Copy(stdin, &buf)
			io.Copy(stdin, file)
		}()
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

// receiveFiles handles incoming file transfers
func receiveFiles(dir string) {
	var fInfo FileInfo
	dec := gob.NewDecoder(os.Stdin)
	if err := dec.Decode(&fInfo); err != nil {
		log.Fatal(err)
	}
	filePath := filepath.Join(dir, fInfo.relPath)
	os.MkdirAll(filepath.Dir(filePath), 0755)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if _, err := io.Copy(file, os.Stdin); err != nil {
		log.Fatal(err)
	}
	file.Close()
}

// sendDeletedFiles sends list of deleted files
func sendDeletedFiles(local, remote FileList, host, remoteDir string) {
	var deleted []FileInfo
	remoteMap := make(map[string]bool)
	for _, f := range remote.Files {
		remoteMap[f.relPath] = true
	}
	for _, lf := range local.Files {
		if !remoteMap[lf.relPath] {
			continue
		}
	}
	for rfName := range remoteMap {
		deleted = append(deleted, FileInfo{relPath: rfName})
	}
	if len(deleted) == 0 {
		return
	}
	cmd := exec.Command("ssh", host, fmt.Sprintf("%s delete %s", os.Args[0], remoteDir))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(FileList{Files: deleted}); err != nil {
			log.Fatal(err)
		}
		io.Copy(stdin, &buf)
	}()
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// deleteFiles removes specified files
func deleteFiles(dir string) {
	var deleted FileList
	dec := gob.NewDecoder(os.Stdin)
	if err := dec.Decode(&deleted); err != nil {
		log.Fatal(err)
	}
	for _, f := range deleted.Files {
		filePath := filepath.Join(dir, f.relPath)
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete %s: %v", filePath, err)
		}
	}
}
