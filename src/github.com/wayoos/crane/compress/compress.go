package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/wayoos/crane/config"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func handleError(_e error) {
	if _e != nil {
		log.Fatal(_e)
	}
}

func TarGzWrite(rootPath string, _path string, tw *tar.Writer, fi os.FileInfo) {
	fr, err := os.Open(_path)
	handleError(err)
	defer fr.Close()

	h := new(tar.Header)
	tarPath := strings.TrimPrefix(_path, rootPath+"/")

	h.Name = tarPath
	//fi.Name()
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()

	err = tw.WriteHeader(h)
	handleError(err)

	_, err = io.Copy(tw, fr)
	handleError(err)
}

func IterDirectory(rootPath string, dirPath string, tw *tar.Writer) {
	dir, err := os.Open(dirPath)
	handleError(err)
	defer dir.Close()
	fis, err := dir.Readdir(0)
	handleError(err)
	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()
		if fi.IsDir() {
			if fi.Name() != ".git" {
				IterDirectory(rootPath, curPath, tw)
			}
		} else {
			//fmt.Printf( "adding... %s\n", curPath )
			TarGzWrite(rootPath, curPath, tw, fi)
		}
	}
}

func TarGz(path string, outFilePath io.Writer) {
	// gzip write
	gw := gzip.NewWriter(outFilePath)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	IterDirectory(path, path, tw)
}

func UnTarGz(tarFilePath string, inPath string) {
	// file to read
	fr, err := os.Open(tarFilePath)
	handleError(err)
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	handleError(err)
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		//fmt.Printf("Contents of %s\n", hdr.Name)

		filePath := inPath + "/" + hdr.Name

		fileDir := filepath.Dir(filePath)

		config.MkdirIfNotExist(fileDir)

		file, err := os.Create(filePath)
		handleError(err)
		_, err = io.Copy(file, tr)
		file.Close()
		handleError(err)

		//		if _, err := io.Copy(os.Stdout, tr); err != nil {
		//			log.Fatalln(err)
		//		}
		fmt.Println()
	}
}

func UnGz(gzFilePath string) {
	// file to read
	fr, err := os.Create(gzFilePath)
	handleError(err)
	defer fr.Close()

	df, err := os.Create(gzFilePath + ".tar")
	handleError(err)
	defer df.Close()

	gfr, err := gzip.NewReader(fr)
	handleError(err)

	io.Copy(df, gfr)
}

func CopyFile(src, dst string) (int64, error) {
	if src == dst {
		return 0, nil
	}
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}
