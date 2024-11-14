package dtos

type SignUpRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FingerPrint string `json:"finger_print"`
}
