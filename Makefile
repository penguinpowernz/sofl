build:
	go build -o usr/bin/sofl ./cmd/sofl

pkg: build
	ian pkg
