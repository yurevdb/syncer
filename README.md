# Syncer

A basic syncrhonisation tool for remote file systems like Google Drive, Dropbox, etc.
It is made with with a whitelist for files to be synchronised.
This is useful sometimes, but not for everyone.

Made for me, by me to syncrhonise my keepass database.

# Build

There is a build.sh script included to easily build the application and install it.
Or you can do it manually as you like.

 1. Clone or download the source files.
 2. Build the syncer and syncerd files.
 3. Copy the files into the /usr/bin directory.
 4. Optional - Added "nohup syncerd" to the cron

# TODO

- [x] Open authorization url automatically in the default browser
- [x] There seems to be a loopback for the code, either handle it or disable it (hanled it)
- [x] Download a file from filename -> changed to remote id
- [x] Create config file for files to sync (config via sqlite)
- [ ] Create syncing:
    - [x] Pull from google drive on startup
    - [x] Check for changes on a timed interval, and if modified pull it
    - [ ] After changes in the file, push to drive
- [x] Create CLI application
    - [x] auth
    - [x] browse
    - [x] status
    - [x] start daemon
    - [x] stop daemon
    - [x] pull
    - [x] push
    - [x] ls
    - [x] add
    - [x] rm
    - [x] help
- [ ] Parse the arguments and flags correctly (don't rely on the inbuild go ones)
- [ ] Handle different vendor by using flags in the cli tool
