// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package kpis

import (
	"github.com/onosproject/onos-lib-go/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"

	hrm_status "github.com/open-edge-platform/infra-managers/host/pkg/status"
	mm_status "github.com/open-edge-platform/infra-managers/maintenance/pkg/status"
	om_status "github.com/open-edge-platform/infra-onboarding/onboarding-manager/pkg/status"
)

// Consts define the metrics labels.
const (
	metricsPrefix = "edge"

	hostID                    = "hostID"
	hostDeviceGUID            = "deviceGuid"
	name                      = "name"
	hostName                  = "hostname"
	hostSerial                = "serial"
	hostTenant                = "projectId"
	hostStatus                = "status"
	hostOnboardingStatus      = "onboardingStatus"
	hostProvisioningStatus    = "provisioningStatus"
	hostUpdateStatus          = "updateStatus"
	hostTotalProvisioningTime = "totalProvisioningTime"
)

// Var definitions of inventory host status metrics inventoryBuilder and static labels.
// builder is used to create metrics in the PrometheusFormat.
var (
	staticLabelsInventory = map[string]string{}
	inventoryBuilder      = prom.NewBuilder(metricsPrefix, "", staticLabelsInventory)
)

// HostStatus defines the common data that can be used
// to output the format of a KPI (e.g., PrometheusFormat).
type HostStatus struct {
	HostID             string
	DeviceGUID         string
	Name               string
	Hostname           string
	Serial             string
	TenantID           string
	Status             string
	UpdateStatus       string
	OnboardingStatus   string
	ProvisioningStatus string
	HasSchedule        bool
}

// HostProvisioningTime defines the common data that can be used
// to output the format of a KPI (e.g., PrometheusFormat).
type HostProvisioningTime struct {
	HostID                string
	DeviceGUID            string
	Name                  string
	Hostname              string
	Serial                string
	TenantID              string
	TotalProvisioningTime float64
}

// InventoryHostsStatus stores each data structure for a host status
// which contains the annotations as defined by HostStatus struct.
type InventoryHostsStatus struct {
	name        string
	description string
	Labels      []string
	LabelValues []string
	// Status is indexed by the hostID unique value.
	Status map[string]HostStatus
}

// InventoryHostsSchedule stores each data structure for a host maintenance
// which contains the annotations as defined by HostStatus struct.
type InventoryHostsSchedule struct {
	name        string
	description string
	Labels      []string
	LabelValues []string
	// Status is indexed by the hostID unique value.
	Status map[string]HostStatus
}

// InventoryHostsProvisioningStatus stores each data structure for a host status
// which contains the annotations as defined by ProvisioningStatus struct.
type InventoryHostsProvisioningStatus struct {
	name        string
	description string
	Labels      []string
	LabelValues []string
	// Status is indexed by the hostID unique value.
	Status map[string]HostStatus
}

// InventoryHostsOnboardingStatus stores each data structure for a host status
// which contains the annotations as defined by OnboardingStatus struct.
type InventoryHostsOnboardingStatus struct {
	name        string
	description string
	Labels      []string
	LabelValues []string
	// Status is indexed by the hostID unique value.
	Status map[string]HostStatus
}

// InventoryHostsUpdateStatus stores each data structure for a host status
// which contains the annotations as defined by MaintenanceStatus struct.
type InventoryHostsUpdateStatus struct {
	name        string
	description string
	Labels      []string
	LabelValues []string
	// Status is indexed by the hostID unique value.
	Status map[string]HostStatus
}

// InventoryHostsTotalProvisioningTime stores each data structure for a host status
// which contains the annotations as defined by Total Provisioning Time struct.
type InventoryHostsTotalProvisoningTime struct {
	name        string
	description string
	Labels      []string
	LabelValues []string
	// Status is indexed by the hostID unique value.
	ProvisioningTime map[string]HostProvisioningTime
}

