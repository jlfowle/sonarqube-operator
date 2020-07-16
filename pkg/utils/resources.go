package utils

import (
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServicePorts(serverType sonarsourcev1alpha1.ServerType) []corev1.ServicePort {
	var servicePorts []corev1.ServicePort

	switch serverType {
	case sonarsourcev1alpha1.AIO, "":
		servicePorts = []corev1.ServicePort{
			{
				Name:     "web",
				Protocol: corev1.ProtocolTCP,
				Port:     sonarsourcev1alpha1.ApplicationWebPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: sonarsourcev1alpha1.ApplicationWebPort,
					StrVal: "",
				},
			},
		}
	case sonarsourcev1alpha1.Application:
		servicePorts = []corev1.ServicePort{
			{
				Name:     "web",
				Protocol: corev1.ProtocolTCP,
				Port:     sonarsourcev1alpha1.ApplicationWebPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: sonarsourcev1alpha1.ApplicationWebPort,
					StrVal: "",
				},
			},
			{
				Name:     "ce",
				Protocol: corev1.ProtocolTCP,
				Port:     sonarsourcev1alpha1.ApplicationCEPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: sonarsourcev1alpha1.ApplicationCEPort,
					StrVal: "",
				},
			},
			{
				Name:     "node",
				Protocol: corev1.ProtocolTCP,
				Port:     sonarsourcev1alpha1.ApplicationPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: sonarsourcev1alpha1.ApplicationPort,
					StrVal: "",
				},
			},
		}
	case sonarsourcev1alpha1.Search:
		servicePorts = []corev1.ServicePort{
			{
				Name:     "search",
				Protocol: corev1.ProtocolTCP,
				Port:     sonarsourcev1alpha1.SearchPort,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: sonarsourcev1alpha1.SearchPort,
					StrVal: "",
				},
			},
		}
	}

	return servicePorts
}

func VerifyService(client client.Client, service1, service2 *corev1.Service) error {
	if !reflect.DeepEqual(service2.Spec.Selector, service1.Spec.Selector) {
		service1.Spec.Selector = service2.Spec.Selector
		return UpdateResource(client, service1, ErrorReasonResourceUpdate, "updated service selector")
	}

	if !reflect.DeepEqual(service2.Spec.Ports, service1.Spec.Ports) {
		service1.Spec.Ports = service2.Spec.Ports
		return UpdateResource(client, service1, ErrorReasonResourceUpdate, "updated service ports")
	}

	if !reflect.DeepEqual(service2.Spec.Type, service1.Spec.Type) {
		service1.Spec.Type = service2.Spec.Type
		return UpdateResource(client, service1, ErrorReasonResourceUpdate, "updated service type")
	}

	if !reflect.DeepEqual(service2.Labels, service1.Labels) {
		service1.Labels = service2.Labels
		return UpdateResource(client, service1, ErrorReasonResourceUpdate, "updated service labels")
	}

	return nil
}
