package models

type Chat struct {
	ID        int64  `db:"id"`
	DoctorID  int64  `db:"doctor_id"`
	PatientID int64  `db:"patient_id"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
	DeletedAt *int64 `db:"deleted_at"`
}

type Message struct {
	ID                 int64   `db:"id"`
	ChatID             int64   `db:"chat_id"`
	SenderID           int64   `db:"sender_id"`
	Content            *string `db:"content"`
	AttachmentURL      *string `db:"attachment_url"`
	AttachmentType     *string `db:"attachment_type"`
	AttachmentMimeType *string `db:"attachment_mime_type"`
	CreatedAt          int64   `db:"created_at"`
	DeletedAt          *int64  `db:"deleted_at"`
}

type ChatMessageDetail struct {
	ID                 int64   `db:"id"`
	SenderID           int64   `db:"sender_id"`
	Login              string  `db:"login"`
	FirstName          string  `db:"first_name"`
	LastName           string  `db:"last_name"`
	Content            *string `db:"content"`
	AttachmentURL      *string `db:"attachment_url"`
	AttachmentType     *string `db:"attachment_type"`
	AttachmentMimeType *string `db:"attachment_mime_type"`
	CreatedAt          int64   `db:"created_at"`
}

type UserChatDetail struct {
	ID             int64  `db:"id"`
	DoctorID       int64  `db:"doctor_id"`
	PatientID      int64  `db:"patient_id"`
	OtherUserID    int64  `db:"other_user_id"`
	OtherLogin     string `db:"login"`
	OtherFirstName string `db:"first_name"`
	OtherLastName  string `db:"last_name"`
	UpdatedAt      int64  `db:"updated_at"`
}
