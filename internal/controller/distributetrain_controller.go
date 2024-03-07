/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/cp"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	devv1 "github.com/labring/sealos/controllers/opt/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DistributeTrainReconciler reconciles a DistributeTrain object
type DistributeTrainReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dev.opt.sealos.io,resources=distributetrains,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dev.opt.sealos.io,resources=distributetrains/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dev.opt.sealos.io,resources=distributetrains/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods;pods/exec,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DistributeTrain object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *DistributeTrainReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	// Create an instance of the CR type to hold the retrieved object
	var dt devv1.DistributeTrain

	// Retrieve the CR from the Kubernetes API server
	if err := r.Get(ctx, req.NamespacedName, &dt); err != nil {
		// Handle error, log, and return
		fmt.Printf("Error getting Job %s: %v\n", req.NamespacedName, err)
		return ctrl.Result{}, err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Error creating in-cluster config: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < dt.Spec.Size; i++ {
		fmt.Println(i)
		if i == 0 {
			// Create a Job object based on the CR specifications
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      req.Name + "-job" + strconv.Itoa(i), // Use a unique name for the Job
					Namespace: dt.Spec.ServiceAccountName,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							// Define the container and restart policy
							Containers: []corev1.Container{
								{
									Name:  "distribute-train-container",
									Image: dt.Spec.Image,
									Env: []corev1.EnvVar{
										{Name: dt.Spec.Env[0].Name, Value: dt.Spec.Env[0].Value},
										{Name: dt.Spec.Env[1].Name, Value: dt.Spec.Env[1].Value},
										// Add other environment variables as needed
									},
									Command: []string{
										"/bin/sh",
										"-c",
										"echo ''Distribute training mission begins...'' &&" +
											"while [ ! -e ''/home/ipok.json'' ]; do sleep 5; done &&" +
											"export NODE_IPS=`cat /home/hostfile.json |paste -d '','' -s`  &&" +
											dt.Spec.MasterCmd +
											" --node_ips=''$NODE_IPS''  --num_nodes=" + dt.Spec.Env[3].Value +
											" --data_url=" + dt.Spec.VolumeMounts[0].MountPath +
											" --train_model_out=" + dt.Spec.VolumeMounts[1].MountPath + "/model-out" +
											" --gpu_num_per_node=" + dt.Spec.Env[6].Value +
											"&& echo ''Distribute training mission is over''", // Use masterCmd from DistributeTrain spec
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse(dt.Spec.MasterResources.Limits.CPU),
											corev1.ResourceMemory: resource.MustParse(dt.Spec.MasterResources.Limits.Memory),
											"nvidia.com/gpu":      resource.MustParse(dt.Spec.MasterResources.Limits.NvidiaGPU),
										},
										// 可选，如果需要设置 requests，可以使用下面的行
										Requests: corev1.ResourceList{
											// 设置 requests 的资源限制
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{MountPath: dt.Spec.VolumeMounts[0].MountPath, Name: dt.Spec.VolumeMounts[0].Name, ReadOnly: dt.Spec.VolumeMounts[0].ReadOnly},
										{MountPath: dt.Spec.VolumeMounts[1].MountPath, Name: dt.Spec.VolumeMounts[1].Name, ReadOnly: dt.Spec.VolumeMounts[1].ReadOnly},
										{MountPath: dt.Spec.VolumeMounts[2].MountPath, Name: dt.Spec.VolumeMounts[2].Name, ReadOnly: dt.Spec.VolumeMounts[2].ReadOnly},
										// Add other volume mounts as needed
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure, // Or use "Never" if appropriate
							Volumes: []corev1.Volume{
								{
									Name: "volume-0",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: dt.Spec.Volumes[0].HostPath.Path,
											Type: new(corev1.HostPathType),
										},
									},
								},
								{
									Name: "volume-1",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: dt.Spec.Volumes[1].HostPath.Path,
											Type: new(corev1.HostPathType),
										},
									},
								},
								{
									Name: "volume-2",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: dt.Spec.Volumes[2].HostPath.Path,
											Type: new(corev1.HostPathType),
										},
									},
								},
							},
							// Add other pod specifications as needed
						},
					},
					// Add other Job specifications as needed
				},
			}
			// Set the owner reference to link the Job to the DistributeTrain CR
			if err := ctrl.SetControllerReference(&dt, job, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}

			// Submit the Job to the Kubernetes API server
			if err := r.Create(ctx, job); err != nil {
				fmt.Printf("Error creating Job for DistributeTrain %s: %v\n", req.NamespacedName, err)
				return ctrl.Result{}, err
			}

		} else {
			// Create a Job object based on the CR specifications
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      req.Name + "-job" + strconv.Itoa(i), // Use a unique name for the Job
					Namespace: dt.Spec.ServiceAccountName,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							// Define the container and restart policy
							Containers: []corev1.Container{
								{
									Name:  "distribute-train-container",
									Image: dt.Spec.Image,
									Env: []corev1.EnvVar{
										{Name: dt.Spec.Env[0].Name, Value: dt.Spec.Env[0].Value},
										{Name: dt.Spec.Env[1].Name, Value: dt.Spec.Env[1].Value},
										// Add other environment variables as needed
									},
									Command: []string{
										"/bin/sh",
										"-c",
										"echo ''Distribute training mission begins...'' &&" +
											"while [ ! -e ''/home/ipok.json'' ]; do sleep 5; done &&" +
											"export NODE_IPS=`cat /home/hostfile.json |paste -d '','' -s`  &&" +
											dt.Spec.SlaveCmd +
											" --node_ips=''$NODE_IPS''  --num_nodes=" + dt.Spec.Env[3].Value +
											" --data_url=" + dt.Spec.VolumeMounts[0].MountPath +
											" --train_model_out=" + dt.Spec.VolumeMounts[1].MountPath + "/model-out" +
											" --gpu_num_per_node=" + dt.Spec.Env[6].Value +
											"&& echo ''Distribute training mission is over''", // Use slaveCmd from DistributeTrain spec
									},
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse(dt.Spec.SlaveResources.Limits.CPU),
											corev1.ResourceMemory: resource.MustParse(dt.Spec.SlaveResources.Limits.Memory),
											"nvidia.com/gpu":      resource.MustParse(dt.Spec.SlaveResources.Limits.NvidiaGPU),
										},
										// 可选，如果需要设置 requests，可以使用下面的行
										Requests: corev1.ResourceList{
											// 设置 requests 的资源限制
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{MountPath: dt.Spec.VolumeMounts[0].MountPath, Name: dt.Spec.VolumeMounts[0].Name, ReadOnly: dt.Spec.VolumeMounts[0].ReadOnly},
										{MountPath: dt.Spec.VolumeMounts[1].MountPath, Name: dt.Spec.VolumeMounts[1].Name, ReadOnly: dt.Spec.VolumeMounts[1].ReadOnly},
										{MountPath: dt.Spec.VolumeMounts[2].MountPath, Name: dt.Spec.VolumeMounts[2].Name, ReadOnly: dt.Spec.VolumeMounts[2].ReadOnly},
										// Add other volume mounts as needed
									},
								},
							},
							RestartPolicy: corev1.RestartPolicyOnFailure, // Or use "Never" if appropriate
							Volumes: []corev1.Volume{
								{
									Name: "volume-0",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: dt.Spec.Volumes[0].HostPath.Path,
											Type: new(corev1.HostPathType),
										},
									},
								},
								{
									Name: "volume-1",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: dt.Spec.Volumes[1].HostPath.Path,
											Type: new(corev1.HostPathType),
										},
									},
								},
								{
									Name: "volume-2",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: dt.Spec.Volumes[2].HostPath.Path,
											Type: new(corev1.HostPathType),
										},
									},
								},
							},
							// Add other pod specifications as needed
						},
					},
					// Add other Job specifications as needed
				},
			}
			// Set the owner reference to link the Job to the DistributeTrain CR
			if err := ctrl.SetControllerReference(&dt, job, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}

			// Submit the Job to the Kubernetes API server
			if err := r.Create(ctx, job); err != nil {
				fmt.Printf("Error creating Job for DistributeTrain %s: %v\n", req.NamespacedName, err)
				return ctrl.Result{}, err
			}
		}
	}

	var podIPs []string
	var podNames []string

	time.Sleep(5 * time.Second)

	// 循环等待每个 Pod 的 IP 地址不为空
	for {
		podIPs = nil
		podNames = nil

		for i := 0; i < dt.Spec.Size; i++ {
			// 获取由 Job 创建的 Pod 列表
			podList := &corev1.PodList{}
			labelSelector := labels.SelectorFromSet(map[string]string{"job-name": req.Name + "-job" + strconv.Itoa(i)})
			listOptions := &client.ListOptions{LabelSelector: labelSelector}
			if err := r.List(ctx, podList, listOptions); err != nil {
				fmt.Printf("Error listing Pods for Job %s: %v\n", req.Name+"-job"+strconv.Itoa(i), err)
				return ctrl.Result{}, err
			}

			// 遍历 Pod 列表，获取每个 Pod 的 IP 地址
			for _, pod := range podList.Items {
				podIP := pod.Status.PodIP
				podName := pod.Name
				if podIP != "" {
					fmt.Printf("Pod IP for Job %s: %s\n", req.Name+"-job"+strconv.Itoa(i), podIP)
				}

				podIPs = append(podIPs, podIP)
				podNames = append(podNames, podName)
			}
		}

		// 检查是否所有的 Pod IP 地址都不为空
		allIPsNotEmpty := true
		for _, ip := range podIPs {
			if ip == "" {
				allIPsNotEmpty = false
				break
			}
		}

		// 如果所有的 Pod IP 地址都不为空，跳出循环
		if allIPsNotEmpty {
			break
		}

		time.Sleep(30 * time.Second)
	}

	filePath := "/home/temp.json"
	filePath_ok := "/home/ipok.json"

	// 尝试创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return ctrl.Result{}, nil
	}
	defer file.Close()

	for i := 0; i < dt.Spec.Size; i++ {
		// 将数据写入文件
		_, err = file.WriteString(podIPs[i] + "\n")
		if err != nil {
			return ctrl.Result{}, nil
		}
	}

	file_ok, err := os.Create(filePath_ok)
	if err != nil {
		return ctrl.Result{}, nil
	}
	defer file_ok.Close()

	for i := 0; i < dt.Spec.Size; i++ {
		err = r.copyFromPod(config, dt.Spec.ServiceAccountName, podNames[i], filePath, "/home/hostfile.json")
		if err != nil {
			panic(err.Error())
		}
	}

	for i := 0; i < dt.Spec.Size; i++ {
		err = r.copyFromPod(config, dt.Spec.ServiceAccountName, podNames[i], filePath_ok, "/home/ipok.json")
		if err != nil {
			panic(err.Error())
		}
	}

	// Your logic using the CR information
	fmt.Printf("Reconciling DistributeTrain %s\n", req.NamespacedName)

	return ctrl.Result{}, nil
}

func (d *DistributeTrainReconciler) copyFromPod(restConfig *rest.Config, namespace, pod, srcDir, dstDir string) error {
	restConfig.APIPath = "/api"
	restConfig.GroupVersion = &schema.GroupVersion{Version: "v1"} // this targets the core api groups so the url path will be /api/v1
	restConfig.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()

	copyOptions := cp.NewCopyOptions(ioStreams)
	configFlags := genericclioptions.NewConfigFlags(false)
	f := cmdutil.NewFactory(configFlags)
	cmd := cp.NewCmdCp(f, ioStreams)
	err := copyOptions.Complete(f, cmd, []string{srcDir, pod + ":" + dstDir})
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	copyOptions.Clientset = clientSet
	copyOptions.ClientConfig = restConfig
	copyOptions.Namespace = namespace

	err = copyOptions.Run()
	if err != nil {
		return fmt.Errorf("could not run copy operation: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DistributeTrainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&devv1.DistributeTrain{}).
		Complete(r)
}
