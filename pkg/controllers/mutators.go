package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
)

type Operation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

type Patch []Operation

func (p Patch) Encode() ([]byte, error) {
	raw, err := json.Marshal(p)
	if err != nil {
		msg := fmt.Sprintf("Failed to marshal patch, error : %s", err.Error())
		log.Error(msg)
		return nil, errors.New(msg)
	}

	return raw, nil
}

var (
	externalIPPatch = Patch{Operation{Op: "add", Path: "/spec/loadBalancerSourceRanges", Value: []string{"82.66.142.85/32", "192.168.1.0/24"}}}
)

func Mutate(review *v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {

	if review.Request.Kind.Kind == "Service" {
		return mutateService(review.Request.Object.Raw)
	}

	response := &v1beta1.AdmissionResponse{
		Allowed: true,
		Result:  &metav1.Status{Message: "Nothing to do here"},
	}

	return response, nil
}

func mutateService(raw []byte) (*v1beta1.AdmissionResponse, error) {
	service := &v1.Service{}

	if err := json.Unmarshal(raw, service); err != nil {
		msg := fmt.Sprintf("Failed to unmarshal service, error %s", err.Error())
		log.Error(msg)
		return nil, errors.New(msg)
	}

	if service.Spec.Type == v1.ServiceTypeLoadBalancer {
		patch, err := externalIPPatch.Encode()
		if err != nil {
			msg := fmt.Sprintf("Failed to generate patch, error : %s", err.Error())
			log.Error(msg)
			return nil, errors.New(msg)
		}

		response := &v1beta1.AdmissionResponse{
			Allowed: true,
			Patch:   patch,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
			Result: &metav1.Status{Message: "Add source ranges to load balancer"},
		}
		return response, nil
	}

	response := &v1beta1.AdmissionResponse{
		Allowed: true,
		Result:  &metav1.Status{Message: "No mutation required"},
	}

	return response, nil
}
