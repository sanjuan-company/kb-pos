package dto

type Address struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"`
}

type SendRequest struct {
	IdempotencyKey string    `json:"idempotency_key"`
	From           *Address  `json:"from"`
	To             []Address `json:"to" binding:"required,min=1,dive"`
	Cc             []Address `json:"cc" binding:"omitempty,dive"`
	Bcc            []Address `json:"bcc" binding:"omitempty,dive"`
	Subject        string    `json:"subject" binding:"required"`
	PlainText      string    `json:"plain_text"`
	HTML           string    `json:"html"`
	ProviderHint   string    `json:"provider_hint"`
}

type StatusRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}
