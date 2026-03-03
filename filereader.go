package gtfsparser

import (
	"archive/zip"
	"strings"
)

func (gtfs *GTFS) readzip() ([]string, error) {
	var files []string

	// Open the zip file
	r, err := zip.OpenReader(gtfs.FileName)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Collect only .txt files
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".txt") {
			files = append(files, f.Name)
		}
	}

	return files, nil
}

func (gtfs *GTFS) VerifyFileExists() ([]string, error) {
	filenames, err := gtfs.readzip()

	return filenames, err
}
