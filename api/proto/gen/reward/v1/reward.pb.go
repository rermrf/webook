// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: reward/v1/reward.proto

package rewardv1

import (
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

type RewardStatus int32

const (
	RewardStatus_RewardStatusUnknown RewardStatus = 0
	RewardStatus_RewardStatusInit    RewardStatus = 1
	RewardStatus_RewardStatusPayed   RewardStatus = 2
	RewardStatus_RewardStatusFailed  RewardStatus = 3
)

// Enum value maps for RewardStatus.
var (
	RewardStatus_name = map[int32]string{
		0: "RewardStatusUnknown",
		1: "RewardStatusInit",
		2: "RewardStatusPayed",
		3: "RewardStatusFailed",
	}
	RewardStatus_value = map[string]int32{
		"RewardStatusUnknown": 0,
		"RewardStatusInit":    1,
		"RewardStatusPayed":   2,
		"RewardStatusFailed":  3,
	}
)

func (x RewardStatus) Enum() *RewardStatus {
	p := new(RewardStatus)
	*p = x
	return p
}

func (x RewardStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RewardStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_reward_v1_reward_proto_enumTypes[0].Descriptor()
}

func (RewardStatus) Type() protoreflect.EnumType {
	return &file_reward_v1_reward_proto_enumTypes[0]
}

func (x RewardStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RewardStatus.Descriptor instead.
func (RewardStatus) EnumDescriptor() ([]byte, []int) {
	return file_reward_v1_reward_proto_rawDescGZIP(), []int{0}
}

type PreRewardRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz     string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId   int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	BizName string `protobuf:"bytes,3,opt,name=biz_name,json=bizName,proto3" json:"biz_name,omitempty"`
	// 被打赏的人，也就是收钱的那个
	TargetUid int64 `protobuf:"varint,4,opt,name=target_uid,json=targetUid,proto3" json:"target_uid,omitempty"`
	// 打赏的人，也就是付钱的那个
	Uid int64 `protobuf:"varint,5,opt,name=uid,proto3" json:"uid,omitempty"`
	// 打赏金额
	Amt int64 `protobuf:"varint,6,opt,name=amt,proto3" json:"amt,omitempty"`
}

func (x *PreRewardRequest) Reset() {
	*x = PreRewardRequest{}
	mi := &file_reward_v1_reward_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PreRewardRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreRewardRequest) ProtoMessage() {}

func (x *PreRewardRequest) ProtoReflect() protoreflect.Message {
	mi := &file_reward_v1_reward_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreRewardRequest.ProtoReflect.Descriptor instead.
func (*PreRewardRequest) Descriptor() ([]byte, []int) {
	return file_reward_v1_reward_proto_rawDescGZIP(), []int{0}
}

func (x *PreRewardRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *PreRewardRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *PreRewardRequest) GetBizName() string {
	if x != nil {
		return x.BizName
	}
	return ""
}

func (x *PreRewardRequest) GetTargetUid() int64 {
	if x != nil {
		return x.TargetUid
	}
	return 0
}

func (x *PreRewardRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *PreRewardRequest) GetAmt() int64 {
	if x != nil {
		return x.Amt
	}
	return 0
}

type PreRewardResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rid     int64  `protobuf:"varint,1,opt,name=rid,proto3" json:"rid,omitempty"`
	CodeUrl string `protobuf:"bytes,2,opt,name=code_url,json=codeUrl,proto3" json:"code_url,omitempty"`
}

func (x *PreRewardResponse) Reset() {
	*x = PreRewardResponse{}
	mi := &file_reward_v1_reward_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PreRewardResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreRewardResponse) ProtoMessage() {}

func (x *PreRewardResponse) ProtoReflect() protoreflect.Message {
	mi := &file_reward_v1_reward_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreRewardResponse.ProtoReflect.Descriptor instead.
func (*PreRewardResponse) Descriptor() ([]byte, []int) {
	return file_reward_v1_reward_proto_rawDescGZIP(), []int{1}
}

func (x *PreRewardResponse) GetRid() int64 {
	if x != nil {
		return x.Rid
	}
	return 0
}

func (x *PreRewardResponse) GetCodeUrl() string {
	if x != nil {
		return x.CodeUrl
	}
	return ""
}

type GetRewardRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rid int64 `protobuf:"varint,1,opt,name=rid,proto3" json:"rid,omitempty"`
	Uid int64 `protobuf:"varint,2,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *GetRewardRequest) Reset() {
	*x = GetRewardRequest{}
	mi := &file_reward_v1_reward_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRewardRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRewardRequest) ProtoMessage() {}

func (x *GetRewardRequest) ProtoReflect() protoreflect.Message {
	mi := &file_reward_v1_reward_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRewardRequest.ProtoReflect.Descriptor instead.
func (*GetRewardRequest) Descriptor() ([]byte, []int) {
	return file_reward_v1_reward_proto_rawDescGZIP(), []int{2}
}

func (x *GetRewardRequest) GetRid() int64 {
	if x != nil {
		return x.Rid
	}
	return 0
}

func (x *GetRewardRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

// 对于调用这个服务的人来说，我只需要知道这比订单的状态
type GetRewardResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status RewardStatus `protobuf:"varint,1,opt,name=status,proto3,enum=reward.v1.RewardStatus" json:"status,omitempty"`
}

func (x *GetRewardResponse) Reset() {
	*x = GetRewardResponse{}
	mi := &file_reward_v1_reward_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRewardResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRewardResponse) ProtoMessage() {}

func (x *GetRewardResponse) ProtoReflect() protoreflect.Message {
	mi := &file_reward_v1_reward_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRewardResponse.ProtoReflect.Descriptor instead.
func (*GetRewardResponse) Descriptor() ([]byte, []int) {
	return file_reward_v1_reward_proto_rawDescGZIP(), []int{3}
}

func (x *GetRewardResponse) GetStatus() RewardStatus {
	if x != nil {
		return x.Status
	}
	return RewardStatus_RewardStatusUnknown
}

var File_reward_v1_reward_proto protoreflect.FileDescriptor

var file_reward_v1_reward_proto_rawDesc = []byte{
	0x0a, 0x16, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64,
	0x2e, 0x76, 0x31, 0x22, 0x99, 0x01, 0x0a, 0x10, 0x50, 0x72, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72,
	0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x62, 0x69, 0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69,
	0x7a, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49,
	0x64, 0x12, 0x19, 0x0a, 0x08, 0x62, 0x69, 0x7a, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x62, 0x69, 0x7a, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x75, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x55, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a,
	0x03, 0x61, 0x6d, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x61, 0x6d, 0x74, 0x22,
	0x40, 0x0a, 0x11, 0x50, 0x72, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x03, 0x72, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x64, 0x65, 0x55, 0x72,
	0x6c, 0x22, 0x36, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x03, 0x72, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x44, 0x0a, 0x11, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17,
	0x2e, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x77, 0x61, 0x72,
	0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a,
	0x6c, 0x0a, 0x0c, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x17, 0x0a, 0x13, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55,
	0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x52, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x49, 0x6e, 0x69, 0x74, 0x10, 0x01, 0x12, 0x15,
	0x0a, 0x11, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x50, 0x61,
	0x79, 0x65, 0x64, 0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x10, 0x03, 0x32, 0x9f, 0x01,
	0x0a, 0x0d, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x46, 0x0a, 0x09, 0x50, 0x72, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x12, 0x1b, 0x2e, 0x72,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x65, 0x52, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x72, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x72, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x46, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x77, 0x61, 0x72, 0x64, 0x12, 0x1b, 0x2e, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1c, 0x2e, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0x8a, 0x01, 0x0a, 0x0d, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2e, 0x76,
	0x31, 0x42, 0x0b, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x27, 0x77, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2f, 0x76, 0x31,
	0x3b, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x58, 0x58, 0xaa,
	0x02, 0x09, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x09, 0x52, 0x65,
	0x77, 0x61, 0x72, 0x64, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x15, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x0a, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_reward_v1_reward_proto_rawDescOnce sync.Once
	file_reward_v1_reward_proto_rawDescData = file_reward_v1_reward_proto_rawDesc
)

func file_reward_v1_reward_proto_rawDescGZIP() []byte {
	file_reward_v1_reward_proto_rawDescOnce.Do(func() {
		file_reward_v1_reward_proto_rawDescData = protoimpl.X.CompressGZIP(file_reward_v1_reward_proto_rawDescData)
	})
	return file_reward_v1_reward_proto_rawDescData
}

var file_reward_v1_reward_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_reward_v1_reward_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_reward_v1_reward_proto_goTypes = []any{
	(RewardStatus)(0),         // 0: reward.v1.RewardStatus
	(*PreRewardRequest)(nil),  // 1: reward.v1.PreRewardRequest
	(*PreRewardResponse)(nil), // 2: reward.v1.PreRewardResponse
	(*GetRewardRequest)(nil),  // 3: reward.v1.GetRewardRequest
	(*GetRewardResponse)(nil), // 4: reward.v1.GetRewardResponse
}
var file_reward_v1_reward_proto_depIdxs = []int32{
	0, // 0: reward.v1.GetRewardResponse.status:type_name -> reward.v1.RewardStatus
	1, // 1: reward.v1.RewardService.PreReward:input_type -> reward.v1.PreRewardRequest
	3, // 2: reward.v1.RewardService.GetReward:input_type -> reward.v1.GetRewardRequest
	2, // 3: reward.v1.RewardService.PreReward:output_type -> reward.v1.PreRewardResponse
	4, // 4: reward.v1.RewardService.GetReward:output_type -> reward.v1.GetRewardResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_reward_v1_reward_proto_init() }
func file_reward_v1_reward_proto_init() {
	if File_reward_v1_reward_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_reward_v1_reward_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_reward_v1_reward_proto_goTypes,
		DependencyIndexes: file_reward_v1_reward_proto_depIdxs,
		EnumInfos:         file_reward_v1_reward_proto_enumTypes,
		MessageInfos:      file_reward_v1_reward_proto_msgTypes,
	}.Build()
	File_reward_v1_reward_proto = out.File
	file_reward_v1_reward_proto_rawDesc = nil
	file_reward_v1_reward_proto_goTypes = nil
	file_reward_v1_reward_proto_depIdxs = nil
}
