package test

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
)

const (
	Namespace = "default"
)

type K8sResource[T any] struct {
	Resource T
}

func int32PTR(i int32) *int32 { return &i }

func NewFakeClient() kubernetes.Interface {
	return fake.NewSimpleClientset()
}

func CreateNamespace(c kubernetes.Interface) *corev1.Namespace {
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
	}
	ns, _ := c.CoreV1().Namespaces().Create(context.TODO(), nsObj, metav1.CreateOptions{})
	return ns
}

func CreateDeployment(c kubernetes.Interface) *appsv1.Deployment {
	depObj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "nginx-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nginx-pvc",
								},
							},
						},
					},
				},
			},
			Replicas: int32PTR(1),
		},
	}
	dep, _ := c.AppsV1().Deployments(Namespace).Create(context.TODO(), depObj, metav1.CreateOptions{})
	return dep
}

func CreateStatefulSet(c kubernetes.Interface) *appsv1.StatefulSet {
	scName := "standard"
	stsObj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-statefulset",
			Namespace: Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nginx-pvc",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName:       "nginx-pvc",
						StorageClassName: &scName,
					},
				},
			},
		},
	}
	sts, _ := c.AppsV1().StatefulSets(Namespace).Create(context.Background(), stsObj, metav1.CreateOptions{})
	return sts
}

func CreateDaemonSet(c kubernetes.Interface) *appsv1.DaemonSet {
	dsObj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-daemonset",
			Namespace: Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
	ds, _ := c.AppsV1().DaemonSets(Namespace).Create(context.TODO(), dsObj, metav1.CreateOptions{})
	return ds
}

func CreatePVC(c kubernetes.Interface, name, namespace, storageClass string) *corev1.PersistentVolumeClaim {
	pvcObj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
		},
	}
	pvc, _ := c.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), pvcObj, metav1.CreateOptions{})
	return pvc
}
