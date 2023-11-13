# cli-torrent

CLI-Torrent is a command-line torrent client written in Go. It supports concurrent downloading of multiple files, providing a user interface and a set of commands to efficiently manage downloading files.

## Features:

- **Concurrent Downloading:** CLI-Torrent enables concurrent downloads of multiple files for increased efficiency.
  
- **User Interface:** Despite being a command-line tool, CLI-Torrent offers a user-friendly interface for seamless interaction.
  
- **Command Management:** Take control of your downloads with intuitive commands. Pause, resume, prioritize, or remove torrents effortlessly.

## Getting Started:

1. **Installation:**
    ```bash
    go get github.com/naumovpavel/cli-torrent
    ```

2. **Usage:**
    ```bash
    cli-torrent
    download torrent-file.torrent dest-file
    ```

3. **Commands:**
    - `show`: shows all torrent files that you attempt to download, press escape to exit from show command
    - `exit`: stops all downloads and exits the program
    - `pause index`: pauses downloading a file with this index
    - `continue index`: continues downloading a file with this index
    - `help` : shows all commands with description
    - `download src dest`: downloads torrent file from src to dest

## Demo:



https://github.com/naumovpavel/cli-torrent/assets/40343628/798e57a0-e47f-44f6-8ca2-b2e7ed756918



