# Changelog

## UNRELEASED

 * Simple relay server for multiplayer and NAT punch-through.
 * PPU sprite fetching is now takes less time and some other minor rendering
   optimizations (probably not noticeable offline, but should have positive
   effect in multiplayer since it is now has more time to process packets).
 * Changed the way CRC32 for ROMs is calculated to match PRG+CHR method used in
   nescartdb, so that we can later match rom hash with the database (to lookup
   for metadata for example). Unfortunately, this also means the old save states
   are not compatible with this version.
 * Experimented with pixel shaders and added a simple CRT effect (can be disabled
   with -nocrt flag).
 * Netplay does not allocate memory for every message received anymore, instead
   it relies on a pool of pre-allocated buffers.
 * In the offline mode there is now a rewind feature which can be used to go back
   in time in 5 second steps to quickly recover from mistakes. Can be used by
   pressing Ctrl+Z or ⌘+Z.
 * Game Genie codes support via the -gg flag.

## v1.0.0 - 2024-01-26

 * First stable release
