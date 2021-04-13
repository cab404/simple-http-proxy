package main

const (
	// Two allocations per connection.
	BUFFER_PIPE = 10 * 1024
	// One temporary allocation per connection. Increase if URLs are really long.
	BUFFER_STATUS_LINE = 10 * 1024
)

func Config() ServerContext {
	return ServerContext{
		Routes: []Route{
			{"aleph", "example.com:80"},
			{"beta", "duckduckgo.com:80"},
			{"gamma", "nixos.org:80"},
			{"mow", ":9999"},
		},
		ListenAddress: ":8883",
	}
}
