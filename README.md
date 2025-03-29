# SYNCER

A basic syncrhonisation tool for remote file systems like Google Drive, Dropbox, etc.
It is made with with a whitelist for files to be synchronised.
This is useful sometimes, but not for everyone.

Made for me, by me to syncrhonise my keepass database.

# Build
 1 Clone or download the source files.
 2 Build the syncer and syncerd files.
 3 Copy the files into the /usr/bin directory.
 4 Optional - Added "nohup syncerd" to the cron

# TODO

- [x] Open authorication url automatically in the default browser
- [x] There seems to be a loopback for the code, either handle it or disable it (hanled it)
- [x] Download a file from filename
- [x] Create config file for files to sync (config via sqlite)
- [ ] Create syncing:
    - [ ] Pull from google drive on startup
    - [ ] Check for changes on a timed interval, and if modified pull it
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
