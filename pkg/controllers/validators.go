package controllers

import (
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Validate(review *v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	request := review.Request

	response := &v1beta1.AdmissionResponse{}

	if request.Namespace == "default" {
		response.Allowed = false
		response.Result = &metav1.Status{Message: "Default namespace is not authorized"}
	}

	return response, nil
}
