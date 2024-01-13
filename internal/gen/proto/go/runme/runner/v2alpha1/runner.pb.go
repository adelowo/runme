// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: runme/runner/v2alpha1/runner.proto

package runnerv2alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ExecuteStop int32

const (
	ExecuteStop_EXECUTE_STOP_UNSPECIFIED ExecuteStop = 0
	ExecuteStop_EXECUTE_STOP_INTERRUPT   ExecuteStop = 1
	ExecuteStop_EXECUTE_STOP_KILL        ExecuteStop = 2
)

// Enum value maps for ExecuteStop.
var (
	ExecuteStop_name = map[int32]string{
		0: "EXECUTE_STOP_UNSPECIFIED",
		1: "EXECUTE_STOP_INTERRUPT",
		2: "EXECUTE_STOP_KILL",
	}
	ExecuteStop_value = map[string]int32{
		"EXECUTE_STOP_UNSPECIFIED": 0,
		"EXECUTE_STOP_INTERRUPT":   1,
		"EXECUTE_STOP_KILL":        2,
	}
)

func (x ExecuteStop) Enum() *ExecuteStop {
	p := new(ExecuteStop)
	*p = x
	return p
}

func (x ExecuteStop) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ExecuteStop) Descriptor() protoreflect.EnumDescriptor {
	return file_runme_runner_v2alpha1_runner_proto_enumTypes[0].Descriptor()
}

func (ExecuteStop) Type() protoreflect.EnumType {
	return &file_runme_runner_v2alpha1_runner_proto_enumTypes[0]
}

func (x ExecuteStop) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ExecuteStop.Descriptor instead.
func (ExecuteStop) EnumDescriptor() ([]byte, []int) {
	return file_runme_runner_v2alpha1_runner_proto_rawDescGZIP(), []int{0}
}

type Project struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// project root folder
	Root string `protobuf:"bytes,1,opt,name=root,proto3" json:"root,omitempty"`
	// list of environment files to try and load
	// start with
	EnvLoadOrder []string `protobuf:"bytes,2,rep,name=env_load_order,json=envLoadOrder,proto3" json:"env_load_order,omitempty"`
}

func (x *Project) Reset() {
	*x = Project{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Project) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Project) ProtoMessage() {}

func (x *Project) ProtoReflect() protoreflect.Message {
	mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Project.ProtoReflect.Descriptor instead.
func (*Project) Descriptor() ([]byte, []int) {
	return file_runme_runner_v2alpha1_runner_proto_rawDescGZIP(), []int{0}
}

func (x *Project) GetRoot() string {
	if x != nil {
		return x.Root
	}
	return ""
}

func (x *Project) GetEnvLoadOrder() []string {
	if x != nil {
		return x.EnvLoadOrder
	}
	return nil
}

type Winsize struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// number of rows (in cells)
	Rows uint32 `protobuf:"varint,1,opt,name=rows,proto3" json:"rows,omitempty"`
	// number of columns (in cells)
	Cols uint32 `protobuf:"varint,2,opt,name=cols,proto3" json:"cols,omitempty"`
	// width in pixels
	X uint32 `protobuf:"varint,3,opt,name=x,proto3" json:"x,omitempty"`
	// height in pixels
	Y uint32 `protobuf:"varint,4,opt,name=y,proto3" json:"y,omitempty"`
}

func (x *Winsize) Reset() {
	*x = Winsize{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Winsize) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Winsize) ProtoMessage() {}

