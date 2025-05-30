// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: resources/schedule/v1/schedule.proto

package schedulev1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	v12 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	v11 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	v1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/location/v1"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The representation of a schedule's status.
type ScheduleStatus int32

const (
	ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED ScheduleStatus = 0
	// Generic maintenance.
	ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE ScheduleStatus = 1 // SCHEDULE_STATUS_SHIPPING = 2; // being shipped/in transit
	// for performing OS updates.
	ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE ScheduleStatus = 3
)

// Enum value maps for ScheduleStatus.
var (
	ScheduleStatus_name = map[int32]string{
		0: "SCHEDULE_STATUS_UNSPECIFIED",
		1: "SCHEDULE_STATUS_MAINTENANCE",
		3: "SCHEDULE_STATUS_OS_UPDATE",
	}
	ScheduleStatus_value = map[string]int32{
		"SCHEDULE_STATUS_UNSPECIFIED": 0,
		"SCHEDULE_STATUS_MAINTENANCE": 1,
		"SCHEDULE_STATUS_OS_UPDATE":   3,
	}
)

func (x ScheduleStatus) Enum() *ScheduleStatus {
	p := new(ScheduleStatus)
	*p = x
	return p
}

func (x ScheduleStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ScheduleStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_resources_schedule_v1_schedule_proto_enumTypes[0].Descriptor()
}

func (ScheduleStatus) Type() protoreflect.EnumType {
	return &file_resources_schedule_v1_schedule_proto_enumTypes[0]
}

func (x ScheduleStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ScheduleStatus.Descriptor instead.
func (ScheduleStatus) EnumDescriptor() ([]byte, []int) {
	return file_resources_schedule_v1_schedule_proto_rawDescGZIP(), []int{0}
}

// A single schedule resource.
type SingleScheduleResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Resource ID, generated by the inventory on Create.
	ResourceId string `protobuf:"bytes,1,opt,name=resource_id,json=resourceId,proto3" json:"resource_id,omitempty"`
	// The schedule status.
	ScheduleStatus ScheduleStatus `protobuf:"varint,2,opt,name=schedule_status,json=scheduleStatus,proto3,enum=resources.schedule.v1.ScheduleStatus" json:"schedule_status,omitempty"` // status of one-time-schedule
	// The schedule's name.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// Resource ID of Site this applies to.
	TargetSite *v1.SiteResource `protobuf:"bytes,4,opt,name=target_site,json=targetSite,proto3" json:"target_site,omitempty"`
	// Resource ID of Host this applies to.
	TargetHost *v11.HostResource `protobuf:"bytes,5,opt,name=target_host,json=targetHost,proto3" json:"target_host,omitempty"`
	// Resource ID of Region this applies to.
	TargetRegion *v1.RegionResource `protobuf:"bytes,7,opt,name=target_region,json=targetRegion,proto3" json:"target_region,omitempty"`
	// The start time in seconds, of the single schedule.
	StartSeconds uint32 `protobuf:"varint,9,opt,name=start_seconds,json=startSeconds,proto3" json:"start_seconds,omitempty"`
	// The end time in seconds, of the single schedule.
	// The value of endSeconds must be equal to or bigger than the value of startSeconds.
	EndSeconds uint32 `protobuf:"varint,10,opt,name=end_seconds,json=endSeconds,proto3" json:"end_seconds,omitempty"`
	// Deprecated, The single schedule resource's unique identifier. Alias of resourceId.
	SingleScheduleID string `protobuf:"bytes,5001,opt,name=single_scheduleID,json=singleScheduleID,proto3" json:"single_scheduleID,omitempty"`
	// The target host ID of the schedule.
	// Only one target can be provided per schedule.
	// This field cannot be used as filter.
	TargetHostId string `protobuf:"bytes,5002,opt,name=target_host_id,json=targetHostId,proto3" json:"target_host_id,omitempty"`
	// The target site ID of the schedule.
	// Only one target can be provided per schedule.
	// This field cannot be used as filter.
	TargetSiteId string `protobuf:"bytes,5003,opt,name=target_site_id,json=targetSiteId,proto3" json:"target_site_id,omitempty"`
	// The target region ID of the schedule.
	// Only one target can be provided per schedule.
	// This field cannot be used as filter.
	TargetRegionId string `protobuf:"bytes,5004,opt,name=target_region_id,json=targetRegionId,proto3" json:"target_region_id,omitempty"`
	// Timestamps associated to the resource.
	Timestamps *v12.Timestamps `protobuf:"bytes,50100,opt,name=timestamps,proto3" json:"timestamps,omitempty"`
}

