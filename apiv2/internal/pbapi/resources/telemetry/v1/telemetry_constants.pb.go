// Code generated by protoc-gen-go-const. DO NOT EDIT.

// source: resources/telemetry/v1/telemetry.proto

package telemetryv1

const (
	// Fields and Edges constants for "TelemetryLogsGroupResource"
	TelemetryLogsGroupResourceFieldResourceId           = "resource_id"
	TelemetryLogsGroupResourceFieldTelemetryLogsGroupId = "telemetry_logs_group_id"
	TelemetryLogsGroupResourceFieldName                 = "name"
	TelemetryLogsGroupResourceFieldCollectorKind        = "collector_kind"
	TelemetryLogsGroupResourceFieldGroups               = "groups"
	TelemetryLogsGroupResourceEdgeTimestamps            = "timestamps"

	// Fields and Edges constants for "TelemetryMetricsGroupResource"
	TelemetryMetricsGroupResourceFieldResourceId              = "resource_id"
	TelemetryMetricsGroupResourceFieldTelemetryMetricsGroupId = "telemetry_metrics_group_id"
	TelemetryMetricsGroupResourceFieldName                    = "name"
	TelemetryMetricsGroupResourceFieldCollectorKind           = "collector_kind"
	TelemetryMetricsGroupResourceFieldGroups                  = "groups"
	TelemetryMetricsGroupResourceEdgeTimestamps               = "timestamps"

	// Fields and Edges constants for "TelemetryLogsProfileResource"
	TelemetryLogsProfileResourceFieldResourceId     = "resource_id"
	TelemetryLogsProfileResourceFieldProfileId      = "profile_id"
	TelemetryLogsProfileResourceFieldTargetInstance = "target_instance"
	TelemetryLogsProfileResourceFieldTargetSite     = "target_site"
	TelemetryLogsProfileResourceFieldTargetRegion   = "target_region"
	TelemetryLogsProfileResourceFieldLogLevel       = "log_level"
	TelemetryLogsProfileResourceFieldLogsGroupId    = "logs_group_id"
	TelemetryLogsProfileResourceEdgeLogsGroup       = "logs_group"
	TelemetryLogsProfileResourceEdgeTimestamps      = "timestamps"

	// Fields and Edges constants for "TelemetryMetricsProfileResource"
	TelemetryMetricsProfileResourceFieldResourceId      = "resource_id"
	TelemetryMetricsProfileResourceFieldProfileId       = "profile_id"
	TelemetryMetricsProfileResourceFieldTargetInstance  = "target_instance"
	TelemetryMetricsProfileResourceFieldTargetSite      = "target_site"
	TelemetryMetricsProfileResourceFieldTargetRegion    = "target_region"
	TelemetryMetricsProfileResourceFieldMetricsInterval = "metrics_interval"
	TelemetryMetricsProfileResourceFieldMetricsGroupId  = "metrics_group_id"
	TelemetryMetricsProfileResourceEdgeMetricsGroup     = "metrics_group"
	TelemetryMetricsProfileResourceEdgeTimestamps       = "timestamps"
)
