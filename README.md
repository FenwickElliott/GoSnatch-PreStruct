# GoSnatch

GoSnach calls the Spotify API and adds whichever song is currently playing to a playlist called 'GoSnatch'.

## Process
First it asks the user for authenication through the Spotify OAuth process and saves the resulting access tokens locally.

Then it uses those keys to consume the Spotify API and discover the currently playing song and adds it to a playlist, creating one if one with the correct name is not found.

## Notifications
Currently it only has working functionality for sending native notifcations on MacOs, I will be happy to impliment native notifications for Windows and linux if ther is any interest.

## Packaged Version
Because running the Go exectuable leaves an unsightly terminal window [here](FenwcikElliott.io/Download/GoSnatch.zip) is a Mac packaged version that does not.