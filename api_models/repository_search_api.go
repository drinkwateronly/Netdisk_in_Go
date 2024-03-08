package api_models

type UserStorageReqAPI struct {
	StorageSize      uint64 `json:"storageSize"`
	TotalStorageSize uint64 `json:"totalStorageSize"`
}
