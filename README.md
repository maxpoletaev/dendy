# Dendy

Dendy is a NES/Famicom emulator written in Go and named after the soviet Famicom
bootleg. It’s nothing serious, so do not expect it to beat any of the existing 
emulators in terms of performance or accuracy. Yet, it is capable of running most
of the games I tried, so it’s not completely useless, and it's still a great 
learning experience.

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

Dendy can be played over the network with another player. Run the emulator with
the `-listen=<host>:<port>` flag to start a netplay server that will be waiting
for the second player to connect via the `-connect=<host>:<port>` flag. Once the
connection is established, the game will start for both sides. The player who
started the server will be controlling the first joystick. Ensure that both
sides are using the same ROM file and the same version of the emulator.

The feature works by synchronizing the state of the emulated NES and sending
the controller input over the network to the other player. The algorithm allows
slight drifts in the clock speed and network latency by generating fake input
events for the other player if the state of the emulator is ahead of the other
player’s state. Theoretically, this should keep the game playable for both
players even if the network connection is not very stable (e.g. over the
Internet), but it is not very well tested.

## Tested Games

<details>
<summary>Expand</summary>

| Game | Status | Issues |
|------|--------|--------|
| Balloon Fight | Playable ||
| Battle City | Playable ||
| Batman| Playable ||
| Bomberman | Playable ||
| Castlevania | Playable ||
| Chip 'n Dale | Playable ||
| Chip 'n Dale 2 | Playable | Characters are positioned 1px above the floor |
| Contra | Playable ||
| Contra Force | Playable | Status bar flickering |
| Darkwing Duck | Playable ||
| Donkey Kong | Playable ||
| Double Dragon 2 | Not Playable | Graphical artifacts |
| Duck Hunt | Playable ||
| Duck Tales | Playable ||
| Ice Climber | Playable ||
| Legend of Zelda | Playable ||
| Mario Bros. | Playable ||
| Megaman | Playable ||
| Megaman 4 | Playable ||
| Metal Gear | Playable ||
| Ninja Cat | Playable ||
| Prince of Persia | Playable | Incorrect sprite/background priority |
| Super Mario Bros. | Playable | Status bar flickering |
| Super Mario Bros. 3 | Crash ||
| Super Contra | Playable ||
| Teenage Mutant Ninja Turtles | Not Playable | Black Screen |
| Teenage Mutant Ninja Turtles 3 | Playable | Status bar is not rendered |
| Teenage Mutant Ninja Turtles: Tournament Fighters | Not Playable | Graphical artifacts |

</details>

## Status

<details>
<summary>Expand</summary>

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
* [ ] Cycle-accurate emulation

### Input/Output

* [x] Graphics output
* [x] Controller 1
* [x] Zapper

### Sound

TODO

### Mappers

The goal is to support top 7 mappers covering the majority of games. The
percentage indicates the number of games that use the mapper according to
nescartdb.com.

* [x] MMC1 (Mapper 1) - 28%
* [x] MMC3 (Mapper 4) - 24%
* [x] UxROM (Mapper 2) - 11%
* [x] NROM (Mapper 0) - 10%
* [ ] CNROM (Mapper 3) - 6%
* [ ] AxROM (Mapper 7) - 3%
* [ ] MMC5 (Mapper 5) - 1%

</details>

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
