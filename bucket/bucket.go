package bucket

type Bucket struct {
	Name     string
	MaxKeys  uint
	Prefix   string
	Marker   string
	Contents []struct {
		Key          string
		LastModified string
		ETag         string
		Size         string
		Owner        struct {
			ID          string
			DisplayName string
		}
		StorageClass string
	}
	CommonPrefixes string
}
