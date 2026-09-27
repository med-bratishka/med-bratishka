package domain

type Pagination struct {
	Limit  int
	Offset int
	Total  int64
}

type ChatSummary struct {
	ID                int64
	DoctorID          int64
	PatientID         int64
	OtherUserID       int64
	OtherLogin        string
	OtherName         string
	UpdatedAt         int64
	LastMessageID     int64
	LastMessage       string
	LastMessageAt     int64
	UnreadCount       int
	LastReadMessageID int64
	HasUnread         bool
}

type PaginatedChats struct {
	Items      []ChatSummary
	Pagination Pagination
}

type AttachmentInput struct {
	Data      []byte
	FileName  string
	MimeType  string
	MediaType string // image | audio
}

type SendMessageInput struct {
	Content    string
	Attachment *AttachmentInput
}

type ChatMessage struct {
	ID                 int64
	SenderID           int64
	SenderName         string
	Content            string
	AttachmentURL      string
	AttachmentName     string
	AttachmentType     string
	AttachmentMimeType string
	CreatedAt          int64
}

type PaginatedChatMessages struct {
	Items      []ChatMessage
	Pagination Pagination
}

type MarkChatReadInput struct {
	LastReadMessageID int64
}