// PrometheusFormat implements the contract behavior of the kpis.KPI
// interface for InventoryHostsStatus.
// The metric value specifies that a host exists in the Inventory
// with the given labels.
// The metric can be filtered by the labels to get a unique or a sum of hosts in
// a specific status mode.
func (h *InventoryHostsStatus) PrometheusFormat() ([]prometheus.Metric, error) {
	metrics := make([]prometheus.Metric, 0, len(h.Status))

	h.Labels = []string{hostID, hostDeviceGUID, name, hostName, hostSerial, hostTenant, hostStatus}
	metricDesc := inventoryBuilder.NewMetricDesc(h.name, h.description, h.Labels, staticLabelsInventory)

	// For each one of the hosts indexed in the h.Status map, it adds a metric
	// of GaugeValue type, with value 1, and the labels associated with the host status.
	for _, hostStatus := range h.Status {
		status := hostStatus.Status
		statuses := []struct {
			v bool
			n string
		}{
			{status == hrm_status.HostStatusEmpty.Status, hrm_status.HostStatusEmpty.Status},
			{status == hrm_status.HostStatusUnknown.Status, hrm_status.HostStatusUnknown.Status},
			{status == hrm_status.HostStatusNoConnection.Status, hrm_status.HostStatusNoConnection.Status},
			{status == hrm_status.HostStatusRunning.Status, hrm_status.HostStatusRunning.Status},
			{status == hrm_status.HostStatusError.Status, hrm_status.HostStatusError.Status},
			{status == hrm_status.HostStatusInvalidating.Status, hrm_status.HostStatusInvalidating.Status},
			{status == hrm_status.HostStatusInvalidated.Status, hrm_status.HostStatusInvalidated.Status},
			{status == hrm_status.HostStatusDeleting.Status, hrm_status.HostStatusDeleting.Status},
		}

		// For each host, it adds one metric with gaugeVal 1 for the current status,
		// and it adds other metrics with gaugeVal 0 for all the other host statuses.
		for _, s := range statuses {
			gaugeVal := func(b bool) float64 {
				if b {
					return 1.0
				}
				return 0.0
			}(s.v)

			metric := inventoryBuilder.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				gaugeVal,
				hostStatus.HostID,
				hostStatus.DeviceGUID,
				hostStatus.Name,
				hostStatus.Hostname,
				hostStatus.Serial,
				hostStatus.TenantID,
				s.n,
			)
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

// PrometheusFormat implements the contract behavior of the kpis.KPI
// interface for InventoryHostsMaintenance.
// The metric value, named gaugeVal, specifies if the given metric with the provided
// labels equals to 1.0 if a host is in maintenance mode or 0.0 if not.
// The metric can be filtered by the labels to get a unique or a sum of hosts in
// maintenance mode.
func (h *InventoryHostsSchedule) PrometheusFormat() ([]prometheus.Metric, error) {
	metrics := make([]prometheus.Metric, 0, len(h.Status))

	h.Labels = []string{hostID, hostDeviceGUID, name, hostName, hostSerial, hostTenant}
	metricDesc := inventoryBuilder.NewMetricDesc(h.name, h.description, h.Labels, staticLabelsInventory)

	// For each one of the hosts indexed in the h.Status map, it adds a metric
	// of GaugeValue type, with value 1 in case a host is in maintenance mode or 0 otherwise,
	// and the labels associated with the host status.
	var gaugeVal float64
	for _, hostStatus := range h.Status {
		if hostStatus.HasSchedule {
			gaugeVal = 1.0
		} else {
			gaugeVal = 0.0
		}

		metric := inventoryBuilder.MustNewConstMetric(
			metricDesc,
			prometheus.GaugeValue,
			gaugeVal,
			hostStatus.HostID,
			hostStatus.DeviceGUID,
			hostStatus.Name,
			hostStatus.Hostname,
			hostStatus.Serial,
			hostStatus.TenantID,
		)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// PrometheusFormat implements the contract behavior of the kpis.KPI
// interface for InventoryHostsOnboardingStatus.
// The metric value specifies that a host exists in the Inventory
// with the given labels.
// The metric can be filtered by the labels to get a unique or a sum of hosts in
// a specific status mode.
//
//nolint:dupl // The code is duplicated to avoid the complexity of the function.
func (h *InventoryHostsOnboardingStatus) PrometheusFormat() ([]prometheus.Metric, error) {
	metrics := make([]prometheus.Metric, 0, len(h.Status))

	h.Labels = []string{hostID, hostDeviceGUID, name, hostName, hostSerial, hostTenant, hostOnboardingStatus}
	metricDesc := inventoryBuilder.NewMetricDesc(h.name, h.description, h.Labels, staticLabelsInventory)

	// For each one of the hosts indexed in the h.Status map, it adds a metric
	// of GaugeValue type, with value 1, and the labels associated with the host status.
	for _, hostStatus := range h.Status {
		onboardingStatus := hostStatus.OnboardingStatus
		statuses := []struct {
			v bool
			n string
		}{
			{onboardingStatus == om_status.OnboardingStatusDone.Status, om_status.OnboardingStatusDone.Status},
			{onboardingStatus == om_status.OnboardingStatusFailed.Status, om_status.OnboardingStatusFailed.Status},
			{onboardingStatus == om_status.InitializationDone.Status, om_status.InitializationDone.Status},
			{onboardingStatus == om_status.InitializationFailed.Status, om_status.InitializationFailed.Status},
		}

		// For each host, it adds one metric with gaugeVal 1 for the current status,
		// and it adds other metrics with gaugeVal 0 for all the other host statuses.
		for _, s := range statuses {
			gaugeVal := func(b bool) float64 {
				if b {
					return 1.0
				}
				return 0.0
			}(s.v)

			metric := inventoryBuilder.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				gaugeVal,
				hostStatus.HostID,
				hostStatus.DeviceGUID,
				hostStatus.Name,
				hostStatus.Hostname,
				hostStatus.Serial,
				hostStatus.TenantID,
				s.n,
			)
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

// PrometheusFormat implements the contract behavior of the kpis.KPI
// interface for InventoryHostsOnboardingStatus.
// The metric value specifies that a host exists in the Inventory
// with the given labels.
// The metric can be filtered by the labels to get a unique or a sum of hosts in
// a specific status mode.
//
//nolint:dupl // The code is duplicated to avoid the complexity of the function.
func (h *InventoryHostsProvisioningStatus) PrometheusFormat() ([]prometheus.Metric, error) {
	metrics := make([]prometheus.Metric, 0, len(h.Status))

	h.Labels = []string{hostID, hostDeviceGUID, name, hostName, hostSerial, hostTenant, hostProvisioningStatus}
	metricDesc := inventoryBuilder.NewMetricDesc(h.name, h.description, h.Labels, staticLabelsInventory)

	// For each one of the hosts indexed in the h.Status map, it adds a metric
	// of GaugeValue type, with value 1, and the labels associated with the host status.
	for _, hostStatus := range h.Status {
		provisioningStatus := hostStatus.ProvisioningStatus
		statuses := []struct {
			v bool
			n string
		}{
			{provisioningStatus == om_status.ProvisioningStatusUnknown.Status, om_status.ProvisioningStatusUnknown.Status},
			{provisioningStatus == om_status.ProvisioningStatusInProgress.Status, om_status.ProvisioningStatusInProgress.Status},
			{provisioningStatus == om_status.ProvisioningStatusFailed.Status, om_status.ProvisioningStatusFailed.Status},
			{provisioningStatus == om_status.ProvisioningStatusDone.Status, om_status.ProvisioningStatusDone.Status},
		}

		// For each host, it adds one metric with gaugeVal 1 for the current status,
		// and it adds other metrics with gaugeVal 0 for all the other host statuses.
		for _, s := range statuses {
			gaugeVal := func(b bool) float64 {
				if b {
					return 1.0
				}
				return 0.0
			}(s.v)

			metric := inventoryBuilder.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				gaugeVal,
				hostStatus.HostID,
				hostStatus.DeviceGUID,
				hostStatus.Name,
				hostStatus.Hostname,
				hostStatus.Serial,
				hostStatus.TenantID,
				s.n,
			)
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

// PrometheusFormat implements the contract behavior of the kpis.KPI
// interface for InventoryHostsOnboardingStatus.
// The metric value specifies that a host exists in the Inventory
// with the given labels.
// The metric can be filtered by the labels to get a unique or a sum of hosts in
// a specific status mode.
func (h *InventoryHostsUpdateStatus) PrometheusFormat() ([]prometheus.Metric, error) {
	metrics := make([]prometheus.Metric, 0, len(h.Status))

	h.Labels = []string{hostID, hostDeviceGUID, name, hostName, hostSerial, hostTenant, hostUpdateStatus}
	metricDesc := inventoryBuilder.NewMetricDesc(h.name, h.description, h.Labels, staticLabelsInventory)

	// For each one of the hosts indexed in the h.Status map, it adds a metric
	// of GaugeValue type, with value 1, and the labels associated with the host status.
	for _, hostStatus := range h.Status {
		updateStatus := hostStatus.UpdateStatus
		statuses := []struct {
			v bool
			n string
		}{
			{updateStatus == mm_status.UpdateStatusUnknown.Status, mm_status.UpdateStatusUnknown.Status},
			{updateStatus == mm_status.UpdateStatusInProgress.Status, mm_status.UpdateStatusInProgress.Status},
			{updateStatus == mm_status.UpdateStatusDone.Status, mm_status.UpdateStatusDone.Status},
			{updateStatus == mm_status.UpdateStatusFailed.Status, mm_status.UpdateStatusFailed.Status},
			{updateStatus == mm_status.UpdateStatusUpToDate.Status, mm_status.UpdateStatusUpToDate.Status},
		}

		// For each host, it adds one metric with gaugeVal 1 for the current status,
		// and it adds other metrics with gaugeVal 0 for all the other host statuses.
		for _, s := range statuses {
			gaugeVal := func(b bool) float64 {
				if b {
					return 1.0
				}
				return 0.0
			}(s.v)

			metric := inventoryBuilder.MustNewConstMetric(
				metricDesc,
				prometheus.GaugeValue,
				gaugeVal,
				hostStatus.HostID,
				hostStatus.DeviceGUID,
				hostStatus.Name,
				hostStatus.Hostname,
				hostStatus.Serial,
				hostStatus.TenantID,
				s.n,
			)
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

// PrometheusFormat implements the contract behavior of the kpis.KPI
// interface for InventoryHostsTotalProvisoningTime.
// The metric value specifies that a host exists in the Inventory
// with the given labels.
// The metric can be filtered by the labels to get a unique or a sum of hosts in
// a specific status mode.
func (h *InventoryHostsTotalProvisoningTime) PrometheusFormat() ([]prometheus.Metric, error) {
	metrics := make([]prometheus.Metric, 0, len(h.ProvisioningTime))

	h.Labels = []string{hostID, hostDeviceGUID, name, hostName, hostSerial, hostTenant}
	metricDesc := inventoryBuilder.NewMetricDesc(h.name, h.description, h.Labels, staticLabelsInventory)

	// For each one of the hosts indexed in the h.Status map, it adds a metric

	for _, hostProvisioningTime := range h.ProvisioningTime {
		gaugeVal := hostProvisioningTime.TotalProvisioningTime // set the Total Provisioning time
		metric := inventoryBuilder.MustNewConstMetric(
			metricDesc,
			prometheus.GaugeValue,
			gaugeVal,
			hostProvisioningTime.HostID,
			hostProvisioningTime.DeviceGUID,
			hostProvisioningTime.Name,
			hostProvisioningTime.Hostname,
			hostProvisioningTime.Serial,
			hostProvisioningTime.TenantID,
		)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}
