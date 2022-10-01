package common

import (
    "archive/tar"
    "compress/gzip"
    "errors"
    "fmt"
    "io"
    "os"
    "strings"
)

func GetFileExtension(filePath string) string {
    pathTokens := strings.Split(filePath, ".")
    return pathTokens[len(pathTokens)-1]
}

func GetFileNameFromPath(filePath string) string {
    pathTokens := strings.Split(filePath, "/")
    return pathTokens[len(pathTokens)-1]
}

func LocalFileExists(fileName string) bool {

    info, err := os.Stat(fileName)
    if os.IsNotExist(err) {
        return false
    }

    return !info.IsDir()
}

func RemoveLocalFile(filePath string) error {

    if LocalFileExists(filePath) {
        return os.Remove(filePath)
    }

    return nil
}

func SaveBytesToFile(fileData io.Reader, filePath string) error {

    if err := RemoveLocalFile(filePath); err != nil {
        return errors.New(fmt.Sprintf("failed to remove existing output file: %v", err))
    }

    outputFile, err := os.Create(filePath)
    if err != nil {
        return errors.New(fmt.Sprintf("failed to create file at %s: %v", filePath, err))
    }
    defer outputFile.Close()

    _, err = io.Copy(outputFile, fileData)
    if err != nil {
        return errors.New(fmt.Sprintf("failed to copy file data to buffer: %v", err))
    }

    return nil
}

func ExtractFileFromGZIP(fileData io.Reader, filePath string) error {

    if err := RemoveLocalFile(filePath); err != nil {
        return errors.New(fmt.Sprintf("failed to remove existing output file: %v", err))
    }

    outputFile, err := os.Create(filePath)
    if err != nil {
        return errors.New(fmt.Sprintf("failed to create file at %s: %v", filePath, err))
    }
    defer outputFile.Close()

    outputFileName := GetFileNameFromPath(filePath)

    // Decompress file in memory
    gz, err := gzip.NewReader(fileData)
    if err != nil {
        return err
    }

    tr := tar.NewReader(gz)
    for {
        var hdr *tar.Header

        hdr, err = tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return errors.New(fmt.Sprintf("failed to read from GZIP file: %v", err))
        }

        if GetFileNameFromPath(hdr.Name) == outputFileName {
            _, err = io.Copy(outputFile, tr)
            if err != nil {
                return errors.New(fmt.Sprintf("failed to copy file data to buffer: %v", err))
            }
        }
    }

    gz.Close()

    return nil
}
