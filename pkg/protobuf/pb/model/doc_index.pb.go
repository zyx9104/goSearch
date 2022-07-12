// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.1
// source: doc_index.proto

package model

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

type DocIndex struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Text string `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	Url  string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
}

func (x *DocIndex) Reset() {
	*x = DocIndex{}
	if protoimpl.UnsafeEnabled {
		mi := &file_doc_index_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DocIndex) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DocIndex) ProtoMessage() {}

func (x *DocIndex) ProtoReflect() protoreflect.Message {
	mi := &file_doc_index_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DocIndex.ProtoReflect.Descriptor instead.
func (*DocIndex) Descriptor() ([]byte, []int) {
	return file_doc_index_proto_rawDescGZIP(), []int{0}
}

func (x *DocIndex) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *DocIndex) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *DocIndex) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

var File_doc_index_proto protoreflect.FileDescriptor

var file_doc_index_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x64, 0x6f, 0x63, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x22, 0x40, 0x0a, 0x08, 0x44, 0x6f, 0x63, 0x49,
	0x6e, 0x64, 0x65, 0x78, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_doc_index_proto_rawDescOnce sync.Once
	file_doc_index_proto_rawDescData = file_doc_index_proto_rawDesc
)

func file_doc_index_proto_rawDescGZIP() []byte {
	file_doc_index_proto_rawDescOnce.Do(func() {
		file_doc_index_proto_rawDescData = protoimpl.X.CompressGZIP(file_doc_index_proto_rawDescData)
	})
	return file_doc_index_proto_rawDescData
}

var file_doc_index_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_doc_index_proto_goTypes = []interface{}{
	(*DocIndex)(nil), // 0: model.DocIndex
}
var file_doc_index_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_doc_index_proto_init() }
func file_doc_index_proto_init() {
	if File_doc_index_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_doc_index_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DocIndex); i {
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
			RawDescriptor: file_doc_index_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_doc_index_proto_goTypes,
		DependencyIndexes: file_doc_index_proto_depIdxs,
		MessageInfos:      file_doc_index_proto_msgTypes,
	}.Build()
	File_doc_index_proto = out.File
	file_doc_index_proto_rawDesc = nil
	file_doc_index_proto_goTypes = nil
	file_doc_index_proto_depIdxs = nil
}
