package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alejandro-garcia/pingtool/common"
)

type prefixes struct {
	Server string
	Prefix string
}

//ConfigModel core struct
type ConfigModel struct {
	Prefixes                     []prefixes
	Servers                      []string
	AlternateAdmin               []string
	NoAdmin                      []string
	UserName                     string
	DefaultPassWord              string
	DefaultUserAlternatePassWord string
	AlternateUser                string
	AlternatePassMask            string
	ServersIPMask                string
	CorsAllowedAddress           []string
	CurrentVersions              map[string]string
}

//Servers global map
var Servers = make(map[string]string)

//Config global struct concrete
var Config ConfigModel

//FormatIP (restCode string) string
func FormatIP(restCode string) string {
	restNum, _ := strconv.Atoi(restCode)

	if (restNum >= 200 && restNum <= 214) || restNum >= 401 {
		restNum -= 200
	}

	return fmt.Sprintf(Config.ServersIPMask, restNum)
}

//SetupAllRestaurants (servers []string)
func SetupAllRestaurants() {
	for _, rest := range Config.Servers {
		Servers[rest] = FormatIP(rest)
	}
}

//SetupCustomRestaurants (warehouses string)
func SetupCustomRestaurants(warehouses string) {
	warehouseList := strings.Split(warehouses, ",")
	for _, itm := range warehouseList {
		if _, ok := Servers[itm]; !ok {
			restNum, err := strconv.Atoi(itm)
			if err == nil && restNum >= 101 && restNum <= 499 {
				Servers[itm] = FormatIP(itm)
			}
		}
	}
}

//RequestCustomRestaurants prompts for custom restaurants
func RequestCustomRestaurants() {
	fmt.Println("Ingrese el nro. de los de restaurantes seguido de <enter>")
	fmt.Println("cuando finalize presione <enter> sin ingresar mas codigos")
	fmt.Println("o para cancelar presione <ctrl>+c para salir")

	var tmpRestCode string

	for {
		tmpRestCode = ""
		fmt.Print("idRest:")
		fmt.Scanln(&tmpRestCode)

		//fmt.Println("*" + tmpRestCode + "*")

		if len(tmpRestCode) == 0 {
			break
		} else {
			if _, ok := Servers[tmpRestCode]; !ok {
				restNum, err := strconv.Atoi(tmpRestCode)
				if err == nil && restNum >= 101 && restNum <= 499 {
					Servers[tmpRestCode] = FormatIP(tmpRestCode)
				}
			}
		}
	}
}

//SetupWarehouseCredential (restCode string) (string, string)
func SetupWarehouseCredential(restCode string) (string, string) {
	user := Config.UserName
	password := Config.DefaultPassWord

	if common.IndexOf(restCode, Config.AlternateAdmin) != -1 {
		password = Config.DefaultUserAlternatePassWord
	} else if common.IndexOf(restCode, Config.NoAdmin) != -1 {
		password = ""

		for _, v := range Config.Prefixes {
			if v.Server == restCode {
				user = Config.AlternateUser
				password = fmt.Sprintf(Config.AlternatePassMask, v.Server, v.Prefix)
				break
			}
		}
	}

	return user, password
}

// ParseVersionNumber takes version of format: 1.0.0.0 and returns it on format: 1.0.0
func ParseVersionNumber(version string) string {
	return strings.Join(strings.Split(version, ".")[:3], ".")
}
