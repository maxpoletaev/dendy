<div align="center">
   <img src="logo.png" width="300" alt="Kivi Logo">
   <p>NES/Famicom emulator with network multiplayer written in Go</p>
   <img src="https://img.shields.io/badge/status-Beta-yellow?style=flat-square" alt=""> 
   <img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="">
</div>

<hr>

Dendy is a NES/Famicom emulator written in Go and named after the soviet Famicom
clone, that I used to have back in my childhood. It’s nothing serious, so do not 
expect it to beat any of the existing emulators in terms of performance or accuracy. 
Yet, it is capable of running most of the games I tried, so it’s not completely useless.

<img src="screenshots.png" alt="Screenshots">

## Try it

```
$ go install github.com/maxpoletaev/dendy/cmd/dendy@latest
$ dendy romfile.nes
```

You may need to install additional dependencies required by raylib. See
https://github.com/gen2brain/raylib-go#requirements for more details.

## Controls

### Controller

Player 1 controller is emulated using the keyboard. The default mapping is as
follows. Multiplayer on a single keyboard is not supported.

```
                   ┆┆
┌───────────────────────────────────────┐
│                                       │
│    [W]                                │
│ [A]   [D]                             │
│    [S]                       [J] [K]  │
│           [Enter] [RShift]            │
│                                       │
└───────────────────────────────────────┘
```

### Zapper (Light Gun)

Zapper is emulated using the mouse and can be used in games like Duck Hunt. Point
the mouse cursor at the right position on the screen and click to shoot.

## Network Multiplayer

Run the emulator with the `-listen=<host>:<port>` flag to start a netplay server
that will be waiting for the second player to connect via the `-connect=<host>:<port>`
flag. Once the connection is established, the game will start for both sides.
The player who started the server will be controlling the first joystick. Ensure
that both sides are using the same ROM file and the same version of the emulator.

The multiplayer feature is based on rollback networking, similar to how it is done in
RetroArch and Fightcade. The emulation always runs at the full speed on both sides,
and the player inputs attached to frame numbers are exchanged over the network. It
compensates for slight drifts in the clock speed and network latency by predicting
input events for the other player when the real input has not yet been received. When
the input from the remote player arrives, it corrects itself by rewinding to the last
known synchronized state and replaying the inputs from that point. Theoretically, this
should keep the game playable for both sides without any local input delay, even if
the network connection is not very stable (e.g., over the Internet). When tested,
latencies of up to 150ms felt pretty comfortable.

## Tested Games

| Game | Status | Issues |
|------|--------|--------|
| [Bad Apple][bad] | Playable ||
| Balloon Fight | Playable ||
| Batman| Playable ||
| Battle City | Playable ||
| Bomberman | Playable ||
| Captain America | Playable ||
| Castlevania | Playable ||
| Chip 'n Dale | Playable ||
| Chip 'n Dale 2 | Playable ||
| Contra Force | Playable ||
| Contra | Playable ||
| Darkwing Duck | Playable ||
| Donkey Kong | Playable ||
| Double Dragon 2 | Playable ||
| Double Dragon 3 | Playable ||
| Duck Hunt | Playable ||
| Duck Tales | Playable ||
| Ice Climber | Playable ||
| Jackal | Not Playable | Black screen |
| Kirby's Adventure | Playable ||
| Legend of Zelda | Playable ||
| Mario Bros. | Playable ||
| Megaman | Playable ||
| Megaman 4 | Playable ||
| Metal Gear | Playable ||
| Mighty Final Fight | Playable ||
| Ninja Cat | Playable ||
| Prince of Persia | Playable | Incorrect sprite/background priority |
| Super Contra | Playable ||
| Super Mario Bros. 3 | Crash ||
| Super Mario Bros. | Playable ||
| Teenage Mutant Ninja Turtles | Playable ||
| Teenage Mutant Ninja Turtles 2 | Playable ||
| Teenage Mutant Ninja Turtles 3 | Playable ||
| Teenage Mutant Ninja Turtles: Tournament Fighters | Not Playable | Graphical artifacts |
| Tiny Toon Adventures | Playable ||