func (x *Winsize) ProtoReflect() protoreflect.Message {
	mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Winsize.ProtoReflect.Descriptor instead.
func (*Winsize) Descriptor() ([]byte, []int) {
	return file_runme_runner_v2alpha1_runner_proto_rawDescGZIP(), []int{1}
}

func (x *Winsize) GetRows() uint32 {
	if x != nil {
		return x.Rows
	}
	return 0
}

func (x *Winsize) GetCols() uint32 {
	if x != nil {
		return x.Cols
	}
	return 0
}

func (x *Winsize) GetX() uint32 {
	if x != nil {
		return x.X
	}
	return 0
}

func (x *Winsize) GetY() uint32 {
	if x != nil {
		return x.Y
	}
	return 0
}

type ExecuteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If it's a relative path and project is not set, directory is used as a base.
	DocumentPath string `protobuf:"bytes,1,opt,name=document_path,json=documentPath,proto3" json:"document_path,omitempty"`
	// project represents a project in which the document is located.
	Project *Project `protobuf:"bytes,2,opt,name=project,proto3,oneof" json:"project,omitempty"`
	// block is either an ID or a name of the block to execute from the document.
	//
	// Types that are assignable to Block:
	//
	//	*ExecuteRequest_BlockId
	//	*ExecuteRequest_BlockName
	Block isExecuteRequest_Block `protobuf_oneof:"block"`
	// directory to execute the program in. If not set,
	// the current working directory is used.
	Directory string `protobuf:"bytes,3,opt,name=directory,proto3" json:"directory,omitempty"`
	// env is a list of additional environment variables
	// that will be injected to the executed program.
	// They will override any env from the project.
	Env []string `protobuf:"bytes,4,rep,name=env,proto3" json:"env,omitempty"`
	// input_data is a byte array that will be send as input
	// to the program.
	InputData []byte `protobuf:"bytes,5,opt,name=input_data,json=inputData,proto3" json:"input_data,omitempty"`
	// stop requests the running process to be stopped.
	// It is allowed only in the consecutive calls.
	Stop ExecuteStop `protobuf:"varint,6,opt,name=stop,proto3,enum=runme.runner.v2alpha1.ExecuteStop" json:"stop,omitempty"`
	// sets pty winsize
	// has no effect in non-interactive mode
	Winsize *Winsize `protobuf:"bytes,7,opt,name=winsize,proto3,oneof" json:"winsize,omitempty"`
	// interactive, if true, will allow to process input_data.
	// When no more data is expected, EOT (0x04) character
	// must be sent in input_data.
	Interactive bool `protobuf:"varint,10,opt,name=interactive,proto3" json:"interactive,omitempty"`
}

func (x *ExecuteRequest) Reset() {
	*x = ExecuteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteRequest) ProtoMessage() {}

func (x *ExecuteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteRequest.ProtoReflect.Descriptor instead.
func (*ExecuteRequest) Descriptor() ([]byte, []int) {
	return file_runme_runner_v2alpha1_runner_proto_rawDescGZIP(), []int{2}
}

func (x *ExecuteRequest) GetDocumentPath() string {
	if x != nil {
		return x.DocumentPath
	}
	return ""
}

func (x *ExecuteRequest) GetProject() *Project {
	if x != nil {
		return x.Project
	}
	return nil
}

func (m *ExecuteRequest) GetBlock() isExecuteRequest_Block {
	if m != nil {
		return m.Block
	}
	return nil
}

func (x *ExecuteRequest) GetBlockId() string {
	if x, ok := x.GetBlock().(*ExecuteRequest_BlockId); ok {
		return x.BlockId
	}
	return ""
}

func (x *ExecuteRequest) GetBlockName() string {
	if x, ok := x.GetBlock().(*ExecuteRequest_BlockName); ok {
		return x.BlockName
	}
	return ""
}

func (x *ExecuteRequest) GetDirectory() string {
	if x != nil {
		return x.Directory
	}
	return ""
}

func (x *ExecuteRequest) GetEnv() []string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *ExecuteRequest) GetInputData() []byte {
	if x != nil {
		return x.InputData
	}
	return nil
}

func (x *ExecuteRequest) GetStop() ExecuteStop {
	if x != nil {
		return x.Stop
	}
	return ExecuteStop_EXECUTE_STOP_UNSPECIFIED
}

