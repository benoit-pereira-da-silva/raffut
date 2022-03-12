# raffut
- Allows to transfer the audio input of a local machine to a distant machine by the network in Realtime via UDP.
- It is written in golang and does not require any audio library. 
- It relies on [gen2brain go bindings](github.com/gen2brain/malgo) of [Miniaudio](https://miniaud.io).
- But i also maintain an alternative implementation using [Portaudio](https://github.com/benoit-pereira-da-silva/raffut-portaudio)

# build:
1. `go build -o raffut cmd/main.go`
2. [Code sign and notarize](https://artyom.dev/notarizing-go-binaries-for-macos.html)

# usage:
## Sender:
1. Configure if necessary the audio loop back.
2. Run for example `./raffut send 192.168.1.4:8383`

## Receiver:
1. On the receiver (192.168.1.4) run `./raffut receive 192.168.1.4:8383`

# Configuring the audio loop back
- Configuring an "audio loop back" means to route by a software the default output to the default input.
- It can be also achieved materially by plugging the audio output in an audio input using a [Direct Input unit](https://en.wikipedia.org/wiki/DI_unit).

## Software audio Loop back solutions on macOs:
- Commercial [Loopback](https://rogueamoeba.com/loopback/)
- Un tested [BlackHole?](https://github.com/ExistentialAudio/BlackHole)
- Un tested [Soundflower](https://github.com/mattingalls/Soundflower)

## Software audio Loop back solutions on Linux:
- [Pulseaudio?](https://gitlab.freedesktop.org/pulseaudio/pulseaudio)
- [Jack Audio?](https://jackaudio.org) 
