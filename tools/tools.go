package tools

import (
	"github.com/asaskevich/govalidator"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckPodIpv4IPReady(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, v := range pod.Status.PodIPs {
		if govalidator.IsIPv4(v.IP) {
			return true
		}
	}
	return false
}

func CheckPodIpv6IPReady(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, v := range pod.Status.PodIPs {
		if govalidator.IsIPv6(v.IP) {
			return true
		}
	}
	return false
}
