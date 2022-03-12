build:
	GOARCH=amd64 GOOS=linux go build -ldflags="-w -s" -o gyrolock

install: build
	sudo cp gyrolock.service /etc/systemd/system
	sudo cp gyrolock /usr/bin/gyrolock

all: install