func (x *SingleScheduleResource) Reset() {
	*x = SingleScheduleResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_resources_schedule_v1_schedule_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SingleScheduleResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SingleScheduleResource) ProtoMessage() {}

func (x *SingleScheduleResource) ProtoReflect() protoreflect.Message {
	mi := &file_resources_schedule_v1_schedule_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SingleScheduleResource.ProtoReflect.Descriptor instead.
func (*SingleScheduleResource) Descriptor() ([]byte, []int) {
	return file_resources_schedule_v1_schedule_proto_rawDescGZIP(), []int{0}
}

func (x *SingleScheduleResource) GetResourceId() string {
	if x != nil {
		return x.ResourceId
	}
	return ""
}

func (x *SingleScheduleResource) GetScheduleStatus() ScheduleStatus {
	if x != nil {
		return x.ScheduleStatus
	}
	return ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED
}

func (x *SingleScheduleResource) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SingleScheduleResource) GetTargetSite() *v1.SiteResource {
	if x != nil {
		return x.TargetSite
	}
	return nil
}

func (x *SingleScheduleResource) GetTargetHost() *v11.HostResource {
	if x != nil {
		return x.TargetHost
	}
	return nil
}

func (x *SingleScheduleResource) GetTargetRegion() *v1.RegionResource {
	if x != nil {
		return x.TargetRegion
	}
	return nil
}

func (x *SingleScheduleResource) GetStartSeconds() uint32 {
	if x != nil {
		return x.StartSeconds
	}
	return 0
}

func (x *SingleScheduleResource) GetEndSeconds() uint32 {
	if x != nil {
		return x.EndSeconds
	}
	return 0
}

func (x *SingleScheduleResource) GetSingleScheduleID() string {
	if x != nil {
		return x.SingleScheduleID
	}
	return ""
}

func (x *SingleScheduleResource) GetTargetHostId() string {
	if x != nil {
		return x.TargetHostId
	}
	return ""
}

func (x *SingleScheduleResource) GetTargetSiteId() string {
	if x != nil {
		return x.TargetSiteId
	}
	return ""
}

func (x *SingleScheduleResource) GetTargetRegionId() string {
	if x != nil {
		return x.TargetRegionId
	}
	return ""
}

func (x *SingleScheduleResource) GetTimestamps() *v12.Timestamps {
	if x != nil {
		return x.Timestamps
	}
	return nil
}

