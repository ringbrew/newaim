package product

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"github.com/ringbrew/newaim/productsearch/internal/domain/product"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type DataReader struct {
	Path string
}

func NewDataReader(path string) *DataReader {
	return &DataReader{
		Path: path,
	}
}

func (dr *DataReader) Read() ([]*product.Product, error) {
	dest := "data/"
	if err := dr.unzipSource(dr.Path, dest); err != nil {
		log.Fatal(err.Error())
	}

	f, err := os.Open("data/sku_list.csv")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	result := make([]*product.Product, 0, 100000)
	reader := csv.NewReader(f)
	index := 0
	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err.Error())
		}

		index++

		result = append(result, &product.Product{
			SKU:         row[0],
			Title:       row[1],
			Description: row[2],
		})
	}

	return result, nil
}

func (dr *DataReader) unzipSource(source, destination string) error {
	// 1. Open the zip file
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 2. Get the absolute destination path
	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}

	// 3. Iterate over zip files inside the archive and unzip each of them
	for _, f := range reader.File {
		err := dr.unzipFile(f, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dr *DataReader) unzipFile(f *zip.File, destination string) error {
	// 4. Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(destination, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// 5. Create directory tree
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// 6. Create a destination file for unzipped content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// 7. Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return err
	}
	return nil
}
