package iohelpers

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/alejandro-garcia/pingtool/common"
)

//FileExists (filePath string) bool
func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//FolderExists (folderPath string) bool
func FolderExists(folderPath string) bool {
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

//GetFileLastWriteTime (filePath string, restCode string)
func GetFileLastWriteTime(filePath string, restCode string) {
	fileName := getFileName(filePath)

	if FileExists(filePath) {
		common.PrnLog(restCode+" : Archivo "+fileName, "yellow", true, false)
		common.PrnLog("- NO ENCONTRADO", "red", false, true)
		return
	}

	info, err := os.Stat(filePath)
	if err == nil {
		lastWriteTime := common.FormatDate(info.ModTime(), "dd/mm/yyyy HH:MM:SS")
		common.PrnLog(restCode+" : Fecha "+fileName+" : "+lastWriteTime, "yellow", false, false)
	}
}

//ReadTextFileContent (filePath string) string
func ReadTextFileContent(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		//log.Fatal(err)
		rest := strings.Split(filePath[2:strings.LastIndex(filePath, "bootdrv")-2], ".")[2]
		common.PrnLog(rest+" : "+err.Error(), "red", false, false)
		return ""
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)

	var result string
	if err != nil {
		result = ""
	} else {
		result = string(b)
	}

	return result
}

func getFileName(filePath string) string {
	return filePath[strings.LastIndex(filePath, common.GetOSSeparator())+1 : len(filePath)]
}

//GetFileExtension (filePath string) string
func GetFileExtension(filePath string) string {
	//powershell -ExecutionPolicy ByPass -File ps-setup.ps1
	return filePath[strings.LastIndex(filePath, ".")+1 : len(filePath)]
}

func getFolderLastWriteTime(folderPath string, restCode string) {
	term := folderPath[strings.LastIndex(folderPath, common.GetOSSeparator())+1 : len(folderPath)]
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		common.PrnLog(restCode+" : Carpeta "+term+" : ", "yellow", true, false)
		common.PrnLog("- NO ENCONTRADA", "red", false, true)
	} else if err == nil {
		lastWriteTime := common.FormatDate(info.ModTime(), "dd/mm/yyyy HH:MM:SS")
		common.PrnLog(restCode+" : Carpeta "+term+" : "+lastWriteTime, "yellow", false, false)
	}
}
