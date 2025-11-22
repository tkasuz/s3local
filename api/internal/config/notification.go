package config

type NotificationDestination struct {
	Type string `json:"type" yaml:"type"`
	URL  string `json:"url" yaml:"url"`
}

type NotificationRule struct {
	Bucket      string                  `json:"bucket" yaml:"bucket"`
	Events      []string                `json:"events" yaml:"events"`
	Prefix      string                  `json:"prefix" yaml:"prefix"`
	Suffix      string                  `json:"suffix" yaml:"suffix"`
	Destination NotificationDestination `json:"destination" yaml:"destination"`
}
