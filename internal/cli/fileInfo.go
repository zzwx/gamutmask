package cli

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileInfoList struct {
	Items []*FileInfo
}

type FileInfo struct {
	InputName   string
	OutputName  string
	MD5         string    `json:",omitempty"`
	Size        int64     // Excessive data in case MD5 appears the same
	CreatedAt   time.Time // Excessive data in case MD5 appears the same
	ProcessedAt time.Time
	Width       int
	Height      int
	// A hidden flag for processing removal data from JSON only
	FileFound bool `json:"-"`
}

func unmarshalJSON(r io.Reader) (*FileInfoList, error) {
	var result FileInfoList
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func marshalJSON(w io.Writer, data *FileInfoList) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

var mtx sync.Mutex

func Supported(fileName string) bool {
	switch filepath.Ext(fileName) {
	case ".jpg", ".jpeg", ".png":
		return true
	}
	return false
}

func Decode(f *os.File) image.Image {
	var img image.Image
	var err error
	switch filepath.Ext(f.Name()) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
		if err != nil {
			panic(err)
		}
	case ".png":
		img, err = png.Decode(f)
		if err != nil {
			panic(err)
		}
	default:
		panic(filepath.Ext(f.Name()) + " Unsupported image type")
	}
	return img
}

// ProcessChangedFilesOnly will check if the input folder has any changes and update output folder + the JSON file
// and execute runFunction(outputFileFullPath,inputFileFullPath) against each changed file
func ProcessChangedFilesOnly(inputFolderName string, outputFolderName string,
	runFunction func(string, string, *RunGamutSettings, func(*os.File) image.Image) (int, error),
	settings *RunGamutSettings) {

	// We don't want this function to be called simultaneously
	mtx.Lock()
	defer mtx.Unlock()

	JSONFileName := inputFolderName + "/_list.json"
	osInputFolderFiles, err := ioutil.ReadDir(inputFolderName)
	if err != nil {
		panic(err)
	}

	fileInfoList := FileInfoList{}

	// Restore data from the file, if exists
	file, err := os.Open(JSONFileName)
	if err != nil {
		// If the file doesn't exist, we'll generate it
	} else {
		result, err := unmarshalJSON(file)
		if err != nil {
			//panic(err)
			// We just skip if the file doesn't contain right data
		} else {
			fileInfoList.Items = append(fileInfoList.Items, result.Items...)
		}
	}
	defer file.Close()

	// Now we'll read the folder and see if data has existed in JSON
	for _, f := range osInputFolderFiles {
		inputFileName := f.Name()
		if !f.IsDir() && Supported(inputFileName) {
			outputFileName := inputFileName + ".png" // Always appending .png

			foundIndex := -1
			processedMD5 := "" // To avoid calculating it twice
			processIt := false
			for index, item := range fileInfoList.Items {
				if item.InputName == inputFileName {
					foundIndex = index
					newMD5 := getFileMD5(inputFolderName + "/" + inputFileName)
					// The file has changed?
					if item.Size != f.Size() || item.CreatedAt != f.ModTime() || item.MD5 != newMD5 {
						processIt = true
						processedMD5 = newMD5
					}
					// The output file doesn't exist?
					if _, err := os.Stat(outputFolderName + "/" + outputFileName); os.IsNotExist(err) {
						processIt = true
						processedMD5 = newMD5
					}
					item.FileFound = true // Mark it as found so it won't be removed
					break
				}
			}

			if foundIndex < 0 || processIt {
				runFunction(outputFolderName+"/"+outputFileName, inputFolderName+"/"+inputFileName, settings, Decode)
				if processedMD5 == "" {
					processedMD5 = getFileMD5(inputFolderName + "/" + inputFileName)
				}
				if foundIndex < 0 {
					fileInfoList.Items = append(fileInfoList.Items, &FileInfo{
						InputName:   inputFileName,
						OutputName:  outputFileName,
						MD5:         processedMD5,
						Size:        f.Size(),
						CreatedAt:   f.ModTime(),
						ProcessedAt: time.Now(),
						FileFound:   true, // Mark it as found so it won't be removed
					})
				} else {
					fileInfoList.Items[foundIndex].MD5 = processedMD5
					fileInfoList.Items[foundIndex].Size = f.Size()
					fileInfoList.Items[foundIndex].CreatedAt = f.ModTime()
					fileInfoList.Items[foundIndex].ProcessedAt = time.Now()
					fileInfoList.Items[foundIndex].FileFound = true // Mark it as found so it won't be removed
				}
			}
		}
	}
	// Get rid of the ones were in JSON that were not found as files (due to "fileFound" flag)
	for index := 0; index < len(fileInfoList.Items); index++ {
		if !fileInfoList.Items[index].FileFound {
			fileInfoList.Items = append(fileInfoList.Items[:index], fileInfoList.Items[index+1:]...)
			index-- // To make it re-run the index that will be index++
		}
	}
	// And now we'll remove the unnecessary files from output folder that don't exist in JSON
	// Remove unnecessary files from output folder
	SanitizeOutputFolder(outputFolderName, fileInfoList)

	out, err := os.Create(JSONFileName)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	marshalJSON(out, &fileInfoList)

}

// Remove files that were not found in JSON's fileInfoList
func SanitizeOutputFolder(outputFolderName string, fileInfoList FileInfoList) {
	outputFolderFiles, err := ioutil.ReadDir(outputFolderName)
	if err != nil {
		panic(err)
	}
	for _, f := range outputFolderFiles {
		if !f.IsDir() && !isInFileInfoListAsOutputName(fileInfoList.Items, f.Name()) && f.Name() != ".gitignore" {
			fmt.Printf("Deleting: %v\n", outputFolderName+"/"+f.Name())
			if err := os.Remove(outputFolderName + "/" + f.Name()); err != nil {
				panic(err)
			}
		}
	}

}

func isInFileInfoListAsInputName(list []*FileInfo, item string) bool {
	for _, l := range list {
		if l.InputName == item {
			return true
		}
	}
	return false
}

func isInFileInfoListAsOutputName(list []*FileInfo, item string) bool {
	for _, l := range list {
		if l.OutputName == item {
			return true
		}
	}
	return false
}

// GetFileMD5 will open the file, calculate and return its MD5 as a sequence of Hex symbols
func getFileMD5(fileName string) string {
	const fileChunk = 8192 // we settle for 8KB
	file, err := os.Open(fileName)
	if err != nil {
		return "" // skip the file that has been removed
		//panic(err)
	}
	defer file.Close()
	info, _ := file.Stat()
	fileSize := info.Size()
	blocks := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	hash := md5.New()
	for i := uint64(0); i < blocks; i++ {
		blockSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		buf := make([]byte, blockSize)
		file.Read(buf)
		io.WriteString(hash, string(buf)) // append into the hash
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