// A repeated-schedule resource.
type RepeatedScheduleResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Resource ID, generated by the inventory on Create.
	ResourceId string `protobuf:"bytes,1,opt,name=resource_id,json=resourceId,proto3" json:"resource_id,omitempty"`
	// The schedule status.
	ScheduleStatus ScheduleStatus `protobuf:"varint,2,opt,name=schedule_status,json=scheduleStatus,proto3,enum=resources.schedule.v1.ScheduleStatus" json:"schedule_status,omitempty"`
	// The schedule's name.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// Resource ID of Site this applies to.
	TargetSite *v1.SiteResource `protobuf:"bytes,4,opt,name=target_site,json=targetSite,proto3" json:"target_site,omitempty"`
	// Resource ID of Host this applies to.
	TargetHost *v11.HostResource `protobuf:"bytes,5,opt,name=target_host,json=targetHost,proto3" json:"target_host,omitempty"`
	// Resource ID of Region this applies to.
	TargetRegion *v1.RegionResource `protobuf:"bytes,21,opt,name=target_region,json=targetRegion,proto3" json:"target_region,omitempty"`
	// The duration in seconds of the repeated schedule, per schedule.
	DurationSeconds int32 `protobuf:"varint,6,opt,name=duration_seconds,json=durationSeconds,proto3" json:"duration_seconds,omitempty"`
	// cron style minutes (0-59), it can be empty only when used in a Filter.
	CronMinutes string `protobuf:"bytes,9,opt,name=cron_minutes,json=cronMinutes,proto3" json:"cron_minutes,omitempty"`
	// cron style hours (0-23), it can be empty only when used in a Filter
	CronHours string `protobuf:"bytes,10,opt,name=cron_hours,json=cronHours,proto3" json:"cron_hours,omitempty"`
	// cron style day of month (1-31), it can be empty only when used in a Filter
	CronDayMonth string `protobuf:"bytes,11,opt,name=cron_day_month,json=cronDayMonth,proto3" json:"cron_day_month,omitempty"`
	// cron style month (1-12), it can be empty only when used in a Filter
	CronMonth string `protobuf:"bytes,12,opt,name=cron_month,json=cronMonth,proto3" json:"cron_month,omitempty"`
	// cron style day of week (0-6), it can be empty only when used in a Filter
	CronDayWeek string `protobuf:"bytes,13,opt,name=cron_day_week,json=cronDayWeek,proto3" json:"cron_day_week,omitempty"`
	// Deprecated, The repeated schedule's unique identifier. Alias of resourceId.
	RepeatedScheduleID string `protobuf:"bytes,5001,opt,name=repeated_scheduleID,json=repeatedScheduleID,proto3" json:"repeated_scheduleID,omitempty"`
	// The target region ID of the schedule.
	// Only one target can be provided per schedule.
	// This field cannot be used as filter.
	TargetHostId string `protobuf:"bytes,5002,opt,name=target_host_id,json=targetHostId,proto3" json:"target_host_id,omitempty"`
	// The target site ID of the schedule.
	// Only one target can be provided per schedule.
	// This field cannot be used as filter.
	TargetSiteId string `protobuf:"bytes,5003,opt,name=target_site_id,json=targetSiteId,proto3" json:"target_site_id,omitempty"`
	// The target region ID of the schedule.
	// Only one target can be provided per schedule.
	// This field cannot be used as filter.
	TargetRegionId string `protobuf:"bytes,5004,opt,name=target_region_id,json=targetRegionId,proto3" json:"target_region_id,omitempty"`
	// Timestamps associated to the resource.
	Timestamps *v12.Timestamps `protobuf:"bytes,50100,opt,name=timestamps,proto3" json:"timestamps,omitempty"`
}

func (x *RepeatedScheduleResource) Reset() {
	*x = RepeatedScheduleResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_resources_schedule_v1_schedule_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepeatedScheduleResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepeatedScheduleResource) ProtoMessage() {}

func (x *RepeatedScheduleResource) ProtoReflect() protoreflect.Message {
	mi := &file_resources_schedule_v1_schedule_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepeatedScheduleResource.ProtoReflect.Descriptor instead.
func (*RepeatedScheduleResource) Descriptor() ([]byte, []int) {
	return file_resources_schedule_v1_schedule_proto_rawDescGZIP(), []int{1}
}

func (x *RepeatedScheduleResource) GetResourceId() string {
	if x != nil {
		return x.ResourceId
	}
	return ""
}

func (x *RepeatedScheduleResource) GetScheduleStatus() ScheduleStatus {
	if x != nil {
		return x.ScheduleStatus
	}
	return ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED
}

func (x *RepeatedScheduleResource) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RepeatedScheduleResource) GetTargetSite() *v1.SiteResource {
	if x != nil {
		return x.TargetSite
	}
	return nil
}

func (x *RepeatedScheduleResource) GetTargetHost() *v11.HostResource {
	if x != nil {
		return x.TargetHost
	}
	return nil
}

func (x *RepeatedScheduleResource) GetTargetRegion() *v1.RegionResource {
	if x != nil {
		return x.TargetRegion
	}
	return nil
}

func (x *RepeatedScheduleResource) GetDurationSeconds() int32 {
	if x != nil {
		return x.DurationSeconds
	}
	return 0
}

func (x *RepeatedScheduleResource) GetCronMinutes() string {
	if x != nil {
		return x.CronMinutes
	}
	return ""
}

func (x *RepeatedScheduleResource) GetCronHours() string {
	if x != nil {
		return x.CronHours
	}
	return ""
}

func (x *RepeatedScheduleResource) GetCronDayMonth() string {
	if x != nil {
		return x.CronDayMonth
	}
	return ""
}

