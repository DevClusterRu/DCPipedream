package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var Consumer map[int]int  //Key: DID, value: state

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
	if hook.DeviceStatus=="offline"{
		state = 1
	}
	if hook.DeviceStatus=="online"{
		state = 2
	}
	if hook.DeviceStatus=="busy"{
		state = 3
	}

	//Consumer[hook.Did] = `DeviceCloudStatuses_Cloud1{Device="` + strconv.Itoa(hook.Did) + `"} ` + strconv.Itoa(state) + "\n"
	Consumer[hook.Did] = state

	w.Header().Set("Content-Type", "application/json")

	var buf []byte
	req.Body.Read(buf)

	fmt.Fprintf(w, `{"Content-type": "`+req.Header.Get("Content-Type")+`",`)
	fmt.Fprintf(w, `"Length": "`+req.Header.Get("Content-length")+`",`)
	fmt.Fprintf(w, `"success": true}`)

}

func SendPusher(metric map[int]int) {

	for {
		s := ""
		for k, v := range metric {
			s += `DeviceCloudStatuses_AllClouds{Device="` + strconv.Itoa(k) + `"} ` + strconv.Itoa(v) + "\n"
			if Consumer[k] < 30 {
				Consumer[k] += 10
			}
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
	for k, v:=range Consumer{
		fmt.Fprintf(w, `DeviceCloudStatuses_AllClouds{Device="` + strconv.Itoa(k) + `"} ` + strconv.Itoa(v) + "\n")
	}
}

func pingAlive()  {
	aliver:=1
	for {
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://3.22.234.194/metrics/job/device_statuses", bytes.NewReader([]byte(`DeviceCloudPipedream_alive `+strconv.Itoa(aliver) + "\n")))
		if err != nil {
			log.Println(err)
			continue
		}
		req.SetBasicAuth("pusher", "rtU2ssvx@")
		req.Header.Add("Content-type", "text/plain")
		res, _ := client.Do(req)
		res.Body.Close()

		if aliver==1 {
			aliver = 2
		} else {
			aliver = 1
		}
		time.Sleep(time.Second*5)
	}
}




func main() {



	Consumer = make(map[int]int)

	go pingAlive()

	go SendPusher(Consumer)

	http.HandleFunc("/getDeviceHealth", hnd)
	http.HandleFunc("/hook", hookCatch)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}


