// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: intr/v1/intr.proto

package intrv1

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

type IncrReadCntRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
}

func (x *IncrReadCntRequest) Reset() {
	*x = IncrReadCntRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IncrReadCntRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IncrReadCntRequest) ProtoMessage() {}

func (x *IncrReadCntRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IncrReadCntRequest.ProtoReflect.Descriptor instead.
func (*IncrReadCntRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{0}
}

func (x *IncrReadCntRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *IncrReadCntRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

type IncrReadCntResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *IncrReadCntResponse) Reset() {
	*x = IncrReadCntResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IncrReadCntResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IncrReadCntResponse) ProtoMessage() {}

func (x *IncrReadCntResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IncrReadCntResponse.ProtoReflect.Descriptor instead.
func (*IncrReadCntResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{1}
}

type LikeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	Uid   int64  `protobuf:"varint,3,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *LikeRequest) Reset() {
	*x = LikeRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LikeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LikeRequest) ProtoMessage() {}

func (x *LikeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LikeRequest.ProtoReflect.Descriptor instead.
func (*LikeRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{2}
}

func (x *LikeRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *LikeRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *LikeRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type LikeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LikeResponse) Reset() {
	*x = LikeResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LikeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LikeResponse) ProtoMessage() {}

func (x *LikeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LikeResponse.ProtoReflect.Descriptor instead.
func (*LikeResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{3}
}

type CancelLikeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	Uid   int64  `protobuf:"varint,3,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *CancelLikeRequest) Reset() {
	*x = CancelLikeRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelLikeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelLikeRequest) ProtoMessage() {}

func (x *CancelLikeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelLikeRequest.ProtoReflect.Descriptor instead.
func (*CancelLikeRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{4}
}

func (x *CancelLikeRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *CancelLikeRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *CancelLikeRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type CancelLikeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CancelLikeResponse) Reset() {
	*x = CancelLikeResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelLikeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelLikeResponse) ProtoMessage() {}

func (x *CancelLikeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelLikeResponse.ProtoReflect.Descriptor instead.
func (*CancelLikeResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{5}
}

type CollectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	Cid   int64  `protobuf:"varint,3,opt,name=cid,proto3" json:"cid,omitempty"`
	Uid   int64  `protobuf:"varint,4,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *CollectRequest) Reset() {
	*x = CollectRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CollectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CollectRequest) ProtoMessage() {}

func (x *CollectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CollectRequest.ProtoReflect.Descriptor instead.
func (*CollectRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{6}
}

func (x *CollectRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *CollectRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *CollectRequest) GetCid() int64 {
	if x != nil {
		return x.Cid
	}
	return 0
}

func (x *CollectRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type CollectResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CollectResponse) Reset() {
	*x = CollectResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CollectResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CollectResponse) ProtoMessage() {}

func (x *CollectResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CollectResponse.ProtoReflect.Descriptor instead.
func (*CollectResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{7}
}

type CancelCollectRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	Uid   int64  `protobuf:"varint,3,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *CancelCollectRequest) Reset() {
	*x = CancelCollectRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelCollectRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelCollectRequest) ProtoMessage() {}

func (x *CancelCollectRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelCollectRequest.ProtoReflect.Descriptor instead.
func (*CancelCollectRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{8}
}

func (x *CancelCollectRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *CancelCollectRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *CancelCollectRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type CancelCollectResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CancelCollectResponse) Reset() {
	*x = CancelCollectResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CancelCollectResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CancelCollectResponse) ProtoMessage() {}

func (x *CancelCollectResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CancelCollectResponse.ProtoReflect.Descriptor instead.
func (*CancelCollectResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{9}
}

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz   string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	Uid   int64  `protobuf:"varint,3,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *GetRequest) Reset() {
	*x = GetRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRequest) ProtoMessage() {}

func (x *GetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRequest.ProtoReflect.Descriptor instead.
func (*GetRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{10}
}

func (x *GetRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *GetRequest) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *GetRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Intr *Interactive `protobuf:"bytes,1,opt,name=intr,proto3" json:"intr,omitempty"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{11}
}

func (x *GetResponse) GetIntr() *Interactive {
	if x != nil {
		return x.Intr
	}
	return nil
}

type Interactive struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz        string `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizId      int64  `protobuf:"varint,2,opt,name=biz_id,json=bizId,proto3" json:"biz_id,omitempty"`
	ReadCnt    int64  `protobuf:"varint,3,opt,name=read_cnt,json=readCnt,proto3" json:"read_cnt,omitempty"`
	LikeCnt    int64  `protobuf:"varint,4,opt,name=like_cnt,json=likeCnt,proto3" json:"like_cnt,omitempty"`
	CollectCnt int64  `protobuf:"varint,5,opt,name=collect_cnt,json=collectCnt,proto3" json:"collect_cnt,omitempty"`
	Liked      bool   `protobuf:"varint,6,opt,name=liked,proto3" json:"liked,omitempty"`
	Collected  bool   `protobuf:"varint,7,opt,name=collected,proto3" json:"collected,omitempty"`
}

func (x *Interactive) Reset() {
	*x = Interactive{}
	mi := &file_intr_v1_intr_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Interactive) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Interactive) ProtoMessage() {}

func (x *Interactive) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Interactive.ProtoReflect.Descriptor instead.
func (*Interactive) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{12}
}

func (x *Interactive) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *Interactive) GetBizId() int64 {
	if x != nil {
		return x.BizId
	}
	return 0
}

func (x *Interactive) GetReadCnt() int64 {
	if x != nil {
		return x.ReadCnt
	}
	return 0
}

func (x *Interactive) GetLikeCnt() int64 {
	if x != nil {
		return x.LikeCnt
	}
	return 0
}

func (x *Interactive) GetCollectCnt() int64 {
	if x != nil {
		return x.CollectCnt
	}
	return 0
}

func (x *Interactive) GetLiked() bool {
	if x != nil {
		return x.Liked
	}
	return false
}

func (x *Interactive) GetCollected() bool {
	if x != nil {
		return x.Collected
	}
	return false
}

type GetByIdsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Biz    string  `protobuf:"bytes,1,opt,name=biz,proto3" json:"biz,omitempty"`
	BizIds []int64 `protobuf:"varint,2,rep,packed,name=biz_ids,json=bizIds,proto3" json:"biz_ids,omitempty"`
}

func (x *GetByIdsRequest) Reset() {
	*x = GetByIdsRequest{}
	mi := &file_intr_v1_intr_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetByIdsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetByIdsRequest) ProtoMessage() {}

func (x *GetByIdsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetByIdsRequest.ProtoReflect.Descriptor instead.
func (*GetByIdsRequest) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{13}
}

func (x *GetByIdsRequest) GetBiz() string {
	if x != nil {
		return x.Biz
	}
	return ""
}

func (x *GetByIdsRequest) GetBizIds() []int64 {
	if x != nil {
		return x.BizIds
	}
	return nil
}

type GetByIdsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Intrs map[int64]*Interactive `protobuf:"bytes,1,rep,name=intrs,proto3" json:"intrs,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *GetByIdsResponse) Reset() {
	*x = GetByIdsResponse{}
	mi := &file_intr_v1_intr_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetByIdsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetByIdsResponse) ProtoMessage() {}

func (x *GetByIdsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_intr_v1_intr_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetByIdsResponse.ProtoReflect.Descriptor instead.
func (*GetByIdsResponse) Descriptor() ([]byte, []int) {
	return file_intr_v1_intr_proto_rawDescGZIP(), []int{14}
}

func (x *GetByIdsResponse) GetIntrs() map[int64]*Interactive {
	if x != nil {
		return x.Intrs
	}
	return nil
}

var File_intr_v1_intr_proto protoreflect.FileDescriptor

var file_intr_v1_intr_proto_rawDesc = []byte{
	0x0a, 0x12, 0x69, 0x6e, 0x74, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x3d, 0x0a,
	0x12, 0x49, 0x6e, 0x63, 0x72, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x62, 0x69, 0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x22, 0x15, 0x0a, 0x13,
	0x49, 0x6e, 0x63, 0x72, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x48, 0x0a, 0x0b, 0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x62, 0x69, 0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x0e, 0x0a,
	0x0c, 0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x4e, 0x0a,
	0x11, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x62, 0x69, 0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75,
	0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x14, 0x0a,
	0x12, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x5d, 0x0a, 0x0e, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x62, 0x69, 0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x63, 0x69, 0x64,
	0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75,
	0x69, 0x64, 0x22, 0x11, 0x0a, 0x0f, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x51, 0x0a, 0x14, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x43,
	0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a,
	0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x62, 0x69, 0x7a, 0x12,
	0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x17, 0x0a, 0x15, 0x43, 0x61, 0x6e, 0x63,
	0x65, 0x6c, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x47, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x62, 0x69,
	0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x37, 0x0a, 0x0b, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x04, 0x69, 0x6e, 0x74,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x52, 0x04, 0x69,
	0x6e, 0x74, 0x72, 0x22, 0xc1, 0x01, 0x0a, 0x0b, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69, 0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x62, 0x69, 0x7a, 0x12, 0x15, 0x0a, 0x06, 0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x62, 0x69, 0x7a, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08,
	0x72, 0x65, 0x61, 0x64, 0x5f, 0x63, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07,
	0x72, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x6c, 0x69, 0x6b, 0x65, 0x5f,
	0x63, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x6c, 0x69, 0x6b, 0x65, 0x43,
	0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x5f, 0x63, 0x6e,
	0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74,
	0x43, 0x6e, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6b, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x05, 0x6c, 0x69, 0x6b, 0x65, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6c,
	0x6c, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x65, 0x64, 0x22, 0x3c, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x42, 0x79,
	0x49, 0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x62, 0x69,
	0x7a, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x62, 0x69, 0x7a, 0x12, 0x17, 0x0a, 0x07,
	0x62, 0x69, 0x7a, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x03, 0x52, 0x06, 0x62,
	0x69, 0x7a, 0x49, 0x64, 0x73, 0x22, 0x9e, 0x01, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x42, 0x79, 0x49,
	0x64, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3a, 0x0a, 0x05, 0x69, 0x6e,
	0x74, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x69, 0x6e, 0x74, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x79, 0x49, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x49, 0x6e, 0x74, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x05, 0x69, 0x6e, 0x74, 0x72, 0x73, 0x1a, 0x4e, 0x0a, 0x0a, 0x49, 0x6e, 0x74, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x49, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32, 0xdb, 0x03, 0x0a, 0x12, 0x49, 0x6e, 0x74, 0x65, 0x72,
	0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x48, 0x0a,
	0x0b, 0x49, 0x6e, 0x63, 0x72, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x12, 0x1b, 0x2e, 0x69,
	0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x63, 0x72, 0x52, 0x65, 0x61, 0x64, 0x43,
	0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x69, 0x6e, 0x74, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x63, 0x72, 0x52, 0x65, 0x61, 0x64, 0x43, 0x6e, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x4c, 0x69, 0x6b, 0x65, 0x12,
	0x14, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e,
	0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x0a,
	0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4c, 0x69, 0x6b, 0x65, 0x12, 0x1a, 0x2e, 0x69, 0x6e, 0x74,
	0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4c, 0x69, 0x6b, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4c, 0x69, 0x6b, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3c, 0x0a, 0x07, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x12, 0x17,
	0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x4e, 0x0a, 0x0d, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x43, 0x6f, 0x6c, 0x6c, 0x65,
	0x63, 0x74, 0x12, 0x1d, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x6e,
	0x63, 0x65, 0x6c, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x61, 0x6e, 0x63,
	0x65, 0x6c, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x30, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x13, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e,
	0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3f, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x42, 0x79, 0x49, 0x64, 0x73, 0x12,
	0x18, 0x2e, 0x69, 0x6e, 0x74, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x79, 0x49,
	0x64, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x69, 0x6e, 0x74, 0x72,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x79, 0x49, 0x64, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x7a, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x6e, 0x74, 0x72,
	0x2e, 0x76, 0x31, 0x42, 0x09, 0x49, 0x6e, 0x74, 0x72, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x23, 0x77, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x69, 0x6e, 0x74, 0x72, 0x2f, 0x76, 0x31, 0x3b, 0x69,
	0x6e, 0x74, 0x72, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x49, 0x58, 0x58, 0xaa, 0x02, 0x07, 0x49, 0x6e,
	0x74, 0x72, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x07, 0x49, 0x6e, 0x74, 0x72, 0x5c, 0x56, 0x31, 0xe2,
	0x02, 0x13, 0x49, 0x6e, 0x74, 0x72, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x08, 0x49, 0x6e, 0x74, 0x72, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_intr_v1_intr_proto_rawDescOnce sync.Once
	file_intr_v1_intr_proto_rawDescData = file_intr_v1_intr_proto_rawDesc
)

func file_intr_v1_intr_proto_rawDescGZIP() []byte {
	file_intr_v1_intr_proto_rawDescOnce.Do(func() {
		file_intr_v1_intr_proto_rawDescData = protoimpl.X.CompressGZIP(file_intr_v1_intr_proto_rawDescData)
	})
	return file_intr_v1_intr_proto_rawDescData
}

var file_intr_v1_intr_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_intr_v1_intr_proto_goTypes = []any{
	(*IncrReadCntRequest)(nil),    // 0: intr.v1.IncrReadCntRequest
	(*IncrReadCntResponse)(nil),   // 1: intr.v1.IncrReadCntResponse
	(*LikeRequest)(nil),           // 2: intr.v1.LikeRequest
	(*LikeResponse)(nil),          // 3: intr.v1.LikeResponse
	(*CancelLikeRequest)(nil),     // 4: intr.v1.CancelLikeRequest
	(*CancelLikeResponse)(nil),    // 5: intr.v1.CancelLikeResponse
	(*CollectRequest)(nil),        // 6: intr.v1.CollectRequest
	(*CollectResponse)(nil),       // 7: intr.v1.CollectResponse
	(*CancelCollectRequest)(nil),  // 8: intr.v1.CancelCollectRequest
	(*CancelCollectResponse)(nil), // 9: intr.v1.CancelCollectResponse
	(*GetRequest)(nil),            // 10: intr.v1.GetRequest
	(*GetResponse)(nil),           // 11: intr.v1.GetResponse
	(*Interactive)(nil),           // 12: intr.v1.Interactive
	(*GetByIdsRequest)(nil),       // 13: intr.v1.GetByIdsRequest
	(*GetByIdsResponse)(nil),      // 14: intr.v1.GetByIdsResponse
	nil,                           // 15: intr.v1.GetByIdsResponse.IntrsEntry
}
var file_intr_v1_intr_proto_depIdxs = []int32{
	12, // 0: intr.v1.GetResponse.intr:type_name -> intr.v1.Interactive
	15, // 1: intr.v1.GetByIdsResponse.intrs:type_name -> intr.v1.GetByIdsResponse.IntrsEntry
	12, // 2: intr.v1.GetByIdsResponse.IntrsEntry.value:type_name -> intr.v1.Interactive
	0,  // 3: intr.v1.InteractiveService.IncrReadCnt:input_type -> intr.v1.IncrReadCntRequest
	2,  // 4: intr.v1.InteractiveService.Like:input_type -> intr.v1.LikeRequest
	4,  // 5: intr.v1.InteractiveService.CancelLike:input_type -> intr.v1.CancelLikeRequest
	6,  // 6: intr.v1.InteractiveService.Collect:input_type -> intr.v1.CollectRequest
	8,  // 7: intr.v1.InteractiveService.CancelCollect:input_type -> intr.v1.CancelCollectRequest
	10, // 8: intr.v1.InteractiveService.Get:input_type -> intr.v1.GetRequest
	13, // 9: intr.v1.InteractiveService.GetByIds:input_type -> intr.v1.GetByIdsRequest
	1,  // 10: intr.v1.InteractiveService.IncrReadCnt:output_type -> intr.v1.IncrReadCntResponse
	3,  // 11: intr.v1.InteractiveService.Like:output_type -> intr.v1.LikeResponse
	5,  // 12: intr.v1.InteractiveService.CancelLike:output_type -> intr.v1.CancelLikeResponse
	7,  // 13: intr.v1.InteractiveService.Collect:output_type -> intr.v1.CollectResponse
	9,  // 14: intr.v1.InteractiveService.CancelCollect:output_type -> intr.v1.CancelCollectResponse
	11, // 15: intr.v1.InteractiveService.Get:output_type -> intr.v1.GetResponse
	14, // 16: intr.v1.InteractiveService.GetByIds:output_type -> intr.v1.GetByIdsResponse
	10, // [10:17] is the sub-list for method output_type
	3,  // [3:10] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_intr_v1_intr_proto_init() }
func file_intr_v1_intr_proto_init() {
	if File_intr_v1_intr_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_intr_v1_intr_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_intr_v1_intr_proto_goTypes,
		DependencyIndexes: file_intr_v1_intr_proto_depIdxs,
		MessageInfos:      file_intr_v1_intr_proto_msgTypes,
	}.Build()
	File_intr_v1_intr_proto = out.File
	file_intr_v1_intr_proto_rawDesc = nil
	file_intr_v1_intr_proto_goTypes = nil
	file_intr_v1_intr_proto_depIdxs = nil
}