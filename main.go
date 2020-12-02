package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alejandro-garcia/pingtool/common"
	"github.com/alejandro-garcia/pingtool/core"
	"github.com/alejandro-garcia/pingtool/iohelpers"
	"github.com/alejandro-garcia/pingtool/webapi"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
)

var serverCode string
var verifyDate string
var option string
var invokeCmd string
var invokeArgs []string

var serverID string
var expectedValue string
var basePath string
var wait bool

//var config interface{}
var svrnotconnected []string
var svrconnected []string

func parseCommandPath(command string) string {
	result := command
	separator := common.GetOSSeparator()
	if !strings.Contains(command, separator) {
		result = basePath + separator + command
	}
	return result
}

func executeCommand(command string, rest string, ip string, user string, password string, args ...string) {
	//powershell -ExecutionPolicy ByPass -File ps-setup.ps1
	fileExtension := iohelpers.GetFileExtension(command)

	//type Cmd exec.Cmd
	//var cmd
	switch fileExtension {
	case "ps1":
		//powershellCommand := "powershell"
		newArgs := append([]string{"-ExecutionPolicy", "ByPass", "-File", parseCommandPath(command)}, args...)
		cmd := exec.Command("powershell", newArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		common.PrnLog(rest+" : ejecutando comando/script externo", "yellow", false, false)
		err := cmd.Run()
		if err != nil {
			common.PrnLog("fallo la invocación del comando: "+command+" / error : "+err.Error(), "red", false, false)
		}
	case "exe":
		//$psstatus= e:\agarcia\psexec.exe \\$ipaddress -u $usr -p $clave -accepteula -h -c -f "$($scriptPath)\$($setupFileName)" "-s"
		newArgs := append([]string{"\\\\" + ip, "-u", user, "-p", password, "-accepteula", "-h", "-c", "-f", parseCommandPath(command)}, args...)
		if len(args) == 0 || len(newArgs[10]) == 0 {
			newArgs[10] = "-s"
		} else {
			newArgs = append(newArgs, "-s")
		}

		for idx, arg := range newArgs {
			fmt.Printf("InvokeArgs[%d] : %s\n", idx, arg)
		}

		cmd := exec.Command("psexec.exe", newArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		common.PrnLog(rest+" : ejecutando comando/script externo", "yellow", false, false)
		err := cmd.Run()
		if err != nil {
			common.PrnLog("fallo la invocación del comando: "+command+" / error : "+err.Error(), "red", false, false)
		}
	}
}

// Pattern to match a windows absolute path: "c:\" and similar
var isAbsWinDrive = regexp.MustCompile(`^[a-zA-Z]\:\\`)

func displayHelp() {
	fmt.Println("pingtool.exe [-serverCode ###] [-option p|w] [-w] [-invokecmd <script o ejecutable> [-invokeargs <argumentos>]] [--notify] [--help|/?]")
	fmt.Println("   serverCode: codigo del restaurante o lista de restaurantes separadas por coma")
	fmt.Println("   option: opciones disponibles: ")
	fmt.Println("           p : [default] verifica que servidores responden al ping")
	fmt.Println("	w: Levanta servidor web (api-rest) en el puerto localhost:4500")
	fmt.Println("   notify: se usa con la opción: p para revisar la lista de servidores en secuencia y esperar hasta que se conecten")
	fmt.Println("   	  todos los servidores seleccionados, con reintentos cada 5 min.")
	fmt.Println("   invokecmd <script o ejecutable>: permite ejecutar un script o .exe cuando se conecten el o los servidores seleccionados.")
	fmt.Println("   invokeargs <lista de argumentos>: especifica la lista de argumentos que se le pasaran al script ó ejecutable")
	fmt.Println("   help ó /? : muestra el mensaje de ayuda")
}

func displayArgs() {
	if serverCode != "" || verifyDate != "" || option != "" {
		fmt.Println("Parametros cmdline:")
	}

	if serverCode != "" {
		fmt.Printf("serverCode: %s\n", serverCode)
	}

	if option != "" {
		fmt.Printf("option : %s\n", option)
	}

	if invokeCmd != "" {
		fmt.Printf("InvokeCmd : %s\n", invokeCmd)
	}

	if len(invokeArgs) > 0 {
		for idx, arg := range invokeArgs {
			fmt.Printf("InvokeArgs[%d] : %s\n", idx, arg)
		}
	}
}

func parseArg(val string) string {
	var result string = ""
	pos := common.IndexOf(val, os.Args)
	if pos != -1 {
		switch val {
		case "--help", "/?":
			displayHelp()
			os.Exit(0)
		case "-w":
			result = "w"
		case "--notify":
			wait = true
		default:
			result = os.Args[pos+1]
		}
	}

	return result
}

func parseArgs() {
	parseArg("--help")
	parseArg("/?")
	parseArg("--notify")
	serverCode = parseArg("-serverCode")
	//verifyDate = parseArg("-verifydate")
	//expectedValue = parseArg("-expected")
	option = parseArg("-w")
	invokeCmd = parseArg("-invokecmd")
	invokeArgs = strings.Split(parseArg("-invokeargs"), " ")

	if len(option) == 0 {
		if len(invokeCmd) > 0 {
			option = "p"
		} else {
			option = parseArg("-option")
		}
	}
}

func handlePing(serverCode string, ipaddress string) bool {
	if !common.Ping(ipaddress) {
		if common.IndexOf(serverCode, svrnotconnected) == -1 {
			svrnotconnected = append(svrnotconnected, serverCode)
		}
		return false
	}

	svrconnected = append(svrconnected, serverCode)
	return true
}

func requestWaitOption() {
	var waitOpt string

	for {
		fmt.Println("Desea notificar de forma secuencial los servidores detectados (s/n): ")
		fmt.Scanln(&waitOpt)
		r, _ := regexp.Compile("^[sn]{1}")
		if r.MatchString(waitOpt) {
			break
		}
	}

	if strings.ToLower(waitOpt) == "s" {
		wait = true
	}
}

func subMenu() string {
	var opt string = ""

	for {
		if len(option) != 0 {
			opt = option
			break
		}

		fmt.Println("Opciones disponibles: ")
		fmt.Println("")
		fmt.Println("[P] Verificar conectividad (PING) de los servidores")
		fmt.Println("[Q] para salir")

		fmt.Print("ingrese la opcion deseada: ")
		fmt.Scanln(&opt)

		r, _ := regexp.Compile("^[pq]{1}")

		if r.MatchString(opt) {
			break
		} else {
			c := exec.Command("cmd", "/c", "cls")
			c.Stdout = os.Stdout
			c.Run()
		}
	}

	opt = strings.ToLower(opt)

	if (serverID == "t" || len(core.Servers) > 1) && opt == "p" && len(os.Args) == 1 {
		requestWaitOption()
	}

	return opt
}

func mainMenu() string {
	for {
		if len(serverCode) != 0 {
			serverID = serverCode
			break
		}

		if option == "w" {
			serverID = "w"
			break
		}

		fmt.Println("Indique el codigo de restaurante que quiere verificar o escriba T para verificarlos todos!")
		fmt.Println("[C] personalizar la lista de servidores a configurar")
		fmt.Println("[T] revisar todos los restaurantes")
		fmt.Println("[W] Iniciar API-WEB")
		fmt.Println("[Q] para salir...")

		fmt.Print("Coloque el numero del ambiente:")

		fmt.Scanln(&serverID)

		if strings.Contains("t|c|q|w|,", serverID) {
			//fmt.Printf("restaurante/opcion: %s", serverID)
			break
		} else if strings.Contains(serverID, ",") {
			break
		} else if _, err := strconv.Atoi(serverID); err == nil {
			//fmt.Printf("restaurante/opcion: %s", serverID)
			core.Servers[serverID] = core.FormatIP(serverID)
			break
		}
	}

	switch strings.ToLower(serverID) {
	case "c":
		core.RequestCustomRestaurants()
	case "t":
		core.SetupAllRestaurants()
	case "w":
		webapi.StartWebServer()
	case "q":
		os.Exit(0)
	default:
		core.SetupCustomRestaurants(serverID)
	}

	// if strings.Contains(serverID, ",") {
	// 	setupCustomRestaurants(serverID)
	// }

	return subMenu()
}

func wellcome() {
	fmt.Println("############################################################################################")
	fmt.Println("############################################################################################")
	fmt.Println("############################## UTILITARIO PING (ASINCRONO) #################################")
	fmt.Println("############################################################################################")
	fmt.Println("############################################################################################")
	fmt.Println(" ")
}

func checkConnectedServers() {
	var wg sync.WaitGroup

	common.PrnLog(fmt.Sprintf("Conteo de Servidores: %d", len(core.Servers)), "yellow", false, false)

	//wg.Add(len(core.Servers))
	j := 0
	for k, v := range core.Servers {
		wg.Add(1)
		go func(rest string, ip string) {
			j++
			// usr, pass := setupWarehouseCredential(rest)
			common.PrnLog(fmt.Sprintf("procesando (%d) restaurant: %s ip: %s", j, rest, ip), "yellow", false, false)
			//checkUpdaterVersion(rest, ip, usr, pass)

			if handlePing(rest, ip) && len(invokeCmd) > 0 {
				invokeCommand(rest, ip)
			}

			wg.Done()
		}(k, v)
	}

	wg.Wait()

	if len(core.Servers) > 1 {
		if len(svrconnected) > 0 {
			sort.Sort(sort.StringSlice(svrconnected))
			connectedMsg := fmt.Sprintf("Servidores conectados (%d): %s", len(svrconnected), strings.Join(svrconnected, ","))
			common.PrnLog(connectedMsg, "yellow", false, false)
		}

		if len(svrnotconnected) > 0 {
			sort.Sort(sort.StringSlice(svrnotconnected))
			common.PrnLog(fmt.Sprintf("Servidores que no responden al ping (%d): %s", len(svrnotconnected), strings.Join(svrnotconnected, ",")), "yellow", false, false)
		}
	} else {
		if len(svrconnected) > 0 {
			common.PrnLog("Servidor Disponible!", "green", false, false)
		} else {
			common.PrnLog("Servidor No-Disponible!", "red", false, false)
		}
	}
}

func invokeCommand(rest string, ip string) {
	if len(invokeCmd) > 0 {
		cmdArgs := append([]string{}, invokeArgs...)

		placeHolder1 := common.IndexOf("{0}", cmdArgs)
		placeHolder2 := common.IndexOf("{1}", cmdArgs)

		if placeHolder1 != -1 {
			cmdArgs[placeHolder1] = rest
		}
		if placeHolder2 != -1 {
			cmdArgs[placeHolder2] = ip
		}

		user, pwd := core.SetupWarehouseCredential(rest)

		executeCommand(invokeCmd, rest, ip, user, pwd, cmdArgs...)
	}
}

func waitForServers(serversList map[string]string) {
	j := 0

	var keysToRemove []string

	for rest, ip := range serversList {
		j++
		// usr, pass := setupWarehouseCredential(rest)
		common.PrnLog(fmt.Sprintf("procesando (%d) restaurant: %s ip: %s", j, rest, ip), "yellow", false, false)
		//checkUpdaterVersion(rest, ip, usr, pass)
		if handlePing(rest, ip) {
			keysToRemove = append(keysToRemove, rest)
			iconPath := fmt.Sprintf("%s%s%s", basePath, common.GetOSSeparator(), "netfol.ico")
			// notification := toast.Notification{
			// 	AppID:   "Pintool",
			// 	Title:   "Servidor Disponible",
			// 	Message: "El Servidor " + rest + " esta disponible",
			// 	Icon:    iconPath,
			// }
			// err := notification.Push()
			// if err != nil {
			// 	common.PrnLog("Error generando mensaje de notificación (toast)", "red", false, false)
			// 	log.Fatalln(err)
			// }

			err := beeep.Notify("Servidor Disponible", "El Servidor "+rest+" esta disponible", iconPath)
			if err != nil {
				common.PrnLog("Error generando mensaje de notificación (toast)", "red", false, false)
				log.Fatalln(err)
			}

			common.PrnLog("El servidor "+rest+" esta disponible!", "green", false, false)

			invokeCommand(rest, ip)
		}
	}

	for _, key := range keysToRemove {
		_, ok := serversList[key]
		if ok {
			delete(serversList, key)
		}
	}

	if len(serversList) > 0 {
		if len(svrnotconnected) > 0 {
			sort.Sort(sort.StringSlice(svrnotconnected))
			common.PrnLog(fmt.Sprintf("Servidores que no responden al ping (%d): %s", len(svrnotconnected), strings.Join(svrnotconnected, ",")), "yellow", false, false)
		}
		fmt.Println("Esperando 5 minutos para reintentar")
		time.Sleep(5 * time.Minute)
		svrnotconnected = nil
		waitForServers(serversList)
	} else {
		fmt.Println("Ya no quedan servidores por revisar... finalizando el proceso!")
	}
}

func main() {
	// es es solo para que compile la app ya que obliga que se usen todos las funciones y variables definidas
	//ping("ve-spfe2")
	//fmt.Printf("ping test: %t\n", ping("ve-spfe2"))
	//	fmt.Println(getRestTermsCounter("111"))
	parseArgs()
	wellcome()
	displayArgs()

	var err error
	basePath, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		common.PrnLog("error obteniendo la ruta de ejecución del programa", "red", false, false)
		return
	}
	configPath := fmt.Sprintf("%s%s%s", basePath, common.GetOSSeparator(), "Config.json")
	//fmt.Println(configPath)
	jsonBinary, err := ioutil.ReadFile(configPath)
	if err != nil {
		common.PrnLog("error cargando el archivo de configuración VpSalesConfig.json", "red", false, false)
		return
	}

	core.Config = core.ConfigModel{}

	jobjErr := json.Unmarshal(jsonBinary, &core.Config)
	if jobjErr != nil {
		common.PrnLog("error leyendo el archivo de configuración VpSalesConfig.json", "red", false, false)
		return
	}

	opt := mainMenu()

	switch strings.ToLower(opt) {
	case "q":
		os.Exit(0)
	case "p":
		if wait {
			waitForServers(core.Servers)
		} else {
			checkConnectedServers()
		}
	}

	color.Unset()
}
