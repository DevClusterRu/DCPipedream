package main

import (
	"DeviceCloudPusher/components"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Consumer map[int]string

type Hook struct {
	CloudUrl string `json:"cloud_url"`
	Did int `json:"did"`
	DeviceStatus string `json:"device_status"`
}

func hookCatch(w http.ResponseWriter, req *http.Request) {
	var hook Hook
	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(req.Body).Decode(&hook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var state int
	//online|offline|busy
	if hook.DeviceStatus=="online"{
		state = 1
	}
	if hook.DeviceStatus=="offline"{
		state = 0
	}
	if hook.DeviceStatus=="busy"{
		state = 2
	}

	Consumer[hook.Did] = `DeviceCloudStatuses_Cloud1{Device="` + strconv.Itoa(hook.Did) + `"} ` + strconv.Itoa(state) + "\n"

	w.Header().Set("Content-Type", "application/json")

	var buf []byte
	req.Body.Read(buf)

	fmt.Fprintf(w, `{"Content-type": "`+req.Header.Get("Content-Type")+`",`)
	fmt.Fprintf(w, `"Length": "`+req.Header.Get("Content-length")+`",`)
	fmt.Fprintf(w, `"success": true}`)

}

func SendPusher(metric *map[int]string) {

	for {
		s := ""

		for _, v := range *metric {
			s += v
		}

		//fmt.Println(s)

		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://3.22.234.194/metrics/job/device_statuses", bytes.NewReader([]byte(s)))
		if err != nil {
			log.Fatalln(err)
		}
		req.SetBasicAuth("pusher", "rtU2ssvx@")
		req.Header.Add("Content-type", "text/plain")
		res, _ := client.Do(req)
		res.Body.Close()

		time.Sleep(5*time.Second)
	}
}

func hnd(w http.ResponseWriter, req *http.Request) {
	for _, v:=range Consumer{
		fmt.Fprintf(w, v)
	}
}

func pingAlive()  {
	for {
		components.SendPusher(`DeviceCloudPipedream_alive 1` + "\n")
		time.Sleep(time.Second*5)
	}
}

func main() {



	Consumer = make(map[int]string)

	go pingAlive()

	go SendPusher(&Consumer)

	http.HandleFunc("/getDeviceHealth", hnd)
	http.HandleFunc("/hook", hookCatch)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}


