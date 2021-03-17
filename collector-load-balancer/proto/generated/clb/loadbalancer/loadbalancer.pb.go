// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: loadbalancer/loadbalancer.proto

package loadbalancer

import (
	clb "github.com/aws-observability/collector-load-balancer/proto/generated/clb"
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

type GetPrometheusTargetsReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TODO: might use one off, can we also use UID?
	PodIp   string `protobuf:"bytes,1,opt,name=pod_ip,json=podIp,proto3" json:"pod_ip,omitempty"`
	PodName string `protobuf:"bytes,2,opt,name=pod_name,json=podName,proto3" json:"pod_name,omitempty"`
}

func (x *GetPrometheusTargetsReq) Reset() {
	*x = GetPrometheusTargetsReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loadbalancer_loadbalancer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPrometheusTargetsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPrometheusTargetsReq) ProtoMessage() {}

func (x *GetPrometheusTargetsReq) ProtoReflect() protoreflect.Message {
	mi := &file_loadbalancer_loadbalancer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPrometheusTargetsReq.ProtoReflect.Descriptor instead.
func (*GetPrometheusTargetsReq) Descriptor() ([]byte, []int) {
	return file_loadbalancer_loadbalancer_proto_rawDescGZIP(), []int{0}
}

func (x *GetPrometheusTargetsReq) GetPodIp() string {
	if x != nil {
		return x.PodIp
	}
	return ""
}

func (x *GetPrometheusTargetsReq) GetPodName() string {
	if x != nil {
		return x.PodName
	}
	return ""
}

type GetPrometheusTargetsRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Files []*clb.PrometheusFileSD `protobuf:"bytes,1,rep,name=files,proto3" json:"files,omitempty"`
}

func (x *GetPrometheusTargetsRes) Reset() {
	*x = GetPrometheusTargetsRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loadbalancer_loadbalancer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPrometheusTargetsRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPrometheusTargetsRes) ProtoMessage() {}

func (x *GetPrometheusTargetsRes) ProtoReflect() protoreflect.Message {
	mi := &file_loadbalancer_loadbalancer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPrometheusTargetsRes.ProtoReflect.Descriptor instead.
func (*GetPrometheusTargetsRes) Descriptor() ([]byte, []int) {
	return file_loadbalancer_loadbalancer_proto_rawDescGZIP(), []int{1}
}

func (x *GetPrometheusTargetsRes) GetFiles() []*clb.PrometheusFileSD {
	if x != nil {
		return x.Files
	}
	return nil
}

var File_loadbalancer_loadbalancer_proto protoreflect.FileDescriptor

var file_loadbalancer_loadbalancer_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6c, 0x6f, 0x61, 0x64, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x2f, 0x6c,
	0x6f, 0x61, 0x64, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x03, 0x63, 0x6c, 0x62, 0x1a, 0x09, 0x63, 0x6c, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x4b, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65,
	0x75, 0x73, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x12, 0x15, 0x0a, 0x06,
	0x70, 0x6f, 0x64, 0x5f, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x6f,
	0x64, 0x49, 0x70, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x6f, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x4b,
	0x0a, 0x17, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x54,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x12, 0x30, 0x0a, 0x05, 0x66, 0x69, 0x6c,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6f, 0x74, 0x65, 0x6c, 0x2e,
	0x63, 0x6c, 0x62, 0x2e, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65, 0x75, 0x73, 0x46, 0x69,
	0x6c, 0x65, 0x53, 0x44, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x32, 0x5f, 0x0a, 0x13, 0x4c,
	0x6f, 0x61, 0x64, 0x42, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x48, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73,
	0x12, 0x1c, 0x2e, 0x63, 0x6c, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74,
	0x68, 0x65, 0x75, 0x73, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x52, 0x65, 0x71, 0x1a, 0x1c,
	0x2e, 0x63, 0x6c, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x65,
	0x75, 0x73, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x52, 0x65, 0x73, 0x42, 0x57, 0x5a, 0x55,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x77, 0x73, 0x2d, 0x6f,
	0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x2f, 0x63, 0x6f, 0x6c,
	0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2d, 0x6c, 0x6f, 0x61, 0x64, 0x2d, 0x62, 0x61, 0x6c, 0x61,
	0x6e, 0x63, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72,
	0x61, 0x74, 0x65, 0x64, 0x2f, 0x63, 0x6c, 0x62, 0x2f, 0x6c, 0x6f, 0x61, 0x64, 0x62, 0x61, 0x6c,
	0x61, 0x6e, 0x63, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_loadbalancer_loadbalancer_proto_rawDescOnce sync.Once
	file_loadbalancer_loadbalancer_proto_rawDescData = file_loadbalancer_loadbalancer_proto_rawDesc
)

func file_loadbalancer_loadbalancer_proto_rawDescGZIP() []byte {
	file_loadbalancer_loadbalancer_proto_rawDescOnce.Do(func() {
		file_loadbalancer_loadbalancer_proto_rawDescData = protoimpl.X.CompressGZIP(file_loadbalancer_loadbalancer_proto_rawDescData)
	})
	return file_loadbalancer_loadbalancer_proto_rawDescData
}

var file_loadbalancer_loadbalancer_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_loadbalancer_loadbalancer_proto_goTypes = []interface{}{
	(*GetPrometheusTargetsReq)(nil), // 0: clb.GetPrometheusTargetsReq
	(*GetPrometheusTargetsRes)(nil), // 1: clb.GetPrometheusTargetsRes
	(*clb.PrometheusFileSD)(nil),    // 2: otel.clb.PrometheusFileSD
}
var file_loadbalancer_loadbalancer_proto_depIdxs = []int32{
	2, // 0: clb.GetPrometheusTargetsRes.files:type_name -> otel.clb.PrometheusFileSD
	0, // 1: clb.LoadBalancerService.GetTargets:input_type -> clb.GetPrometheusTargetsReq
	1, // 2: clb.LoadBalancerService.GetTargets:output_type -> clb.GetPrometheusTargetsRes
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_loadbalancer_loadbalancer_proto_init() }
func file_loadbalancer_loadbalancer_proto_init() {
	if File_loadbalancer_loadbalancer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_loadbalancer_loadbalancer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPrometheusTargetsReq); i {
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
		file_loadbalancer_loadbalancer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPrometheusTargetsRes); i {
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
			RawDescriptor: file_loadbalancer_loadbalancer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_loadbalancer_loadbalancer_proto_goTypes,
		DependencyIndexes: file_loadbalancer_loadbalancer_proto_depIdxs,
		MessageInfos:      file_loadbalancer_loadbalancer_proto_msgTypes,
	}.Build()
	File_loadbalancer_loadbalancer_proto = out.File
	file_loadbalancer_loadbalancer_proto_rawDesc = nil
	file_loadbalancer_loadbalancer_proto_goTypes = nil
	file_loadbalancer_loadbalancer_proto_depIdxs = nil
}