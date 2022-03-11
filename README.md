# raffut
- Realtime UDP audio streaming without any dependance in go using [Miniaudio](http://https://miniaud.io)
- There is an alternative implementation using [Portaudio](https://github.com/benoit-pereira-da-silva/raffut-portaudio)

# build:
1. `go build -o raffut cmd/main.go`
2. [Code sign and notarize](https://artyom.dev/notarizing-go-binaries-for-macos.html)

# usage:
## Sender:
1. Configure if necessary the audio loop back to route the default output to default input.
2. Run for example `./raffut send 192.168.1.4:8383`

## Receiver:
1. On the receiver (192.168.1.4) run `./raffut receive 192.168.1.4:8383`

# Configuring the audio loop back
## On macOs:
- Commercial[Loopback](https://rogueamoeba.com/loopback/)
- Un tested [BlackHole?](https://github.com/ExistentialAudio/BlackHole)
- Un tested [Soundflower](https://github.com/mattingalls/Soundflower)

## On Linux:
- [Pulseaudio?](https://gitlab.freedesktop.org/pulseaudio/pulseaudio)
- [Jack Audio?](https://jackaudio.org) 