func (x *ExecuteRequest) GetWinsize() *Winsize {
	if x != nil {
		return x.Winsize
	}
	return nil
}

func (x *ExecuteRequest) GetInteractive() bool {
	if x != nil {
		return x.Interactive
	}
	return false
}

type isExecuteRequest_Block interface {
	isExecuteRequest_Block()
}

type ExecuteRequest_BlockId struct {
	BlockId string `protobuf:"bytes,8,opt,name=block_id,json=blockId,proto3,oneof"`
}

type ExecuteRequest_BlockName struct {
	BlockName string `protobuf:"bytes,9,opt,name=block_name,json=blockName,proto3,oneof"`
}

func (*ExecuteRequest_BlockId) isExecuteRequest_Block() {}

func (*ExecuteRequest_BlockName) isExecuteRequest_Block() {}

type ProcessPID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pid int64 `protobuf:"varint,1,opt,name=pid,proto3" json:"pid,omitempty"`
}

func (x *ProcessPID) Reset() {
	*x = ProcessPID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProcessPID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProcessPID) ProtoMessage() {}

func (x *ProcessPID) ProtoReflect() protoreflect.Message {
	mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProcessPID.ProtoReflect.Descriptor instead.
func (*ProcessPID) Descriptor() ([]byte, []int) {
	return file_runme_runner_v2alpha1_runner_proto_rawDescGZIP(), []int{3}
}

func (x *ProcessPID) GetPid() int64 {
	if x != nil {
		return x.Pid
	}
	return 0
}

type ExecuteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// exit_code is sent only in the final message.
	ExitCode *wrapperspb.UInt32Value `protobuf:"bytes,1,opt,name=exit_code,json=exitCode,proto3" json:"exit_code,omitempty"`
	// stdout_data contains bytes from stdout since the last response.
	StdoutData []byte `protobuf:"bytes,2,opt,name=stdout_data,json=stdoutData,proto3" json:"stdout_data,omitempty"`
	// stderr_data contains bytes from stderr since the last response.
	StderrData []byte `protobuf:"bytes,3,opt,name=stderr_data,json=stderrData,proto3" json:"stderr_data,omitempty"`
	// pid contains the process' PID
	// this is only sent once in an initial response for background processes.
	Pid *ProcessPID `protobuf:"bytes,4,opt,name=pid,proto3" json:"pid,omitempty"`
}

func (x *ExecuteResponse) Reset() {
	*x = ExecuteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteResponse) ProtoMessage() {}

