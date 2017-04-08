#docker run -it -p 8080:9090 madmock bash
#docker build -t madmock .
#docker run -e PORT=7070 -p 8080:7070 --rm -it madmock bash
# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.6

ENV TARGET github.com
ENV PORT 9090

# Download and Build mad-mock application inside the container.
RUN go get github.com/skiarn/madmock
RUN go install github.com/skiarn/madmock

# Run the application when the container starts.
# Enter host to mock. Change this to your system ip. ex: -u=127.0.0.1:9090
ENTRYPOINT /go/bin/madmock -u=$TARGET -p=$PORT

# Service listens on port $PORT.
EXPOSE $PORT
