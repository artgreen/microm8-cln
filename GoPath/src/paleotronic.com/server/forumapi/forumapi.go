package forumapi

type protoMessage struct{}

func (*protoMessage) Reset()         {}
func (*protoMessage) String() string { return "" }
func (*protoMessage) ProtoMessage()  {}

type ForumDetails struct {
	protoMessage `json:"-"`
	ForumId      int32  `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	Name         string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description  string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

func (m *ForumDetails) GetName() string {
	if m == nil {
		return ""
	}
	return m.Name
}

func (m *ForumDetails) GetDescription() string {
	if m == nil {
		return ""
	}
	return m.Description
}

type MessageDetails struct {
	protoMessage `json:"-"`
	ForumId      int32  `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	MessageId    int32  `protobuf:"varint,2,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	ParentId     int32  `protobuf:"varint,3,opt,name=parent_id,json=parentId,proto3" json:"parent_id,omitempty"`
	Subject      string `protobuf:"bytes,4,opt,name=subject,proto3" json:"subject,omitempty"`
	Body         string `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`
	Creator      string `protobuf:"bytes,6,opt,name=creator,proto3" json:"creator,omitempty"`
	Created      int64  `protobuf:"varint,7,opt,name=created,proto3" json:"created,omitempty"`
}

type FetchMessagesRequest struct {
	protoMessage `json:"-"`
	ForumId      int32 `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	ParentId     int32 `protobuf:"varint,2,opt,name=parent_id,json=parentId,proto3" json:"parent_id,omitempty"`
}

type FetchMessagesResponse struct {
	protoMessage `json:"-"`
	Messages     []*MessageDetails `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

type FetchForumsResponse struct {
	protoMessage `json:"-"`
	Forums       []*ForumDetails `protobuf:"bytes,1,rep,name=forums,proto3" json:"forums,omitempty"`
}

type FetchUnreadMessagesRequest struct {
	protoMessage `json:"-"`
	ForumId      int32 `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
}

type FetchUnreadMessagesResponse struct {
	protoMessage `json:"-"`
	MessageIds   []int32 `protobuf:"varint,1,rep,packed,name=message_ids,json=messageIds,proto3" json:"message_ids,omitempty"`
}

type PostMessageRequest struct {
	protoMessage `json:"-"`
	ForumId      int32  `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	ParentId     int32  `protobuf:"varint,2,opt,name=parent_id,json=parentId,proto3" json:"parent_id,omitempty"`
	Subject      string `protobuf:"bytes,3,opt,name=subject,proto3" json:"subject,omitempty"`
	Body         string `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
}

type PostMessageResponse struct {
	protoMessage `json:"-"`
	Message      *MessageDetails `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type FetchMessageRequest struct {
	protoMessage `json:"-"`
	ForumId      int32 `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	MessageId    int32 `protobuf:"varint,2,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
}

type FetchMessageResponse struct {
	protoMessage `json:"-"`
	Message      *MessageDetails `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

type MarkMessageRequest struct {
	protoMessage `json:"-"`
	ForumId      int32 `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	MessageId    int32 `protobuf:"varint,2,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
}

type MarkMessageResponse struct {
	protoMessage `json:"-"`
	Ok           bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

type SearchForumRequest struct {
	protoMessage `json:"-"`
	ForumId      int32  `protobuf:"varint,1,opt,name=forum_id,json=forumId,proto3" json:"forum_id,omitempty"`
	SearchTerm   string `protobuf:"bytes,2,opt,name=search_term,json=searchTerm,proto3" json:"search_term,omitempty"`
}

type SearchForumResponse struct {
	protoMessage `json:"-"`
	Messages     []*MessageDetails `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

type ForumsWithNewActivityResponse struct {
	protoMessage `json:"-"`
	ForumIds     []int32         `protobuf:"varint,1,rep,packed,name=forum_ids,json=forumIds,proto3" json:"forum_ids,omitempty"`
	Forums       []*ForumDetails `protobuf:"bytes,2,rep,name=forums,proto3" json:"forums,omitempty"`
}
