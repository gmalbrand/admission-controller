package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"

	"github.com/gmalbrand/admission-controller/pkg/controllers"
	"github.com/gmalbrand/http-tools/pkg/logger"
	"github.com/gmalbrand/http-tools/pkg/monitoring"
	"github.com/gmalbrand/http-tools/pkg/utils"
)

var (
	certFile      = flag.String("cert", "/etc/certs/server.pem", "Path to TLS certificate")
	certKeyFile   = flag.String("key", "/etc/certs/key.pem", "Path to TLS certificate key")
	listeningPort = flag.Int("port", 8080, "Server listening port")
)

func extractAdmissionReview(request *http.Request) (*v1beta1.AdmissionReview, error) {
	var body []byte
	if request.Body != nil {
		if data, err := ioutil.ReadAll(request.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		log.Error("Failed to process request: empty payload")
		return nil, errors.New("Failed to process request: empty payload")
	}

	admissionReview := &v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, admissionReview); err != nil {
		msg := fmt.Sprintf("Failed to unmarshal request body, error : %s", err.Error())
		log.Error(msg)
		return nil, errors.New(msg)
	}

	return admissionReview, nil
}

func wrapResponse(review *v1beta1.AdmissionReview, response *v1beta1.AdmissionResponse) *v1beta1.AdmissionReview {
	response.UID = review.Request.UID
	result := &v1beta1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: response,
	}
	return result
}

func writeAdmissionReview(w http.ResponseWriter, review *v1beta1.AdmissionReview) error {
	resp, err := json.Marshal(review)

	if err != nil {
		msg := fmt.Sprintf("Failed to marshal response, error : %s", err.Error())
		log.Error(msg)
		return errors.New(msg)
	}

	if _, err := w.Write(resp); err != nil {
		msg := fmt.Sprintf("Failed to write response, error : %s", err.Error())
		log.Error(msg)
		return errors.New(msg)
	}
	return nil
}

func validate(w http.ResponseWriter, req *http.Request) {

	review, err := extractAdmissionReview(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := controllers.Validate(review)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeAdmissionReview(w, wrapResponse(review, response))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func mutate(w http.ResponseWriter, req *http.Request) {
	review, err := extractAdmissionReview(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Infof("Mutating request : %s", review)
	response, err := controllers.Mutate(review)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = writeAdmissionReview(w, wrapResponse(review, response))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := monitoring.NewMonitoredMux()
	utils.InitLog()
	flag.Parse()

	if *certFile == "" || *certKeyFile == "" {
		log.Fatal("You must provide a certificate and its associated key")
	}

	mux.HandleFunc("/validate", validate)
	mux.HandleFunc("/mutate", mutate)
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })

	err := http.ListenAndServeTLS(fmt.Sprintf(":%d", *listeningPort), *certFile, *certKeyFile, logger.AccessCombinedLog(mux.Server()))

	if err != nil {
		log.Fatal(err.Error())
	}
}