func (x *RepeatedScheduleResource) GetCronMonth() string {
	if x != nil {
		return x.CronMonth
	}
	return ""
}

func (x *RepeatedScheduleResource) GetCronDayWeek() string {
	if x != nil {
		return x.CronDayWeek
	}
	return ""
}

func (x *RepeatedScheduleResource) GetRepeatedScheduleID() string {
	if x != nil {
		return x.RepeatedScheduleID
	}
	return ""
}

func (x *RepeatedScheduleResource) GetTargetHostId() string {
	if x != nil {
		return x.TargetHostId
	}
	return ""
}

func (x *RepeatedScheduleResource) GetTargetSiteId() string {
	if x != nil {
		return x.TargetSiteId
	}
	return ""
}

func (x *RepeatedScheduleResource) GetTargetRegionId() string {
	if x != nil {
		return x.TargetRegionId
	}
	return ""
}

func (x *RepeatedScheduleResource) GetTimestamps() *v12.Timestamps {
	if x != nil {
		return x.Timestamps
	}
	return nil
}

var File_resources_schedule_v1_schedule_proto protoreflect.FileDescriptor

var file_resources_schedule_v1_schedule_proto_rawDesc = []byte{
	0x0a, 0x24, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x73, 0x63, 0x68, 0x65,
	0x64, 0x75, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x73, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f,
	0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20,
	0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x22, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x70,
	0x75, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x6f, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb7, 0x07, 0x0a, 0x16, 0x53, 0x69, 0x6e, 0x67,
	0x6c, 0x65, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x12, 0x45, 0x0a, 0x0b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x24, 0xe0, 0x41, 0x03, 0xba, 0x48, 0x1e, 0x72,
	0x1c, 0x18, 0x13, 0x32, 0x18, 0x5e, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x73, 0x63, 0x68, 0x65,
	0x2d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0a, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64, 0x12, 0x53, 0x0a, 0x0f, 0x73, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x25, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x73,
	0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0e,
	0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x34,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x20, 0xba, 0x48,
	0x1d, 0x72, 0x1b, 0x18, 0x32, 0x32, 0x17, 0x5e, 0x24, 0x7c, 0x5e, 0x5b, 0x61, 0x2d, 0x7a, 0x41,
	0x2d, 0x5a, 0x2d, 0x5f, 0x30, 0x2d, 0x39, 0x2e, 0x2f, 0x3a, 0x20, 0x5d, 0x2b, 0x24, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x49, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x73,
	0x69, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03,
	0xe0, 0x41, 0x03, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x53, 0x69, 0x74, 0x65, 0x12,
	0x48, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x2e, 0x63, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x6f, 0x73, 0x74,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x4f, 0x0a, 0x0d, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x25, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x6c, 0x6f, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0c, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x35, 0x0a, 0x0d, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x0d, 0x42, 0x10, 0xe0, 0x41, 0x02, 0xba, 0x48, 0x0a, 0x2a, 0x08, 0x18, 0xff, 0xff, 0xff, 0xff,
	0x0f, 0x28, 0x01, 0x52, 0x0c, 0x73, 0x74, 0x61, 0x72, 0x74, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x73, 0x12, 0x2e, 0x0a, 0x0b, 0x65, 0x6e, 0x64, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x0d, 0xba, 0x48, 0x0a, 0x2a, 0x08, 0x18, 0xff, 0xff,
	0xff, 0xff, 0x0f, 0x28, 0x01, 0x52, 0x0a, 0x65, 0x6e, 0x64, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64,
	0x73, 0x12, 0x52, 0x0a, 0x11, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x5f, 0x73, 0x63, 0x68, 0x65,
	0x64, 0x75, 0x6c, 0x65, 0x49, 0x44, 0x18, 0x89, 0x27, 0x20, 0x01, 0x28, 0x09, 0x42, 0x24, 0xe0,
	0x41, 0x03, 0xba, 0x48, 0x1e, 0x72, 0x1c, 0x18, 0x13, 0x32, 0x18, 0x5e, 0x73, 0x69, 0x6e, 0x67,
	0x6c, 0x65, 0x73, 0x63, 0x68, 0x65, 0x2d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b,
	0x38, 0x7d, 0x24, 0x52, 0x10, 0x73, 0x69, 0x6e, 0x67, 0x6c, 0x65, 0x53, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x65, 0x49, 0x44, 0x12, 0x48, 0x0a, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f,
	0x68, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x8a, 0x27, 0x20, 0x01, 0x28, 0x09, 0x42, 0x21,
	0xe0, 0x41, 0x04, 0xba, 0x48, 0x1b, 0x72, 0x19, 0x18, 0x0d, 0x32, 0x15, 0x5e, 0x24, 0x7c, 0x5e,
	0x68, 0x6f, 0x73, 0x74, 0x2d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d,
	0x24, 0x52, 0x0c, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x49, 0x64, 0x12,
	0x48, 0x0a, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x73, 0x69, 0x74, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x8b, 0x27, 0x20, 0x01, 0x28, 0x09, 0x42, 0x21, 0xe0, 0x41, 0x04, 0xba, 0x48, 0x1b,
	0x72, 0x19, 0x18, 0x0d, 0x32, 0x15, 0x5e, 0x24, 0x7c, 0x5e, 0x73, 0x69, 0x74, 0x65, 0x2d, 0x5b,
	0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0c, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x53, 0x69, 0x74, 0x65, 0x49, 0x64, 0x12, 0x4e, 0x0a, 0x10, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x8c, 0x27,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x23, 0xe0, 0x41, 0x04, 0xba, 0x48, 0x1d, 0x72, 0x1b, 0x18, 0x0f,
	0x32, 0x17, 0x5e, 0x24, 0x7c, 0x5e, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2d, 0x5b, 0x30, 0x2d,
	0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x46, 0x0a, 0x0a, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x73, 0x18, 0xb4, 0x87, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1f, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x73,
	0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x73, 0x22, 0xf0, 0x0a, 0x0a, 0x18, 0x52, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64, 0x53, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x47,
	0x0a, 0x0b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x26, 0xe0, 0x41, 0x03, 0xba, 0x48, 0x20, 0x72, 0x1e, 0x18, 0x15, 0x32,
	0x1a, 0x5e, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64, 0x73, 0x63, 0x68, 0x65, 0x2d, 0x5b,
	0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0a, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x49, 0x64, 0x12, 0x53, 0x0a, 0x0f, 0x73, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x25, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x73, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c,
	0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x0e, 0x73, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x34, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x20, 0xba, 0x48, 0x1d, 0x72,
	0x1b, 0x18, 0x32, 0x32, 0x17, 0x5e, 0x24, 0x7c, 0x5e, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a,
	0x2d, 0x5f, 0x30, 0x2d, 0x39, 0x2e, 0x2f, 0x3a, 0x20, 0x5d, 0x2b, 0x24, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x49, 0x0a, 0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x73, 0x69, 0x74,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x73, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x69, 0x74, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03, 0xe0, 0x41,
	0x03, 0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x53, 0x69, 0x74, 0x65, 0x12, 0x48, 0x0a,
	0x0b, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x22, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x63,
	0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x48, 0x6f, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x74, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x48, 0x6f, 0x73, 0x74, 0x12, 0x4f, 0x0a, 0x0d, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x15, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0c, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x39, 0x0a, 0x10, 0x64, 0x75, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x05, 0x42, 0x0e, 0xe0, 0x41, 0x02, 0xba, 0x48, 0x08, 0x1a, 0x06, 0x18, 0x80, 0xa3, 0x05,
	0x28, 0x01, 0x52, 0x0f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x63, 0x6f,
	0x6e, 0x64, 0x73, 0x12, 0x63, 0x0a, 0x0c, 0x63, 0x72, 0x6f, 0x6e, 0x5f, 0x6d, 0x69, 0x6e, 0x75,
	0x74, 0x65, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x42, 0x40, 0xe0, 0x41, 0x02, 0xba, 0x48,
	0x3a, 0x72, 0x38, 0x32, 0x36, 0x5e, 0x28, 0x5b, 0x2a, 0x5d, 0x7c, 0x28, 0x5b, 0x30, 0x2d, 0x39,
	0x5d, 0x7c, 0x28, 0x5b, 0x31, 0x2d, 0x35, 0x5d, 0x5b, 0x30, 0x2d, 0x39, 0x5d, 0x29, 0x29, 0x28,
	0x28, 0x2c, 0x28, 0x5b, 0x30, 0x2d, 0x39, 0x5d, 0x7c, 0x28, 0x5b, 0x31, 0x2d, 0x35, 0x5d, 0x5b,
	0x30, 0x2d, 0x39, 0x5d, 0x29, 0x29, 0x29, 0x2a, 0x29, 0x29, 0x24, 0x52, 0x0b, 0x63, 0x72, 0x6f,
	0x6e, 0x4d, 0x69, 0x6e, 0x75, 0x74, 0x65, 0x73, 0x12, 0x61, 0x0a, 0x0a, 0x63, 0x72, 0x6f, 0x6e,
	0x5f, 0x68, 0x6f, 0x75, 0x72, 0x73, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x42, 0x42, 0xe0, 0x41,
	0x02, 0xba, 0x48, 0x3c, 0x72, 0x3a, 0x32, 0x38, 0x5e, 0x28, 0x5b, 0x2a, 0x5d, 0x7c, 0x28, 0x5b,
	0x30, 0x2d, 0x39, 0x5d, 0x7c, 0x31, 0x5b, 0x30, 0x2d, 0x39, 0x5d, 0x7c, 0x32, 0x5b, 0x30, 0x2d,
	0x33, 0x5d, 0x29, 0x28, 0x28, 0x2c, 0x28, 0x5b, 0x30, 0x2d, 0x39, 0x5d, 0x7c, 0x31, 0x5b, 0x30,
	0x2d, 0x39, 0x5d, 0x7c, 0x32, 0x5b, 0x30, 0x2d, 0x33, 0x5d, 0x29, 0x29, 0x2a, 0x29, 0x29, 0x24,
	0x52, 0x09, 0x63, 0x72, 0x6f, 0x6e, 0x48, 0x6f, 0x75, 0x72, 0x73, 0x12, 0x70, 0x0a, 0x0e, 0x63,
	0x72, 0x6f, 0x6e, 0x5f, 0x64, 0x61, 0x79, 0x5f, 0x6d, 0x6f, 0x6e, 0x74, 0x68, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x4a, 0xe0, 0x41, 0x02, 0xba, 0x48, 0x44, 0x72, 0x42, 0x32, 0x40, 0x5e,
	0x28, 0x5b, 0x2a, 0x5d, 0x7c, 0x28, 0x5b, 0x31, 0x2d, 0x39, 0x5d, 0x7c, 0x28, 0x5b, 0x31, 0x32,
	0x5d, 0x5b, 0x30, 0x2d, 0x39, 0x5d, 0x29, 0x7c, 0x33, 0x5b, 0x30, 0x31, 0x5d, 0x29, 0x28, 0x28,
	0x2c, 0x28, 0x5b, 0x31, 0x2d, 0x39, 0x5d, 0x7c, 0x28, 0x5b, 0x31, 0x32, 0x5d, 0x5b, 0x30, 0x2d,
	0x39, 0x5d, 0x29, 0x7c, 0x33, 0x5b, 0x30, 0x31, 0x5d, 0x29, 0x29, 0x2a, 0x29, 0x29, 0x24, 0x52,
	0x0c, 0x63, 0x72, 0x6f, 0x6e, 0x44, 0x61, 0x79, 0x4d, 0x6f, 0x6e, 0x74, 0x68, 0x12, 0x53, 0x0a,
	0x0a, 0x63, 0x72, 0x6f, 0x6e, 0x5f, 0x6d, 0x6f, 0x6e, 0x74, 0x68, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x34, 0xe0, 0x41, 0x02, 0xba, 0x48, 0x2e, 0x72, 0x2c, 0x32, 0x2a, 0x5e, 0x28, 0x5b,
	0x2a, 0x5d, 0x7c, 0x28, 0x5b, 0x31, 0x2d, 0x39, 0x5d, 0x7c, 0x31, 0x5b, 0x30, 0x31, 0x32, 0x5d,
	0x29, 0x28, 0x28, 0x2c, 0x28, 0x5b, 0x31, 0x2d, 0x39, 0x5d, 0x7c, 0x31, 0x5b, 0x30, 0x31, 0x32,
	0x5d, 0x29, 0x29, 0x2a, 0x29, 0x29, 0x24, 0x52, 0x09, 0x63, 0x72, 0x6f, 0x6e, 0x4d, 0x6f, 0x6e,
	0x74, 0x68, 0x12, 0x4a, 0x0a, 0x0d, 0x63, 0x72, 0x6f, 0x6e, 0x5f, 0x64, 0x61, 0x79, 0x5f, 0x77,
	0x65, 0x65, 0x6b, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x42, 0x26, 0xe0, 0x41, 0x02, 0xba, 0x48,
	0x20, 0x72, 0x1e, 0x32, 0x1c, 0x5e, 0x28, 0x5b, 0x2a, 0x5d, 0x7c, 0x28, 0x5b, 0x30, 0x2d, 0x36,
	0x5d, 0x29, 0x28, 0x28, 0x2c, 0x28, 0x5b, 0x30, 0x2d, 0x36, 0x5d, 0x29, 0x29, 0x2a, 0x29, 0x29,
	0x24, 0x52, 0x0b, 0x63, 0x72, 0x6f, 0x6e, 0x44, 0x61, 0x79, 0x57, 0x65, 0x65, 0x6b, 0x12, 0x58,
	0x0a, 0x13, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x73, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x65, 0x49, 0x44, 0x18, 0x89, 0x27, 0x20, 0x01, 0x28, 0x09, 0x42, 0x26, 0xe0, 0x41,
	0x03, 0xba, 0x48, 0x20, 0x72, 0x1e, 0x18, 0x15, 0x32, 0x1a, 0x5e, 0x72, 0x65, 0x70, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x73, 0x63, 0x68, 0x65, 0x2d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d,
	0x7b, 0x38, 0x7d, 0x24, 0x52, 0x12, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64, 0x53, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x49, 0x44, 0x12, 0x48, 0x0a, 0x0e, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x8a, 0x27, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x21, 0xe0, 0x41, 0x04, 0xba, 0x48, 0x1b, 0x72, 0x19, 0x18, 0x0d, 0x32, 0x15, 0x5e,
	0x24, 0x7c, 0x5e, 0x68, 0x6f, 0x73, 0x74, 0x2d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d,
	0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0c, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x48, 0x6f, 0x73, 0x74,
	0x49, 0x64, 0x12, 0x48, 0x0a, 0x0e, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x73, 0x69, 0x74,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x8b, 0x27, 0x20, 0x01, 0x28, 0x09, 0x42, 0x21, 0xe0, 0x41, 0x04,
	0xba, 0x48, 0x1b, 0x72, 0x19, 0x18, 0x0d, 0x32, 0x15, 0x5e, 0x24, 0x7c, 0x5e, 0x73, 0x69, 0x74,
	0x65, 0x2d, 0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0c,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x53, 0x69, 0x74, 0x65, 0x49, 0x64, 0x12, 0x4e, 0x0a, 0x10,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64,
	0x18, 0x8c, 0x27, 0x20, 0x01, 0x28, 0x09, 0x42, 0x23, 0xe0, 0x41, 0x04, 0xba, 0x48, 0x1d, 0x72,
	0x1b, 0x18, 0x0f, 0x32, 0x17, 0x5e, 0x24, 0x7c, 0x5e, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x2d,
	0x5b, 0x30, 0x2d, 0x39, 0x61, 0x2d, 0x66, 0x5d, 0x7b, 0x38, 0x7d, 0x24, 0x52, 0x0e, 0x74, 0x61,
	0x72, 0x67, 0x65, 0x74, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x46, 0x0a, 0x0a,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x73, 0x18, 0xb4, 0x87, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x73, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x73, 0x2a, 0x71, 0x0a, 0x0e, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1f, 0x0a, 0x1b, 0x53, 0x43, 0x48, 0x45, 0x44, 0x55,
	0x4c, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1f, 0x0a, 0x1b, 0x53, 0x43, 0x48, 0x45, 0x44,
	0x55, 0x4c, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x4d, 0x41, 0x49, 0x4e, 0x54,
	0x45, 0x4e, 0x41, 0x4e, 0x43, 0x45, 0x10, 0x01, 0x12, 0x1d, 0x0a, 0x19, 0x53, 0x43, 0x48, 0x45,
	0x44, 0x55, 0x4c, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x4f, 0x53, 0x5f, 0x55,
	0x50, 0x44, 0x41, 0x54, 0x45, 0x10, 0x03, 0x42, 0x63, 0x5a, 0x61, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x2d, 0x65, 0x64, 0x67, 0x65, 0x2d,
	0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x69, 0x6e, 0x66, 0x72, 0x61, 0x2d, 0x63,
	0x6f, 0x72, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x76, 0x32, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x2f, 0x76,
	0x31, 0x3b, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_resources_schedule_v1_schedule_proto_rawDescOnce sync.Once
	file_resources_schedule_v1_schedule_proto_rawDescData = file_resources_schedule_v1_schedule_proto_rawDesc
)

