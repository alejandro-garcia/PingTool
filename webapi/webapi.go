package webapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/alejandro-garcia/pingtool/common"
	"github.com/alejandro-garcia/pingtool/core"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type connectedInfo struct {
	count string
	list  string
}

type warehouse struct {
	ID     string `json:"id"`
	IP     string `json:"ip"`
	Online bool   `json:"online"`
}

//StartWebServer package main function
func StartWebServer() {
	core.SetupAllRestaurants()
	r := mux.NewRouter()
	r.HandleFunc("/warehouses", getWarehouses).Methods("GET")
	r.HandleFunc("/warehouseList", getWarehousesList).Methods("GET")
	r.HandleFunc("/warehouses-text", getWarehousesText).Methods("GET")
	r.HandleFunc("/sayHello", sayHello).Methods("GET")
	// handlers below are from private package
	r.HandleFunc("/warehouseInfo/{id}", getWarehouseInfo).Methods("GET")
	r.HandleFunc("/warehouseNonUpdated/{app}", getWarehouseNonUpdated).Methods("GET")
	r.HandleFunc("/duplicatedSalesHeader/{id}/{begindate}/{enddate}", getSalesPrinterDuplicatesHeaders).Methods("GET")
	r.HandleFunc("/warehouseApps", getWarehouseApps).Methods("GET")
	common.PrnLog("Iniciando API Web en el Puerto :4500", "yellow", false, false)
	c := cors.New(cors.Options{
		AllowedOrigins:   core.Config.CorsAllowedAddress,
		AllowCredentials: true,
	})
	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":4500", handler))
}

func getWarehousesText(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json")
	conn, disconn := connectedWarehousesText()
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Servidores Conectados ("+conn.count+"): "+conn.list+"\r\n")
	io.WriteString(w, "Servidores No-Conectados ("+disconn.count+"): "+disconn.list)
	//json.NewEncoder(w).Encode(warehouses)
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Hola puto!")
}

func connectedWarehousesText() (connectedInfo, connectedInfo) {
	var connected []string
	var disconnected []string
	var resultConnected connectedInfo
	var resultDisconnected connectedInfo

	var wg sync.WaitGroup

	//queue := make(chan Warehouse, 1)

	j := 0
	for k, v := range core.Servers {
		wg.Add(1)
		go func(rest string, ip string) {
			j++
			// usr, pass := setupWarehouseCredential(rest)
			common.PrnLog(fmt.Sprintf("procesando (%d) restaurant: %s ip: %s", j, rest, ip), "yellow", false, false)
			//checkUpdaterVersion(rest, ip, usr, pass)

			if common.Ping(ip) {
				connected = append(connected, rest)
			} else {
				disconnected = append(disconnected, rest)
			}

			wg.Done()
		}(k, v)
	}

	// go func(){
	// 	defer wg.Done()
	// 	for t:= range queue {

	// 	}
	// }

	wg.Wait()

	resultConnected.count = strconv.Itoa(len(connected))
	resultConnected.list = ""
	resultDisconnected.count = strconv.Itoa(len(disconnected))
	resultDisconnected.list = ""

	if len(connected) > 0 {
		sort.Sort(sort.StringSlice(connected))
		resultConnected.list = strings.Join(connected, ",")
	}

	if len(disconnected) > 0 {
		sort.Sort(sort.StringSlice(disconnected))
		resultDisconnected.list = strings.Join(disconnected, ",")
	}

	fmt.Println("*** WAREHOUSES CONECTADOS *** ")
	fmt.Println(connected)
	fmt.Println("*** FIN: WAREHOUSES CONECTADOS *** ")

	fmt.Println("*** WAREHOUSES DES-CONECTADOS *** ")
	fmt.Println(disconnected)
	fmt.Println("*** FIN: WAREHOUSES DES-CONECTADOS *** ")

	return resultConnected, resultDisconnected
}

func connectedWarehouses() []warehouse {
	var result []warehouse
	var wg sync.WaitGroup

	queue := make(chan warehouse, 1)

	j := 0
	for k, v := range core.Servers {
		wg.Add(1)
		go func(rest string, ip string) {
			j++
			// usr, pass := setupWarehouseCredential(rest)
			common.PrnLog(fmt.Sprintf("procesando (%d) restaurant: %s ip: %s", j, rest, ip), "yellow", false, false)
			//checkUpdaterVersion(rest, ip, usr, pass)
			//result = append(result, Warehouse{Id: rest, Ip: ip, Online: handlePing(rest, ip)})
			queue <- warehouse{ID: rest, IP: ip, Online: common.Ping(ip)}

			//wg.Done()
		}(k, v)
	}

	go func() {
		// 	defer wg.Done()
		for t := range queue {
			result = append(result, t)
			wg.Done()
		}
	}()

	wg.Wait()

	//sort result slice
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

func getWarehouses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	warehouses := connectedWarehouses()
	json.NewEncoder(w).Encode(warehouses)
}

func getWarehousesList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if core.Config.Servers != nil {
		allServers := append([]string{"(todos)"}, core.Config.Servers...)
		json.NewEncoder(w).Encode(allServers)
	} else {
		var emptyList []string
		json.NewEncoder(w).Encode(emptyList)
	}
}
