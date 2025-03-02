
https://rclone.org/commands/rclone_mount/
rclone mount the google drive folder. 
will need a google API key (theyre free)


A good tutorial: https://rclone.org/drive/

```
rclone mount --cache-dir Cache/ --vfs-cache-mode full --vfs-cache-poll-interval 30s  VexDrive:/2024-2025/Notebook/Obsidian Mount/
```
Those settings work for me, if you wanna play around with it, go for it and lmk if you find better ones