func file_resources_schedule_v1_schedule_proto_rawDescGZIP() []byte {
	file_resources_schedule_v1_schedule_proto_rawDescOnce.Do(func() {
		file_resources_schedule_v1_schedule_proto_rawDescData = protoimpl.X.CompressGZIP(file_resources_schedule_v1_schedule_proto_rawDescData)
	})
	return file_resources_schedule_v1_schedule_proto_rawDescData
}

var file_resources_schedule_v1_schedule_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_resources_schedule_v1_schedule_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_resources_schedule_v1_schedule_proto_goTypes = []interface{}{
	(ScheduleStatus)(0),              // 0: resources.schedule.v1.ScheduleStatus
	(*SingleScheduleResource)(nil),   // 1: resources.schedule.v1.SingleScheduleResource
	(*RepeatedScheduleResource)(nil), // 2: resources.schedule.v1.RepeatedScheduleResource
	(*v1.SiteResource)(nil),          // 3: resources.location.v1.SiteResource
	(*v11.HostResource)(nil),         // 4: resources.compute.v1.HostResource
	(*v1.RegionResource)(nil),        // 5: resources.location.v1.RegionResource
	(*v12.Timestamps)(nil),           // 6: resources.common.v1.Timestamps
}
var file_resources_schedule_v1_schedule_proto_depIdxs = []int32{
	0,  // 0: resources.schedule.v1.SingleScheduleResource.schedule_status:type_name -> resources.schedule.v1.ScheduleStatus
	3,  // 1: resources.schedule.v1.SingleScheduleResource.target_site:type_name -> resources.location.v1.SiteResource
	4,  // 2: resources.schedule.v1.SingleScheduleResource.target_host:type_name -> resources.compute.v1.HostResource
	5,  // 3: resources.schedule.v1.SingleScheduleResource.target_region:type_name -> resources.location.v1.RegionResource
	6,  // 4: resources.schedule.v1.SingleScheduleResource.timestamps:type_name -> resources.common.v1.Timestamps
	0,  // 5: resources.schedule.v1.RepeatedScheduleResource.schedule_status:type_name -> resources.schedule.v1.ScheduleStatus
	3,  // 6: resources.schedule.v1.RepeatedScheduleResource.target_site:type_name -> resources.location.v1.SiteResource
	4,  // 7: resources.schedule.v1.RepeatedScheduleResource.target_host:type_name -> resources.compute.v1.HostResource
	5,  // 8: resources.schedule.v1.RepeatedScheduleResource.target_region:type_name -> resources.location.v1.RegionResource
	6,  // 9: resources.schedule.v1.RepeatedScheduleResource.timestamps:type_name -> resources.common.v1.Timestamps
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_resources_schedule_v1_schedule_proto_init() }
func file_resources_schedule_v1_schedule_proto_init() {
	if File_resources_schedule_v1_schedule_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_resources_schedule_v1_schedule_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SingleScheduleResource); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_resources_schedule_v1_schedule_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepeatedScheduleResource); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_resources_schedule_v1_schedule_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_resources_schedule_v1_schedule_proto_goTypes,
		DependencyIndexes: file_resources_schedule_v1_schedule_proto_depIdxs,
		EnumInfos:         file_resources_schedule_v1_schedule_proto_enumTypes,
		MessageInfos:      file_resources_schedule_v1_schedule_proto_msgTypes,
	}.Build()
	File_resources_schedule_v1_schedule_proto = out.File
	file_resources_schedule_v1_schedule_proto_rawDesc = nil
	file_resources_schedule_v1_schedule_proto_goTypes = nil
	file_resources_schedule_v1_schedule_proto_depIdxs = nil
}
