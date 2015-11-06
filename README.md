# git-mtaturnstiles

A small project to help me learn Go.  The first version (MTATstile.go) is a simple command line program.

The MTA releases a csv file every Saturday, so this program will check for the last Saturday's file. If found,
the file will be processed, if not the file will be downloaded to the local directory.