func (x *ExecuteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_runme_runner_v2alpha1_runner_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteResponse.ProtoReflect.Descriptor instead.
func (*ExecuteResponse) Descriptor() ([]byte, []int) {
	return file_runme_runner_v2alpha1_runner_proto_rawDescGZIP(), []int{4}
}

func (x *ExecuteResponse) GetExitCode() *wrapperspb.UInt32Value {
	if x != nil {
		return x.ExitCode
	}
	return nil
}

func (x *ExecuteResponse) GetStdoutData() []byte {
	if x != nil {
		return x.StdoutData
	}
	return nil
}

func (x *ExecuteResponse) GetStderrData() []byte {
	if x != nil {
		return x.StderrData
	}
	return nil
}

func (x *ExecuteResponse) GetPid() *ProcessPID {
	if x != nil {
		return x.Pid
	}
	return nil
}

var File_runme_runner_v2alpha1_runner_proto protoreflect.FileDescriptor

var file_runme_runner_v2alpha1_runner_proto_rawDesc = []byte{
	0x0a, 0x22, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2f, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2f, 0x76,
	0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6e,
	0x65, 0x72, 0x2e, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61,
	0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x43, 0x0a, 0x07, 0x50,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x65, 0x6e,
	0x76, 0x5f, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x0c, 0x65, 0x6e, 0x76, 0x4c, 0x6f, 0x61, 0x64, 0x4f, 0x72, 0x64, 0x65, 0x72,
	0x22, 0x4d, 0x0a, 0x07, 0x57, 0x69, 0x6e, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x72,
	0x6f, 0x77, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x72, 0x6f, 0x77, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x63,
	0x6f, 0x6c, 0x73, 0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x01,
	0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x01, 0x79, 0x22,
	0xbb, 0x03, 0x0a, 0x0e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x70,
	0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x64, 0x6f, 0x63, 0x75, 0x6d,
	0x65, 0x6e, 0x74, 0x50, 0x61, 0x74, 0x68, 0x12, 0x3d, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65,
	0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x48, 0x01, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x88, 0x01, 0x01, 0x12, 0x1b, 0x0a, 0x08, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f,
	0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x03, 0x65, 0x6e, 0x76, 0x12, 0x1d, 0x0a, 0x0a, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x5f, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x36, 0x0a, 0x04, 0x73, 0x74, 0x6f, 0x70, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x22, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72,
	0x2e, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74,
	0x65, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x04, 0x73, 0x74, 0x6f, 0x70, 0x12, 0x3d, 0x0a, 0x07, 0x77,
	0x69, 0x6e, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x72,
	0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x32, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0x2e, 0x57, 0x69, 0x6e, 0x73, 0x69, 0x7a, 0x65, 0x48, 0x02, 0x52, 0x07,
	0x77, 0x69, 0x6e, 0x73, 0x69, 0x7a, 0x65, 0x88, 0x01, 0x01, 0x12, 0x20, 0x0a, 0x0b, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0b, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x42, 0x07, 0x0a, 0x05,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x77, 0x69, 0x6e, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x1e, 0x0a,
	0x0a, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x50, 0x49, 0x44, 0x12, 0x10, 0x0a, 0x03, 0x70,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x70, 0x69, 0x64, 0x22, 0xc3, 0x01,
	0x0a, 0x0f, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x39, 0x0a, 0x09, 0x65, 0x78, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x55, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x08, 0x65, 0x78, 0x69, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0a, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1f, 0x0a,
	0x0b, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0a, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x44, 0x61, 0x74, 0x61, 0x12, 0x33,
	0x0a, 0x03, 0x70, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x72, 0x75,
	0x6e, 0x6d, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x32, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x50, 0x49, 0x44, 0x52, 0x03,
	0x70, 0x69, 0x64, 0x2a, 0x5e, 0x0a, 0x0b, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x53, 0x74,
	0x6f, 0x70, 0x12, 0x1c, 0x0a, 0x18, 0x45, 0x58, 0x45, 0x43, 0x55, 0x54, 0x45, 0x5f, 0x53, 0x54,
	0x4f, 0x50, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00,
	0x12, 0x1a, 0x0a, 0x16, 0x45, 0x58, 0x45, 0x43, 0x55, 0x54, 0x45, 0x5f, 0x53, 0x54, 0x4f, 0x50,
	0x5f, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x52, 0x55, 0x50, 0x54, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11,
	0x45, 0x58, 0x45, 0x43, 0x55, 0x54, 0x45, 0x5f, 0x53, 0x54, 0x4f, 0x50, 0x5f, 0x4b, 0x49, 0x4c,
	0x4c, 0x10, 0x02, 0x32, 0x6f, 0x0a, 0x0d, 0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x5e, 0x0a, 0x07, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x12,
	0x25, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x76,
	0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x72,
	0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x45,
	0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x28, 0x01, 0x30, 0x01, 0x42, 0x56, 0x5a, 0x54, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x66, 0x75, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x6d,
	0x65, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2f, 0x72, 0x75,
	0x6e, 0x6e, 0x65, 0x72, 0x2f, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x72, 0x75,
	0x6e, 0x6e, 0x65, 0x72, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_runme_runner_v2alpha1_runner_proto_rawDescOnce sync.Once
	file_runme_runner_v2alpha1_runner_proto_rawDescData = file_runme_runner_v2alpha1_runner_proto_rawDesc
)

func file_runme_runner_v2alpha1_runner_proto_rawDescGZIP() []byte {
	file_runme_runner_v2alpha1_runner_proto_rawDescOnce.Do(func() {
		file_runme_runner_v2alpha1_runner_proto_rawDescData = protoimpl.X.CompressGZIP(file_runme_runner_v2alpha1_runner_proto_rawDescData)
	})
	return file_runme_runner_v2alpha1_runner_proto_rawDescData
}

var file_runme_runner_v2alpha1_runner_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_runme_runner_v2alpha1_runner_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_runme_runner_v2alpha1_runner_proto_goTypes = []interface{}{
	(ExecuteStop)(0),               // 0: runme.runner.v2alpha1.ExecuteStop
	(*Project)(nil),                // 1: runme.runner.v2alpha1.Project
	(*Winsize)(nil),                // 2: runme.runner.v2alpha1.Winsize
	(*ExecuteRequest)(nil),         // 3: runme.runner.v2alpha1.ExecuteRequest
	(*ProcessPID)(nil),             // 4: runme.runner.v2alpha1.ProcessPID
	(*ExecuteResponse)(nil),        // 5: runme.runner.v2alpha1.ExecuteResponse
	(*wrapperspb.UInt32Value)(nil), // 6: google.protobuf.UInt32Value
}
var file_runme_runner_v2alpha1_runner_proto_depIdxs = []int32{
	1, // 0: runme.runner.v2alpha1.ExecuteRequest.project:type_name -> runme.runner.v2alpha1.Project
	0, // 1: runme.runner.v2alpha1.ExecuteRequest.stop:type_name -> runme.runner.v2alpha1.ExecuteStop
	2, // 2: runme.runner.v2alpha1.ExecuteRequest.winsize:type_name -> runme.runner.v2alpha1.Winsize
	6, // 3: runme.runner.v2alpha1.ExecuteResponse.exit_code:type_name -> google.protobuf.UInt32Value
	4, // 4: runme.runner.v2alpha1.ExecuteResponse.pid:type_name -> runme.runner.v2alpha1.ProcessPID
	3, // 5: runme.runner.v2alpha1.RunnerService.Execute:input_type -> runme.runner.v2alpha1.ExecuteRequest
	5, // 6: runme.runner.v2alpha1.RunnerService.Execute:output_type -> runme.runner.v2alpha1.ExecuteResponse
	6, // [6:7] is the sub-list for method output_type
	5, // [5:6] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_runme_runner_v2alpha1_runner_proto_init() }
func file_runme_runner_v2alpha1_runner_proto_init() {
	if File_runme_runner_v2alpha1_runner_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_runme_runner_v2alpha1_runner_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Project); i {
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
		file_runme_runner_v2alpha1_runner_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Winsize); i {
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
		file_runme_runner_v2alpha1_runner_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteRequest); i {
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
		file_runme_runner_v2alpha1_runner_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProcessPID); i {
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
		file_runme_runner_v2alpha1_runner_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteResponse); i {
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
	file_runme_runner_v2alpha1_runner_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*ExecuteRequest_BlockId)(nil),
		(*ExecuteRequest_BlockName)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_runme_runner_v2alpha1_runner_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_runme_runner_v2alpha1_runner_proto_goTypes,
		DependencyIndexes: file_runme_runner_v2alpha1_runner_proto_depIdxs,
		EnumInfos:         file_runme_runner_v2alpha1_runner_proto_enumTypes,
		MessageInfos:      file_runme_runner_v2alpha1_runner_proto_msgTypes,
	}.Build()
	File_runme_runner_v2alpha1_runner_proto = out.File
	file_runme_runner_v2alpha1_runner_proto_rawDesc = nil
	file_runme_runner_v2alpha1_runner_proto_goTypes = nil
	file_runme_runner_v2alpha1_runner_proto_depIdxs = nil
}
