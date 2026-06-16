package crypto

// what gets stored in MongoDB per secret
type EncryptedSecret struct {
	EncryptedValue string `json:"encrypted_value"`
	EncryptedDEK   string `json:"encrypted_dek"`
	KEKVersion     int    `json:"kek_version"`
	IV             string `json:"iv"`
	AuthTag        string `json:"auth_tag"`
	Algorithm      string `json:"algorithm"`
}
