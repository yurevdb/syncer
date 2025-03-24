# SYNCER

A basic google drive sync tool.
Made for me to sync my keepass database.

Also nice to learn a bit of go.

# TODO

- [x] Open authorication url automatically in the default browser
- [x] There seems to be a loopback for the code, either handle it or disable it (hanled it)
- [x] Download a file from filename
- [x] Create config file for files to sync (config via sqlite)
- [ ] Create syncing:
    - [ ] Pull from google drive on startup
    - [ ] Check for changes on a timed interval, and if modified pull it
    - [ ] After changes in the file, push to drive
- [ ] Create CLI application
    - [x] auth
    - [x] status
    - [ ] start
    - [ ] stop
    - [x] ls
    - [x] add
    - [x] rm
    - [x] help
