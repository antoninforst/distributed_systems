// Protobuf message support for proto.proto.
package grpc

import (
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Request struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NodeId        string                 `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Timestamp     uint64                 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request) Reset() {
	*x = Request{}
	protoimpl.X.MessageStateOf(protoimpl.Pointer(x)).StoreMessageInfo(&file_proto_msgTypes[0])
}
func (x *Request) String() string { return protoimpl.X.MessageStringOf(x) }
func (*Request) ProtoMessage()    {}
func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}
func (x *Request) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}
func (x *Request) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

type Reply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	NodeId        string                 `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Timestamp     uint64                 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Reply) Reset() {
	*x = Reply{}
	protoimpl.X.MessageStateOf(protoimpl.Pointer(x)).StoreMessageInfo(&file_proto_msgTypes[1])
}
func (x *Reply) String() string { return protoimpl.X.MessageStringOf(x) }
func (*Reply) ProtoMessage()    {}
func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}
func (x *Reply) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}
func (x *Reply) GetTimestamp() uint64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

var File_proto_proto protoreflect.FileDescriptor
var file_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_goTypes = []any{(*Request)(nil), (*Reply)(nil)}
var file_proto_depIdxs = []int32{
	0, // 0: mutualexclusion.MutualExclusionService.RequestPermission:input_type -> mutualexclusion.Request
	1, // 1: mutualexclusion.MutualExclusionService.RequestPermission:output_type -> mutualexclusion.Reply
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() {
	str := func(s string) *string { return &s }
	i32 := func(v int32) *int32 { return &v }
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	stringType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	uintType := descriptorpb.FieldDescriptorProto_TYPE_UINT64
	makeMessage := func(name string) *descriptorpb.DescriptorProto {
		return &descriptorpb.DescriptorProto{Name: str(name), Field: []*descriptorpb.FieldDescriptorProto{
			{Name: str("node_id"), Number: i32(1), Label: &label, Type: &stringType, JsonName: str("nodeId")},
			{Name: str("timestamp"), Number: i32(2), Label: &label, Type: &uintType},
		}}
	}
	file := &descriptorpb.FileDescriptorProto{
		Name: str("proto.proto"), Package: str("mutualexclusion"), Syntax: str("proto3"),
		Options:     &descriptorpb.FileOptions{GoPackage: str("task4/grpc;grpc")},
		MessageType: []*descriptorpb.DescriptorProto{makeMessage("Request"), makeMessage("Reply")},
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: str("MutualExclusionService"), Method: []*descriptorpb.MethodDescriptorProto{{
				Name: str("RequestPermission"), InputType: str(".mutualexclusion.Request"), OutputType: str(".mutualexclusion.Reply"),
			}},
		}},
	}
	raw, err := proto.Marshal(file)
	if err != nil {
		panic(err)
	}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(Request{}).PkgPath(),
			RawDescriptor: raw,
			NumMessages:   2,
			NumServices:   1,
		},
		GoTypes:           file_proto_goTypes,
		DependencyIndexes: file_proto_depIdxs,
		MessageInfos:      file_proto_msgTypes,
	}.Build()
	File_proto_proto = out.File
	file_proto_goTypes = nil
	file_proto_depIdxs = nil
}
