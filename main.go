package main

import (
	"encoding/json"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	configFile       = kingpin.Flag("config.file", "RocketChat configuration file.").Default("config/rocketchat.yml").String()
	listenAddress    = kingpin.Flag("listen.address", "The address to listen on for HTTP requests.").Default(":9876").String()
	rocketChatClient RocketChatClient
)

// Webhook http response
type JSONResponse struct {
	Status  int
	Message string
}

func webhook(w http.ResponseWriter, r *http.Request) {

	data, err := readRequestBody(r)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Format notifications and send it
	SendNotification(rocketChatClient, data)

	// Returns a 200 if everything went smoothly
	sendJSONResponse(w, http.StatusOK, "Success")
}

// Starts 2 listeners
// - first one to give a status on the receiver itself
// - second one to actually process the data
func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	config := loadConfig(*configFile)

	rocketChatClient = GetRocketChatAuthenticatedClient(config)

	http.HandleFunc("/webhook", webhook)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("listening on: %v", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func sendJSONResponse(w http.ResponseWriter, status int, message string) {
	data := JSONResponse{
		Status:  status,
		Message: message,
	}
	bytes, _ := json.Marshal(data)

	w.WriteHeader(status)
	w.Write(bytes)
}

func readRequestBody(r *http.Request) (template.Data, error) {

	// Do not forget to close the body at the end
	defer r.Body.Close()

	// Extract data from the body in the Data template provided by AlertManager
	data := template.Data{}
	err := json.NewDecoder(r.Body).Decode(&data)

	return data, err
}

func loadConfig(configFile string) Config {
	config := Config{}

	// Load the config from the file
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		log.Fatalf("Error: %v", errYAML)
	}

	return config

}
