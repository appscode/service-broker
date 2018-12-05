package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/pkg/eventer"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

var defaultDBPort = core.ServicePort{
	Name:       "db",
	Protocol:   core.ProtocolTCP,
	Port:       6379,
	TargetPort: intstr.FromString("db"),
}

func (c *Controller) ensureService(redis *api.Redis) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(redis, redis.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(redis)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(redis *api.Redis, serviceName string) error {
	service, err := c.Client.CoreV1().Services(redis.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindRedis ||
		service.Labels[api.LabelDatabaseName] != redis.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, redis.Namespace, serviceName)
	}

	return nil
}

func (c *Controller) createService(redis *api.Redis) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      redis.OffshootName(),
		Namespace: redis.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = redis.OffshootSelectors()
		in.Annotations = redis.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = redis.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			redis.Spec.ServiceTemplate.Spec.Ports,
		)

		if redis.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = redis.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if redis.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = redis.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = redis.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = redis.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = redis.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = redis.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if redis.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = redis.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	return ok, err
}

func (c *Controller) ensureStatsService(redis *api.Redis) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if redis.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if stats Service name exists
	if err := c.checkService(redis, redis.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, redis)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      redis.StatsService().ServiceName(),
		Namespace: redis.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = redis.OffshootLabels()
		in.Spec.Selector = redis.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       redis.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
