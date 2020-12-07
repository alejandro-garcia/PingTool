package common

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/vigneshuvi/GoDateFormat"
)

var lastColorCode string

//FormatDate (date time.Time, format string) string
func FormatDate(date time.Time, format string) string {
	return date.Format(GoDateFormat.ConvertFormat(format))
}

//PrnLog console output improvement
func PrnLog(msg string, colorCode string, noNewLine bool, noAddTime bool) {
	if len(colorCode) == 0 {
		colorCode = "yellow"
	}

	if colorCode != "yellow" {
		color.Unset()
	}
	if lastColorCode != colorCode {
		lastColorCode = colorCode

		switch colorCode {
		case "yellow":
			color.Set(color.FgYellow)
		case "green":
			color.Set(color.FgGreen)
		case "red":
			color.Set(color.FgRed)
		case "white":
			color.Set(color.FgWhite)
		default:
			color.Set(color.FgYellow)
		}
	}

	if !noAddTime {
		msg = FormatDate(time.Now(), "yyyy-mm-dd HH:MM:SS") + " - " + msg
	}

	if noNewLine {
		//fmt.Printf(msg)
		//log.Printf(msg)
		//log.Print(msg)
		fmt.Print(msg)
	} else {
		fmt.Println(msg)
		//fmt.Println(msg)
		//log.Println(msg)
		//log.Print(msg + "\n")
	}

	//agrego la traza de tiempo solo a la salida al archivo de logs.
	//if !NoAddTime {
	//   msg = time.Now().Format("02/01/2006 15:04:05") + " " + msg
	//}

	//color.Unset()

	//esta validacion es para controlar el salto de linea adentro del arreglo que va al log.
	//   if ($script:nonewlineFlag -eq $false){
	// 	  $script:lineas += $msg
	// 	if ($NoNewLine){
	// 	   $script:nonewlineFlag = $true;
	// 	}
	//   } else {
	// 	$script:lineas[$script:lineas.Count-1] += $msg
	// 	if ($NoNewLine -eq $false){
	// 	   $script:nonewlineFlag = $false;
	// 	}
	//   }
}

//Ping (addr string) bool
func Ping(addr string) bool {

	if runtime.GOOS == "windows" {
		out, _ := exec.Command("ping", addr, "-n", "3", "-w", "10").Output()
		//TODO: contemplar tambien respuestas en ingles del commando.
		result := strings.Contains(string(out), "Respuesta desde") || strings.Contains(string(out), "Response from") || strings.Contains(string(out), "Reply from")
		return result
	} else if runtime.GOOS == "linux" {
		out, _ := exec.Command("ping", addr, "-c", "3", "-W", "1", "-q").Output()
		result := !strings.Contains(string(out), ", 0% packet loss")
		return result
	} else {
		return false
	}
}

//NetUse (ipAddress string, user string, password string) bool
func NetUse(ipAddress string, user string, password string) bool {
	result := true

	out, err := exec.Command("net", "use", `\\`+ipAddress, `/user:`+user, password).Output()

	// c.Stdout = os.Stdout
	// err := c.Run()

	if err != nil {
		//common.PrnLog("Error conectando unidad de red..."+err.Error(), "red", false, false)
		//os.Exit(0)
		result = false
	} else {
		//fmt.Println("net use out: " + string(out))
		result = strings.Contains(string(out), "successfully") || strings.Contains(string(out), "correctamente")
	}
	return result
}

//NetDisconnect (ipAddress string) bool
func NetDisconnect(ipAddress string) bool {
	result := true

	c := exec.Command("net", "use", `\\`+ipAddress, "/delete", "2>&1>null")
	c.Stdout = os.Stdout
	err := c.Run()

	if err != nil {
		result = false
	}
	//  else {
	// 	result = strings.Contains(string(out), "successfully")
	// }
	return result
}

//Contains (s []int, searchterm int) bool
func Contains(s []int, searchterm int) bool {
	i := sort.SearchInts(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

//IndexOf (element string, data []string) int
func IndexOf(element string, data []string) int {
	for k, v := range data {
		if strings.ToLower(element) == strings.ToLower(v) {
			return k
		}
	}
	return -1
}

// GetOSSeparator get appropiate separator for host
func GetOSSeparator() string {
	if runtime.GOOS == "windows" {
		return "\\"
	}
	return "/"
}

