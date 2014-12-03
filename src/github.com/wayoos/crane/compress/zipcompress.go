package compress

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	singleFileByteLimit = 107374182400 // 1 GB
	chunkSize           = 4096         // 4 KB
)

func copyContents(r io.Reader, w io.Writer) error {
	var size int64
	b := make([]byte, chunkSize)
	for {
		// we limit the size to avoid zip bombs
		size += chunkSize
		if size > singleFileByteLimit {
			return errors.New("file too large, please contact us for assistance")
		}
		// read chunk into memory
		length, err := r.Read(b[:cap(b)])
		if err != nil {
			if err != io.EOF {
				return err
			}
			if length == 0 {
				break
			}
		}
		// write chunk to zip file
		_, err = w.Write(b[:length])
		if err != nil {
			return err
		}
	}
	return nil
}

// We need a struct internally because the filepath WalkFunc
// doesn't allow custom params. So we save them here so it can
// access them
type zipper struct {
	srcFolder string
	destFile  string
	writer    *zip.Writer
}

// internal function to zip a file, called by filepath.Walk on each file
func (z *zipper) zipFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	// only zip files (directories are created by the files inside of them)
	// TODO allow creating folder when no files are inside
	if !f.Mode().IsRegular() || f.Size() == 0 {
		return nil
	}
	// open file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	// create new file in zip
	fileName := strings.TrimPrefix(path, z.srcFolder+"/")
	w, err := z.writer.Create(fileName)
	if err != nil {
		return err
	}
	// copy contents of the file to the zip writer
	err = copyContents(file, w)
	if err != nil {
		return err
	}
	return nil
}

// internal function to zip a folder
func (z *zipper) zipFolder() error {
	// create zip file
	zipFile, err := os.Create(z.destFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	// create zip writer
	z.writer = zip.NewWriter(zipFile)
	err = filepath.Walk(z.srcFolder, z.zipFile)
	if err != nil {
		return nil
	}
	// close the zip file
	err = z.writer.Close()
	if err != nil {
		return err
	}
	return nil
}

// ZipFolder zips the given folder to the a zip file
// with the given name
func ZipFolder(srcFolder string, destFile string) error {
	z := &zipper{
		srcFolder: srcFolder,
		destFile:  destFile,
	}
	return z.zipFolder()
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	//	os.MkdirAll(dest, 0755)

	destInfo, err := os.Stat(dest)
	if err != nil {
		return err
	}
	destMode := destInfo.Mode()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		//println("zip entry: " + f.Name)

		outpath := filepath.Join(dest, f.Name)

		//println("zip entry path: " + outpath)

		pathonly := path.Dir(outpath)

		//println("zip entry abs path: " + pathonly)

		if _, err := os.Stat(pathonly); os.IsNotExist(err) {
			os.MkdirAll(pathonly, destMode)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(outpath, f.Mode())
		} else {
			f, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
