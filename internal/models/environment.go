package models

// EnvironmentRequest represents the request to create or update an environment
type EnvironmentRequest struct {
	Username    string            `json:"username" binding:"required"`
	Image       string            `json:"image,omitempty"`
	Ports       []int             `json:"ports,omitempty"` 
	VolumeSize  string            `json:"volumeSize,omitempty"`
	VolumeMounts []VolumeMount    `json:"volumeMounts,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

// VolumeMount represents a volume mount in a container
type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

// Environment represents a user environment
type Environment struct {
	Username     string            `json:"username"`
	Namespace    string            `json:"namespace"`
	Image        string            `json:"image"`
	Ports        []int             `json:"ports"`
	VolumeSize   string            `json:"volumeSize"`
	VolumeMounts []VolumeMount     `json:"volumeMounts,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Status       string            `json:"status"`
	URL          string            `json:"url"`
	CreatedAt    string            `json:"createdAt"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}