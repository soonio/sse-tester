

receiver:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o listener ./listen/main.go \
	&& scp listener $(banana):/home/ubuntu \
	&& rm -rf listener

sender:
	go build -ldflags "-s -w" -trimpath -o pusher ./push/main.go \
	&& ./pusher
