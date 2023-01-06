package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/gmalbrand/http-tools/pkg/utils"
	log "github.com/sirupsen/logrus"
)

var (
	caCertFile     = flag.String("ca-cert", "", "CA Certificate file in pem format")
	serverCertFile = flag.String("server-cert", "", "Server cert file in pem format")
	serverKeyFile  = flag.String("server-key", "", "Server key file in pem format")
)

type DeploymentConfig struct {
	Namespace   string
	Service     string
	Application string
	ServerCert  string
	ServerKey   string
	CACert      string
	Port        int
}

func fileToBase64(filepath string) (string, error) {
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		return "", fmt.Errorf("Open file %s fails with error : %s", filepath, err.Error())
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func initFlag(config *DeploymentConfig) {
	flag.StringVar(&config.Namespace, "namespace", "admin", "Kubernetes namespace")
	flag.StringVar(&config.Service, "service", "controller", "Kubernetes namespace")
	flag.StringVar(&config.Application, "application", "admission-controller", "Kubernetes namespace")
	flag.IntVar(&config.Port, "container-port", 8080, "Container port (if you change this parameter, update the dockerfile accordingly")
}

func main() {
	var err error
	utils.InitLog()
	log.Info("Generating deployment manifest from template")

	config := DeploymentConfig{Application: "admission-controller"}
	initFlag(&config)
	flag.Parse()

	if *serverCertFile == "" || *serverKeyFile == "" {
		log.Fatal("You must provide server certificate (-server-cert) and server key (-server-key))")
	}

	if *caCertFile == "" {
		log.Fatal("You must provide CA certificate (-ca-cert)")
	}

	if config.ServerCert, err = fileToBase64(*serverCertFile); err != nil {
		log.Fatal(err.Error())
	}

	if config.ServerKey, err = fileToBase64(*serverCertFile); err != nil {
		log.Fatal(err.Error())
	}

	if config.CACert, err = fileToBase64(*caCertFile); err != nil {
		log.Fatal(err.Error())
	}

	template, err := template.ParseFiles("configs/deployment.yaml")

	if err != nil {
		log.Fatal(err.Error())
	}

	if err := template.Execute(os.Stdout, config); err != nil {
		log.Fatal(err.Error())
	}
}