[bad]: https://www.nesdev.org/wiki/Bad_Apple

## Status

### CPU

* [x] Official opcodes
* [x] Unofficial opcodes
* [x] Runtime disassembly
* [x] Cycle-accurate emulation
* [x] Accurate clock speed
* [x] Interrupts

### Graphics

* [x] Background rendering
* [x] Sprite rendering
* [x] 8×16 sprites
* [x] Palettes
* [x] Scrolling
* [ ] Color emphasis

### Input/Output

* [x] Graphics output
* [x] Controllers
* [x] Zapper

### Sound

The sound chip emulation is still work in progress and is not very reliable yet.
It may occasionally produce some pops and crackles, but it should be good enough
for most games.

* [x] Square channels
* [x] Triangle channel
* [x] Noise channel
* [x] Length counter
* [x] Envelope
* [x] Sweep
* [x] DMC

### Mappers

The goal is to support top 7 mappers covering the majority of games. The
percentage indicates the number of games that use the mapper according to
nescartdb.com.

* [x] MMC1 (Mapper 1) - 28%
* [x] MMC3 (Mapper 4) - 24%
* [x] UxROM (Mapper 2) - 11%
* [x] NROM (Mapper 0) - 10%
* [x] CNROM (Mapper 3) - 6%
* [ ] AxROM (Mapper 7) - 3%
* [ ] MMC5 (Mapper 5) - 1%

## Resources

Although NES emulation is a pretty well-covered topic, It is still a very
interesting and challenging project to work on. Here are some of the resources
that I found particularly useful while writing this emulator. Big thanks to
everyone who made them!

### Documentation

* [NESDev Wiki](https://www.nesdev.org/wiki/Nesdev_Wiki)
* [MOS 6502 CPU Reference](https://web.archive.org/web/20210429110213/http://obelisk.me.uk/6502/) by Andrew Jabobs, 2009
* [Extra Instructions of the 65xx Series CPU](http://www.ffd2.com/fridge/docs/6502-NMOS.extra.opcodes) by Adam Vardy, 1996
* [NES Rendering Overview](https://austinmorlan.com/posts/nes_rendering_overview/) by Austin Morlan, 2019
* [Making NES Games in Assembly](https://famicom.party/book/) by Kevin Zurawel, 2021
* [Retroarch Netplay README](https://github.com/libretro/RetroArch/blob/master/network/netplay/README)
* [Audio pseudo-code](https://forums.nesdev.org/viewtopic.php?t=13767)

### Videos

* The [NES Emulator from Scratch][nesemu] series covers most of the topics from
  the CPU to the sound, but I found the two videos about the PPU to be the most
  useful for understanding the obscure details of the NES rendering pipeline:
  [[1]][ppu1], [[2]][ppu2].

[nesemu]: https://www.youtube.com/playlist?list=PLrOv9FMX8xJHqMvSGB_9G9nZZ_4IgteYf
[ppu1]: https://www.youtube.com/watch?v=-THeUXqR3zY&list=PLrOv9FMX8xJHqMvSGB_9G9nZZ_4IgteYf&index=5
[ppu2]: https://www.youtube.com/watch?v=cksywUTZxlY&list=PLrOv9FMX8xJHqMvSGB_9G9nZZ_4IgteYf&index=6

### Code

During bad times, it’s always nice to look at other people’s code to see how
they solved the same problems. Here are some of the emulators written by other
people that I often referred to when I was stuck:

* [github.com/OneLoneCoder/olcNES](https://github.com/OneLoneCoder/olcNES)
* [github.com/ad-sho-loko/goones](https://github.com/ad-sho-loko/goones)
* [github.com/fogleman/nes](https://github.com/fogleman/nes)
* [github.com/marethyu/nesty](https://github.com/marethyu/nesty)
