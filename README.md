# WhenIsTheQ
A commandline tool for interacting with transiter to create a mac menubar item.

## Installation

### Set up transiter
You must install transiter https://docs.transiter.dev/ and run the transiter server
locally. (You could can also point whenistheq at a instance of transiter running remotely
but make sure you have permission). Once you have tranister running. You'll also need too
make sure that you install the system you care about `transiter install us-ny-subway`
You should also set up some kind of start up script that ensures that transiter is running in
the background at all times. (I still haven't done this myself though :shrug:)


### Build whenistheq
`go build && mv whenistheq $HOME/.local/bin/`

### Install hammerspoon
Hammmerspoon is only necessary for setting up the menubar application. If you just want to
use whenistheq on the cli, you can skip this bit
Follow the installation instructions at https://www.hammerspoon.org/go/ 

## Usage
### Station IDs
To query data about a specific station you'll need to get it's station id. To do this use
`whenistheq station_lookup <MY STATION>` this will perform a fuzzy search on all of the
stations in configured system. (note a single physical station may have multiple IDs for the
various platforms)

### Getting the time of the next train
 `whenistheq next_train --line Q --station R16 --direction downtown` You can specify the direction
as a station ID or the headsign used by the MTA (only available for the NYC subway) the
headsign is usually one of  uptown, downtown, manhattan, outbound

If you want the time till the next train instead of the absolute time you can use the `--diff` flag

### general flags
* `--addr` sets the address of the transiter server (default: http://localhost:8080)
* `system` sets the name of the transit system you're querying (default: us-ny-subway)

### Icon generation (NYC only)
`whenistheq icon --line Q --output qIcon.png` Creates a 64x64 png icon for the given subway line

## Menubar setup
If you want a menubar that displays the time to the next departure, append the init.lua file
to hammerspoon/inti.lua to your $HOME/.hammerspoon/init.lua replace the station id, line and
direction with your desired parameters and reload your config. Note you will need to manually
generate an icon (or download one from the internet) and place it in the .hammerspoon directory.

## Wishlist
* a real nice config system
* install scripts
* first party menubar icon? (not super motivated to do this tbh)
* Maybe remove the tranister dependency? (You can call MTA GTFS endpoints directly but transiter
  takes care of polling, caching and feed discovery which is nice)

