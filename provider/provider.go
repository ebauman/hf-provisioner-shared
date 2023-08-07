package provider

// Provider is an instance of a hobbyfarm machine provider
type Provider interface {
	// Name is the name of the provider. Should return a short string uniqely identifying
	// the provider. This short string should identify the type of provider,
	// e.g. "aws" or "digitalocean", not a specific _instance_ of a provider
	Name() string
}
