package chatapi

type protoMessage struct{}

func (*protoMessage) Reset()         {}
func (*protoMessage) String() string { return "" }
func (*protoMessage) ProtoMessage()  {}

type ChatDetails struct {
	protoMessage  `json:"-"`
	ChatId        int32    `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Name          string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Topic         string   `protobuf:"bytes,3,opt,name=topic,proto3" json:"topic,omitempty"`
	ActiveUsers   []string `protobuf:"bytes,4,rep,name=active_users,json=activeUsers,proto3" json:"active_users,omitempty"`
	InactiveUsers []string `protobuf:"bytes,5,rep,name=inactive_users,json=inactiveUsers,proto3" json:"inactive_users,omitempty"`
	RemintHost    string   `protobuf:"bytes,6,opt,name=remint_host,json=remintHost,proto3" json:"remint_host,omitempty"`
	RemintPort    int32    `protobuf:"varint,7,opt,name=remint_port,json=remintPort,proto3" json:"remint_port,omitempty"`
}

type MessageDetails struct {
	protoMessage `json:"-"`
	ChatId       int32  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Creator      string `protobuf:"bytes,2,opt,name=creator,proto3" json:"creator,omitempty"`
	Message      string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	IsAction     bool   `protobuf:"varint,4,opt,name=is_action,json=isAction,proto3" json:"is_action,omitempty"`
	Created      int64  `protobuf:"varint,5,opt,name=created,proto3" json:"created,omitempty"`
}

type FetchMessagesRequest struct {
	protoMessage `json:"-"`
	ChatId       int32 `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Count        int32 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

type FetchMessagesResponse struct {
	protoMessage `json:"-"`
	Messages     []*MessageDetails `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

type FetchChatsResponse struct {
	protoMessage `json:"-"`
	Joined       []*ChatDetails `protobuf:"bytes,1,rep,name=joined,proto3" json:"joined,omitempty"`
	NotJoined    []*ChatDetails `protobuf:"bytes,2,rep,name=not_joined,json=notJoined,proto3" json:"not_joined,omitempty"`
}

type PostChatMessageRequest struct {
	protoMessage `json:"-"`
	ChatId       int32  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Message      string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	IsAction     bool   `protobuf:"varint,3,opt,name=is_action,json=isAction,proto3" json:"is_action,omitempty"`
}

type PostChatMessageResponse struct {
	protoMessage `json:"-"`
	Message      *MessageDetails `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type JoinChatRequest struct {
	protoMessage `json:"-"`
	ChatId       int32 `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
}

type JoinChatResponse struct {
	protoMessage `json:"-"`
	ChatId       int32             `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Details      *ChatDetails      `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty"`
	Backlog      []*MessageDetails `protobuf:"bytes,3,rep,name=backlog,proto3" json:"backlog,omitempty"`
}

type LeaveChatRequest struct {
	protoMessage `json:"-"`
	ChatId       int32 `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
}

type LeaveChatResponse struct {
	protoMessage `json:"-"`
	ChatId       int32  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Username     string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
}

type UpdateTopicRequest struct {
	protoMessage `json:"-"`
	ChatId       int32  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Topic        string `protobuf:"bytes,2,opt,name=topic,proto3" json:"topic,omitempty"`
}

type UpdateTopicResponse struct {
	protoMessage `json:"-"`
	Details      *ChatDetails `protobuf:"bytes,1,opt,name=details,proto3" json:"details,omitempty"`
}

type BroadcastChatDetails struct {
	protoMessage `json:"-"`
	ChatId       int32        `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Details      *ChatDetails `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty"`
	User         string       `protobuf:"bytes,3,opt,name=user,proto3" json:"user,omitempty"`
}

type BroadcastChatJoin struct {
	protoMessage `json:"-"`
	ChatId       int32  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Username     string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
}

type BroadcastChatLeave struct {
	protoMessage `json:"-"`
	ChatId       int32  `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Username     string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
}

type BroadcastChatMessage struct {
	protoMessage `json:"-"`
	ChatId       int32           `protobuf:"varint,1,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	Message      *MessageDetails `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}
