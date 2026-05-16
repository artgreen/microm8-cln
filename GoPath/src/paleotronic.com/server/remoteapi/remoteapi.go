package remoteapi

type protoMessage struct{}

func (*protoMessage) Reset()         {}
func (*protoMessage) String() string { return "" }
func (*protoMessage) ProtoMessage()  {}

type SlotStatus struct {
	protoMessage     `json:"-"`
	Slotid           int32    `protobuf:"varint,1,opt,name=slotid,proto3" json:"slotid,omitempty"`
	WorkingDirectory string   `protobuf:"bytes,2,opt,name=working_directory,json=workingDirectory,proto3" json:"working_directory,omitempty"`
	Host             string   `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
	Port             int32    `protobuf:"varint,4,opt,name=port,proto3" json:"port,omitempty"`
	Profile          string   `protobuf:"bytes,5,opt,name=profile,proto3" json:"profile,omitempty"`
	State            string   `protobuf:"bytes,6,opt,name=state,proto3" json:"state,omitempty"`
	ActiveUsers      []string `protobuf:"bytes,7,rep,name=active_users,json=activeUsers,proto3" json:"active_users,omitempty"`
	Resources        []string `protobuf:"bytes,8,rep,name=resources,proto3" json:"resources,omitempty"`
}

type RemoteStatusRequest struct {
	protoMessage `json:"-"`
	Status       *SlotStatus `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

type RemoteStatusResponse struct {
	protoMessage `json:"-"`
	Ok           bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

type RemoteListRequest struct {
	protoMessage `json:"-"`
}

type RemoteListResponse struct {
	protoMessage `json:"-"`
	Slots        []*SlotStatus `protobuf:"bytes,1,rep,name=slots,proto3" json:"slots,omitempty"`
}
