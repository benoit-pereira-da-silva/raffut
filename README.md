# raffut
Realtime UDP audio streaming in go.

# build:
1. `go build -o raffut cmd/portaudio/main.go` or `go build -o raffut cmd/miniaudio/main.go` or `go build -o raffut cmd/console/main.go`
2. [Code sign and notarize](https://artyom.dev/notarizing-go-binaries-for-macos.html)

# usage:
## Sender:
1. If you use `portaudio` install it. On macOS `brew install portaudio`
2. Configure if necessary the audio loop back to route the default output to default input.
3. Run for example `./raffut send 192.168.1.4:8383`

## Receiver:
1. If you use `portaudio` install it. On macOS `brew install portaudio`
2. On the receiver (192.168.1.4) run `./raffut receive 192.168.1.4:8383`

# Configuring the audio loop back
## On macOs:
- Commercial [Loopback](https://rogueamoeba.com/loopback/)
- Un tested [BlackHole?](https://github.com/ExistentialAudio/BlackHole)
- Un tested [Soundflower](https://github.com/mattingalls/Soundflower)

## On Linux:
- [Pulseaudio?](https://gitlab.freedesktop.org/pulseaudio/pulseaudio)
- [Jack Audio?](https://jackaudio.org) 
