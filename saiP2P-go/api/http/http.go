package api

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/saiset-co/saiP2P-go/core"
	"go.uber.org/zap"
)

type Router struct {
	Core *core.Core
}

func New(core *core.Core) *Router {
	return &Router{
		Core: core,
	}
}

// Routes - http routes for the app
func (r *Router) Routes() {
	http.HandleFunc("/send", r.SendApi)   // send msgs by http
	http.HandleFunc("/stats", r.GetStats) //get stats
	http.HandleFunc("/test_big_msg", r.TestBigMsg)
}

// GetStats - return app stats for a debug purposes
func (r *Router) GetStats(resp http.ResponseWriter, _ *http.Request) {
	stats := r.Core.GetStats()
	data, err := json.Marshal(stats)
	if err != nil {
		err := map[string]interface{}{"Status": "NOK", "Error": err.Error()}
		errBody, _ := json.Marshal(err)
		log.Println(err)
		resp.Write(errBody)
		return
	}
	resp.Write(data)
}

// SendApi - api for accept messages from the http
func (r *Router) SendApi(resp http.ResponseWriter, req *http.Request) {
	message, err := io.ReadAll(req.Body)
	if err != nil {
		err := map[string]interface{}{"Status": "NOK", "Error": err.Error()}
		errBody, _ := json.Marshal(err)
		log.Println(err)
		resp.Write(errBody)
		return
	}

	to := req.URL.Query().Get("to")

	recipients := []string{}

	if len(to) > 1 {
		recipients = strings.Split(to, ",")
	} else if len(to) == 1 {
		recipients = append(recipients, to)
	}

	if err != nil {
		err := map[string]interface{}{"Status": "NOK", "Error": err.Error()}
		errBody, _ := json.Marshal(err)
		log.Println(err)
		resp.Write(errBody)
		return
	}

	address := r.Core.GetRealAddress()

	err = r.Core.SendMsg(message, recipients, address)
	if err != nil {
		err := map[string]interface{}{"Status": "NOK", "Error": err.Error()}
		errBody, _ := json.Marshal(err)
		log.Println(err)
		resp.Write(errBody)
		return
	}
	r.Core.Logger.Debug("Send", zap.Strings("recepients", recipients))

	response := map[string]interface{}{"Status": "OK"}
	responseBody, _ := json.Marshal(response)
	resp.Write(responseBody)
}

// TestBigMsg - handler for testing big msgs,sending via p2p
func (r *Router) TestBigMsg(resp http.ResponseWriter, req *http.Request) {
	sizeStr := req.URL.Query().Get("size")

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		resp.Write([]byte(err.Error()))
		return
	}

	recipients := []string{}

	address := r.Core.GetRealAddress()

	message := make([]byte, size)
	rand.Read(message) // set random values

	err = r.Core.SendMsg(message, recipients, address)
	if err != nil {
		err := map[string]interface{}{"Status": "NOK", "Error": err.Error()}
		errBody, _ := json.Marshal(err)
		log.Println(err)
		resp.Write(errBody)
		return
	}
	r.Core.Logger.Debug("Send", zap.Strings("recepients", recipients))

	response := map[string]interface{}{"Status": "OK"}
	responseBody, _ := json.Marshal(response)
	resp.Write(responseBody)
}
